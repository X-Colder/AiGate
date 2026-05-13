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
      <el-table-column prop="name" label="名称" width="180" />
      <el-table-column prop="key_prefix" label="Key" width="160" />
      <el-table-column prop="rate_limit_qpm" label="QPM限制" width="100" />
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column prop="last_used_at" label="最后使用" width="180" />
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button size="small" type="danger" @click="handleRevoke(row)">吊销</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="createVisible" title="创建 API Key" width="400px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称"><el-input v-model="form.name" placeholder="如 production-key" /></el-form-item>
        <el-form-item label="QPM限制"><el-input-number v-model="form.rate_limit_qpm" :min="1" :max="1000" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { developerApi } from '../../api/index.js'

const keys = ref([])
const createVisible = ref(false)
const newKey = ref('')
const form = ref({ name: '', rate_limit_qpm: 60 })

const loadKeys = async () => { const res = await developerApi.listKeys(); keys.value = res.data || [] }
const showCreate = () => { form.value = { name: '', rate_limit_qpm: 60 }; createVisible.value = true }
const handleCreate = async () => {
  const res = await developerApi.createKey(form.value)
  newKey.value = res.data.key
  createVisible.value = false; ElMessage.success('创建成功'); loadKeys()
}
const handleRevoke = (row) => {
  ElMessageBox.confirm('吊销后该 Key 将立即失效，确定吊销？', '提示', { type: 'warning' }).then(async () => {
    await developerApi.revokeKey(row.id); ElMessage.success('已吊销'); loadKeys()
  }).catch(() => {})
}
onMounted(loadKeys)
</script>

<style scoped>
.page-container { padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
</style>
