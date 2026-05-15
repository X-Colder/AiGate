<template>
  <div class="page-container">
    <div class="page-header">
      <h2>模型管理</h2>
      <el-button type="primary" @click="showCreate">新增模型</el-button>
    </div>
    <el-table :data="models" stripe>
      <el-table-column prop="name" label="模型名称" width="140" />
      <el-table-column prop="provider" label="提供者" width="100" />
      <el-table-column prop="model_id" label="模型ID" width="160" />
      <el-table-column label="网关" width="140">
        <template #default="{ row }">
          <el-tag v-if="row.gateway_id" size="small">{{ getGatewayName(row.gateway_id) }}</el-tag>
          <span v-else style="color:#999">未绑定</span>
        </template>
      </el-table-column>
      <el-table-column prop="billing_mode" label="计费" width="80">
        <template #default="{ row }"><el-tag size="small" type="info">{{ billingLabel(row.billing_mode) }}</el-tag></template>
      </el-table-column>
      <el-table-column label="输入/输出单价" width="160">
        <template #default="{ row }">¥{{ row.input_price_per_1k }} / ¥{{ row.output_price_per_1k }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-switch :model-value="row.status===1" @change="v=>toggleStatus(row,v)" size="small" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="formVisible" :title="editing ? '编辑模型' : '新增模型'" width="700px" top="5vh">
      <el-tabs>
        <el-tab-pane label="基本信息">
          <el-form :model="form" label-width="100px">
            <el-row :gutter="16">
              <el-col :span="12"><el-form-item label="模型名称"><el-input v-model="form.name" /></el-form-item></el-col>
              <el-col :span="12"><el-form-item label="模型ID"><el-input v-model="form.model_id" placeholder="如 deepseek-chat" /></el-form-item></el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="提供者">
                  <el-select v-model="form.provider" style="width:100%">
                    <el-option v-for="p in providers" :key="p" :label="p" :value="p" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="绑定网关">
                  <el-select v-model="form.gateway_id" style="width:100%" clearable placeholder="选择网关">
                    <el-option v-for="g in gateways" :key="g.id" :label="g.name+' ('+g.provider+')'" :value="g.id" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="2" /></el-form-item>
            <el-row :gutter="16">
              <el-col :span="8">
                <el-form-item label="计费模式">
                  <el-select v-model="form.billing_mode" style="width:100%">
                    <el-option label="月租+Token" value="both" />
                    <el-option label="仅月租" value="monthly" />
                    <el-option label="仅Token" value="token" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="8"><el-form-item label="输入单价"><el-input-number v-model="form.input_price_per_1k" :min="0" :precision="4" style="width:100%" /></el-form-item></el-col>
              <el-col :span="8"><el-form-item label="输出单价"><el-input-number v-model="form.output_price_per_1k" :min="0" :precision="4" style="width:100%" /></el-form-item></el-col>
            </el-row>
            <el-row :gutter="16" v-if="form.billing_mode === 'both' || form.billing_mode === 'monthly'">
              <el-col :span="8">
                <el-form-item label="月租价格">
                  <el-input-number v-model="form.monthly_price" :min="0" :precision="2" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :span="4" style="line-height:32px;padding-top:30px;color:#999">元/月</el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="8"><el-form-item label="按次单价" v-if="form.billing_mode==='per_request'"><el-input-number v-model="form.request_price" :min="0" :precision="4" style="width:100%" /></el-form-item></el-col>
              <el-col :span="8"><el-form-item label="免费额度" v-if="form.billing_mode==='free_tier'"><el-input-number v-model="form.free_quota" :min="0" style="width:100%" /></el-form-item></el-col>
              <el-col :span="8"><el-form-item label="月配额" v-if="form.billing_mode==='quota'"><el-input-number v-model="form.monthly_quota" :min="0" style="width:100%" /></el-form-item></el-col>
            </el-row>
            <el-form-item label="上下文长度"><el-input-number v-model="form.max_context_length" :min="1024" :step="1024" /></el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="接口文档">
          <el-input v-model="form.doc_content" type="textarea" :rows="18" placeholder="支持 Markdown 格式，描述模型调用方式、参数说明、返回格式等" />
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { modelApi, gatewayApi } from '../api/index.js'

const models = ref([])
const gateways = ref([])
const formVisible = ref(false)
const editing = ref(false)
const editId = ref('')
const providers = ['openai', 'anthropic', 'deepseek', 'doubao', 'qwen', 'kimi']

const defaultForm = { name: '', provider: 'openai', model_id: '', description: '', gateway_id: '', billing_mode: 'both', input_price_per_1k: 0, output_price_per_1k: 0, request_price: 0, free_quota: 0, monthly_quota: 0, monthly_price: 0, max_context_length: 4096, doc_content: '' }
const form = ref({ ...defaultForm })

const billingLabel = (m) => ({ both: '月租+Token', monthly: '仅月租', token: '仅Token', prepaid: '预充值', quota: '月配额', per_request: '按次', free_tier: '免费' }[m] || m)
const getGatewayName = (id) => { const g = gateways.value.find(x => x.id === id); return g ? g.name : id.slice(0, 8) }

const loadData = async () => {
  const [mRes, gRes] = await Promise.all([modelApi.list(), gatewayApi.list()])
  models.value = mRes.data || []
  gateways.value = gRes.data || []
}
const showCreate = () => { form.value = { ...defaultForm }; editing.value = false; formVisible.value = true }
const showEdit = (row) => { form.value = { ...row }; editing.value = true; editId.value = row.id; formVisible.value = true }
const toggleStatus = async (row, v) => { await modelApi.update(row.id, { status: v ? 1 : 0 }); loadData() }
const handleSubmit = async () => {
  if (editing.value) { await modelApi.update(editId.value, form.value) }
  else { await modelApi.create(form.value) }
  formVisible.value = false; ElMessage.success('操作成功'); loadData()
}
const handleDelete = (row) => {
  ElMessageBox.confirm('确定删除该模型？', '提示', { type: 'warning' }).then(async () => {
    await modelApi.remove(row.id); ElMessage.success('已删除'); loadData()
  }).catch(() => {})
}
onMounted(loadData)
</script>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
</style>
