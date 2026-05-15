<template>
  <div class="page-container">
    <h2>概览</h2>
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="6"><el-card shadow="hover"><el-statistic title="可用余额 (元)" :value="balance.balance + balance.free_balance" :precision="4" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><el-statistic title="本月调用次数" :value="usage.total_requests" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><el-statistic title="本月消耗 Tokens" :value="usage.total_tokens" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><el-statistic title="本月消费 (元)" :value="usage.total_cost" :precision="4" /></el-card></el-col>
    </el-row>
    <el-card style="margin-top:20px">
      <h3>活跃订阅</h3>
      <p v-if="subscriptions.length === 0" style="color:#999">暂无活跃订阅</p>
      <el-table v-else :data="subscriptions" size="small">
        <el-table-column prop="model_name" label="模型名称" />
        <el-table-column prop="end_date" label="到期时间" width="180" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="subscriptionStatusType(row.status)" size="small">{{ subscriptionStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-card style="margin-top:20px">
      <h3>API Keys</h3>
      <p v-if="keys.length === 0" style="color:#999">暂无 API Key，请前往 <router-link to="/developer/apikeys">API Keys</router-link> 页面创建</p>
      <el-table v-else :data="keys" size="small">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="key_prefix" label="Key" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { developerApi } from '../../api/index.js'

const balance = ref({ balance: 0, free_balance: 0 })
const usage = ref({ total_requests: 0, total_tokens: 0, total_cost: 0 })
const keys = ref([])
const subscriptions = ref([])

const subscriptionStatusLabel = (s) => ({ active: '生效中', expired: '已过期', cancelled: '已取消' }[s] || s)
const subscriptionStatusType = (s) => ({ active: 'success', expired: 'info', cancelled: 'danger' }[s] || 'info')

onMounted(async () => {
  const [bRes, kRes] = await Promise.all([developerApi.getBalance(), developerApi.listKeys()])
  balance.value = bRes.data || { balance: 0, free_balance: 0 }
  keys.value = (kRes.data || []).slice(0, 5)
  const today = new Date()
  const startDate = `${today.getFullYear()}-${String(today.getMonth()+1).padStart(2,'0')}-01`
  const endDate = today.toISOString().slice(0, 10)
  try {
    const uRes = await developerApi.getUsageSummary({ start_date: startDate, end_date: endDate })
    usage.value = uRes.data || { total_requests: 0, total_tokens: 0, total_cost: 0 }
  } catch(e) {}
  try {
    const sRes = await developerApi.listSubscriptions()
    subscriptions.value = sRes.data || []
  } catch(e) {}
})
</script>

<style scoped>
.page-container { padding: 24px; }
.stat-cards { margin-top: 16px; }
</style>
