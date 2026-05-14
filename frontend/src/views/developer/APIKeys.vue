<template>
  <div class="page-container">
    <div class="page-header">
      <h2>API Keys</h2>
      <el-button type="primary" @click="showCreate">创建 Key</el-button>
    </div>
    <el-alert v-if="newKey" type="success" :closable="true" @close="newKey=''" style="margin-bottom:16px">
      <p><strong>新创建的 Key（仅显示一次，请妥善保存）：</strong></p>
      <code style="word-break:break-all">{{ newKey }}</code>
    </el-alert>
    <el-table :data="keys" stripe>
      <el-table-column prop="name" label="名称" width="140" />
      <el-table-column prop="key_prefix" label="Key" width="130" />
      <el-table-column label="绑定模型" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="m in getModelNames(row.models)" :key="m" size="small" style="margin:2px">{{ m }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="有效期" width="120">
        <template #default="{ row }">{{ row.expires_at ? formatDate(row.expires_at) : '永久' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="statusType(row)" size="small">{{ statusLabel(row) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="last_used_at" label="最后使用" width="140">
        <template #default="{ row }">{{ row.last_used_at ? formatDate(row.last_used_at) : '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status===1" size="small" @click="handleDisable(row)">禁用</el-button>
          <el-button v-else-if="row.status===0" size="small" type="success" @click="handleEnable(row)">启用</el-button>
          <el-button size="small" @click="showExtend(row)">延期</el-button>
          <el-button size="small" @click="showEditModels(row)">模型</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建对话框 -->
    <el-dialog v-model="createVisible" title="创建 API Key" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称"><el-input v-model="form.name" placeholder="如 production-key" /></el-form-item>
        <el-form-item label="绑定模型">
          <el-select v-model="form.modelIds" multiple style="width:100%" placeholder="至少选择一个模型">
            <el-option v-for="m in availableModels" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="有效期">
          <el-select v-model="form.expiryType" style="width:100%">
            <el-option label="长期有效" value="permanent" />
            <el-option label="30 天" value="30d" />
            <el-option label="90 天" value="90d" />
            <el-option label="180 天" value="180d" />
            <el-option label="1 年" value="1y" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.expiryType==='custom'" label="到期日期">
          <el-date-picker v-model="form.customExpiry" type="date" value-format="YYYY-MM-DDT00:00:00Z" style="width:100%" />
        </el-form-item>
        <el-form-item label="QPM限制"><el-input-number v-model="form.rate_limit_qpm" :min="1" :max="1000" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :disabled="form.modelIds.length===0">创建</el-button>
      </template>
    </el-dialog>

    <!-- 延期对话框 -->
    <el-dialog v-model="extendVisible" title="延期 API Key" width="400px">
      <el-form label-width="80px">
        <el-form-item label="新到期日"><el-date-picker v-model="extendDate" type="date" value-format="YYYY-MM-DDT00:00:00Z" style="width:100%" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="extendVisible = false">取消</el-button>
        <el-button type="primary" @click="handleExtend">确认</el-button>
      </template>
    </el-dialog>

    <!-- 修改模型对话框 -->
    <el-dialog v-model="editModelsVisible" title="修改绑定模型" width="400px">
      <el-select v-model="editModelIds" multiple style="width:100%" placeholder="至少选择一个模型">
        <el-option v-for="m in availableModels" :key="m.id" :label="m.name" :value="m.id" />
      </el-select>
      <template #footer>
        <el-button @click="editModelsVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateModels" :disabled="editModelIds.length===0">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { developerApi } from '../../api/index.js'

const keys = ref([])
const availableModels = ref([])
const createVisible = ref(false)
const extendVisible = ref(false)
const editModelsVisible = ref(false)
const newKey = ref('')
const extendKeyId = ref('')
const extendDate = ref('')
const editModelsKeyId = ref('')
const editModelIds = ref([])

const form = ref({ name: '', modelIds: [], expiryType: 'permanent', customExpiry: '', rate_limit_qpm: 60 })

const getModelNames = (models) => {
  if (!models) return []
  return models.split(',').map(id => {
    const m = availableModels.value.find(x => x.id === id)
    return m ? m.name : id.slice(0, 8)
  })
}
const formatDate = (d) => d ? new Date(d).toLocaleDateString() : ''
const statusType = (row) => { if (row.status === 0) return 'danger'; if (row.expires_at && new Date(row.expires_at) < new Date()) return 'warning'; return 'success' }
const statusLabel = (row) => { if (row.status === 0) return '禁用'; if (row.expires_at && new Date(row.expires_at) < new Date()) return '已过期'; return '启用' }

const computeExpiry = () => {
  const t = form.value.expiryType
  if (t === 'permanent') return null
  if (t === 'custom') return form.value.customExpiry || null
  const days = { '30d': 30, '90d': 90, '180d': 180, '1y': 365 }[t]
  const d = new Date(); d.setDate(d.getDate() + days)
  return d.toISOString()
}

const loadData = async () => {
  const [kRes, mRes] = await Promise.all([developerApi.listKeys(), developerApi.listModels()])
  keys.value = kRes.data || []
  availableModels.value = mRes.data || []
}
const showCreate = () => { form.value = { name: '', modelIds: [], expiryType: 'permanent', customExpiry: '', rate_limit_qpm: 60 }; createVisible.value = true }
const handleCreate = async () => {
  const payload = { name: form.value.name, models: form.value.modelIds.join(','), rate_limit_qpm: form.value.rate_limit_qpm, expires_at: computeExpiry() }
  const res = await developerApi.createKey(payload)
  newKey.value = res.data.key
  createVisible.value = false; ElMessage.success('创建成功'); loadData()
}
const handleEnable = async (row) => { await developerApi.updateKey(row.id, { status: 1 }); ElMessage.success('已启用'); loadData() }
const handleDisable = async (row) => { await developerApi.updateKey(row.id, { status: 0 }); ElMessage.success('已禁用'); loadData() }
const showExtend = (row) => { extendKeyId.value = row.id; extendDate.value = ''; extendVisible.value = true }
const handleExtend = async () => { await developerApi.updateKey(extendKeyId.value, { expires_at: extendDate.value }); extendVisible.value = false; ElMessage.success('已延期'); loadData() }
const showEditModels = (row) => { editModelsKeyId.value = row.id; editModelIds.value = row.models ? row.models.split(',') : []; editModelsVisible.value = true }
const handleUpdateModels = async () => { await developerApi.updateKey(editModelsKeyId.value, { models: editModelIds.value.join(',') }); editModelsVisible.value = false; ElMessage.success('已更新'); loadData() }
const handleDelete = (row) => {
  ElMessageBox.confirm('删除后不可恢复，确定删除？', '提示', { type: 'warning' }).then(async () => {
    await developerApi.deleteKey(row.id); ElMessage.success('已删除'); loadData()
  }).catch(() => {})
}
onMounted(loadData)
</script>

<style scoped>
.page-container { padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
</style>
