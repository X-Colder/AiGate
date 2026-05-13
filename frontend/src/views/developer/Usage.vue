<template>
  <div class="page-container">
    <h2>用量统计</h2>
    <el-form :inline="true" style="margin-bottom:20px">
      <el-form-item label="日期范围">
        <el-date-picker v-model="dateRange" type="daterange" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" style="width:260px" />
      </el-form-item>
      <el-form-item><el-button type="primary" @click="loadData">查询</el-button></el-form-item>
    </el-form>
    <el-row :gutter="16" class="stat-cards">
      <el-col :span="6"><el-card><el-statistic title="调用次数" :value="summary.total_requests" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="总Tokens" :value="summary.total_tokens" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="总消费(元)" :value="summary.total_cost" :precision="4" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="平均延迟(ms)" :value="summary.avg_latency_ms" :precision="1" /></el-card></el-col>
    </el-row>
    <el-card style="margin-top:20px">
      <h3>调用记录</h3>
      <el-table :data="records" stripe size="small">
        <el-table-column prop="model_name" label="模型" width="150" />
        <el-table-column prop="input_tokens" label="输入Tokens" width="100" />
        <el-table-column prop="output_tokens" label="输出Tokens" width="100" />
        <el-table-column prop="total_tokens" label="总Tokens" width="100" />
        <el-table-column label="费用" width="100"><template #default="{row}">¥{{ row.cost.toFixed(4) }}</template></el-table-column>
        <el-table-column prop="latency_ms" label="延迟(ms)" width="90" />
        <el-table-column prop="created_at" label="时间" width="180" />
      </el-table>
      <el-pagination v-if="total > 20" :total="total" :page-size="20" v-model:current-page="page" @current-change="loadRecords" style="margin-top:12px" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { developerApi } from '../../api/index.js'

const today = new Date().toISOString().slice(0, 10)
const monthStart = today.slice(0, 8) + '01'
const dateRange = ref([monthStart, today])
const summary = ref({ total_requests: 0, total_tokens: 0, total_cost: 0, avg_latency_ms: 0 })
const records = ref([])
const total = ref(0)
const page = ref(1)

const getParams = () => ({ start_date: dateRange.value[0], end_date: dateRange.value[1] })
const loadData = async () => {
  const res = await developerApi.getUsageSummary(getParams())
  summary.value = res.data || {}
  loadRecords()
}
const loadRecords = async () => {
  const res = await developerApi.getUsageRecords({ ...getParams(), page: page.value, page_size: 20 })
  records.value = res.data?.list || []
  total.value = res.data?.total || 0
}
onMounted(loadData)
</script>

<style scoped>
.page-container { padding: 24px; }
.stat-cards { margin-bottom: 16px; }
</style>
