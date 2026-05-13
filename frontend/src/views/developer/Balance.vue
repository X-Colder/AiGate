<template>
  <div class="page-container">
    <h2>余额账单</h2>
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="6"><el-card shadow="hover"><el-statistic title="可用余额 (元)" :value="balance.balance" :precision="4" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><el-statistic title="免费额度 (元)" :value="balance.free_balance" :precision="4" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><el-statistic title="累计充值 (元)" :value="balance.total_recharged" :precision="2" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><el-statistic title="累计消费 (元)" :value="balance.total_consumed" :precision="4" /></el-card></el-col>
    </el-row>
    <el-card style="margin-top:20px">
      <h3>交易流水</h3>
      <el-table :data="transactions" stripe size="small">
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.amount > 0 ? 'success' : 'danger'" size="small">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="120">
          <template #default="{ row }"><span :style="{color: row.amount > 0 ? '#67c23a' : '#f56c6c'}">{{ row.amount > 0 ? '+' : '' }}{{ row.amount.toFixed(4) }}</span></template>
        </el-table-column>
        <el-table-column label="余额" width="120"><template #default="{ row }">¥{{ row.balance.toFixed(4) }}</template></el-table-column>
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="created_at" label="时间" width="180" />
      </el-table>
      <el-pagination v-if="total > 20" :total="total" :page-size="20" v-model:current-page="page" @current-change="loadTxns" style="margin-top:12px" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { developerApi } from '../../api/index.js'

const balance = ref({ balance: 0, free_balance: 0, total_recharged: 0, total_consumed: 0 })
const transactions = ref([])
const total = ref(0)
const page = ref(1)

const typeLabel = (t) => ({ recharge: '充值', consume: '消费', refund: '退款', free_grant: '赠送' }[t] || t)
const loadTxns = async () => {
  const res = await developerApi.getTransactions({ page: page.value, page_size: 20 })
  transactions.value = res.data?.list || []
  total.value = res.data?.total || 0
}
onMounted(async () => {
  const res = await developerApi.getBalance()
  balance.value = res.data || {}
  loadTxns()
})
</script>

<style scoped>
.page-container { padding: 24px; }
.stat-cards { margin-top: 16px; }
</style>
