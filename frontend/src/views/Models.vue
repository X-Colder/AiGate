<template>
  <div class="page-container">
    <div class="page-header">
      <h2>模型管理</h2>
      <el-button type="primary" @click="showCreate">新增模型</el-button>
    </div>
    <el-table :data="models" stripe>
      <el-table-column prop="name" label="模型名称" width="150" />
      <el-table-column prop="provider" label="提供者" width="100" />
      <el-table-column prop="model_id" label="模型ID" width="180" />
      <el-table-column prop="billing_mode" label="计费模式" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ billingModeLabel(row.billing_mode) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="输入价格" width="120">
        <template #default="{ row }">¥{{ row.input_price_per_1k }}/1K</template>
      </el-table-column>
      <el-table-column label="输出价格" width="120">
        <template #default="{ row }">¥{{ row.output_price_per_1k }}/1K</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showEdit(row)">编辑</el-button>
          <el-button size="small" type="warning" @click="showDocEditor(row)">文档</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="formVisible" :title="editing ? '编辑模型' : '新增模型'" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="模型名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="提供者">
          <el-select v-model="form.provider" style="width:100%">
            <el-option v-for="p in providers" :key="p" :label="p" :value="p" />
          </el-select>
        </el-form-item>
        <el-form-item label="模型ID"><el-input v-model="form.model_id" placeholder="如 gpt-4, deepseek-chat" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="2" /></el-form-item>
        <el-form-item label="计费模式">
          <el-select v-model="form.billing_mode" style="width:100%">
            <el-option label="预充值余额" value="prepaid" />
            <el-option label="按月配额" value="quota" />
            <el-option label="按次计费" value="per_request" />
            <el-option label="免费额度" value="free_tier" />
          </el-select>
        </el-form-item>
        <el-form-item label="输入单价"><el-input-number v-model="form.input_price_per_1k" :min="0" :precision="4" /> 元/1K tokens</el-form-item>
        <el-form-item label="输出单价"><el-input-number v-model="form.output_price_per_1k" :min="0" :precision="4" /> 元/1K tokens</el-form-item>
        <el-form-item label="按次单价" v-if="form.billing_mode==='per_request'"><el-input-number v-model="form.request_price" :min="0" :precision="4" /> 元/次</el-form-item>
        <el-form-item label="免费额度" v-if="form.billing_mode==='free_tier'"><el-input-number v-model="form.free_quota" :min="0" /> tokens</el-form-item>
        <el-form-item label="月配额" v-if="form.billing_mode==='quota'"><el-input-number v-model="form.monthly_quota" :min="0" /> tokens</el-form-item>
        <el-form-item label="上下文长度"><el-input-number v-model="form.max_context_length" :min="1024" :step="1024" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="docVisible" title="编辑接口文档" width="700px">
      <el-input v-model="docContent" type="textarea" :rows="20" placeholder="支持 Markdown 格式" />
      <template #footer>
        <el-button @click="docVisible = false">取消</el-button>
        <el-button type="primary" @click="handleDocSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { modelApi } from '../api/index.js'

const models = ref([])
const formVisible = ref(false)
const docVisible = ref(false)
const editing = ref(false)
const editId = ref('')
const docEditId = ref('')
const docContent = ref('')
const providers = ['openai', 'anthropic', 'deepseek', 'doubao', 'qwen', 'kimi']

const defaultForm = { name: '', provider: 'openai', model_id: '', description: '', billing_mode: 'prepaid', input_price_per_1k: 0, output_price_per_1k: 0, request_price: 0, free_quota: 0, monthly_quota: 0, max_context_length: 4096, status: 1 }
const form = ref({ ...defaultForm })

const billingModeLabel = (mode) => ({ prepaid: '预充值', quota: '月配额', per_request: '按次', free_tier: '免费额度' }[mode] || mode)

const loadModels = async () => { const res = await modelApi.list(); models.value = res.data || [] }
const showCreate = () => { form.value = { ...defaultForm }; editing.value = false; formVisible.value = true }
const showEdit = (row) => { form.value = { ...row }; editing.value = true; editId.value = row.id; formVisible.value = true }
const showDocEditor = (row) => { docEditId.value = row.id; docContent.value = row.doc_content || ''; docVisible.value = true }

const handleSubmit = async () => {
  if (editing.value) { await modelApi.update(editId.value, form.value) }
  else { await modelApi.create(form.value) }
  formVisible.value = false; ElMessage.success('操作成功'); loadModels()
}
const handleDelete = (row) => {
  ElMessageBox.confirm('确定删除该模型？', '提示', { type: 'warning' }).then(async () => {
    await modelApi.remove(row.id); ElMessage.success('已删除'); loadModels()
  }).catch(() => {})
}
const handleDocSave = async () => {
  await modelApi.updateDoc(docEditId.value, { doc_content: docContent.value })
  docVisible.value = false; ElMessage.success('文档已保存'); loadModels()
}
onMounted(loadModels)
</script>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
</style>
