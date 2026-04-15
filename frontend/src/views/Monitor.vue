<template>
    <div class="page-container">
        <!-- 筛选栏 -->
        <el-card shadow="hover" style="margin-bottom:20px">
            <el-form :inline="true" :model="queryForm">
                <el-form-item label="网关">
                    <el-select v-model="queryForm.gateway_id" placeholder="全部网关" clearable style="width:200px">
                        <el-option v-for="gw in gateways" :key="gw.id" :label="gw.name" :value="gw.id" />
                    </el-select>
                </el-form-item>
                <el-form-item label="日期范围">
                    <el-date-picker v-model="dateRange" type="daterange" range-separator="至" start-placeholder="开始日期"
                        end-placeholder="结束日期" value-format="YYYY-MM-DD" />
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="fetchData">
                        <el-icon>
                            <Search />
                        </el-icon> 查询
                    </el-button>
                </el-form-item>
            </el-form>
        </el-card>

        <!-- 概览统计卡片 -->
        <el-row :gutter="20" style="margin-bottom:20px">
            <el-col :span="4" v-for="item in summaryCards" :key="item.label">
                <el-card shadow="hover" class="summary-card">
                    <div class="summary-value" :style="{ color: item.color }">{{ item.value }}</div>
                    <div class="summary-label">{{ item.label }}</div>
                </el-card>
            </el-col>
        </el-row>

        <!-- 折线图 -->
        <el-row :gutter="20">
            <el-col :span="12">
                <el-card shadow="hover">
                    <template #header><span style="font-weight:bold;color:#1d3a5f">请求量趋势</span></template>
                    <v-chart :option="requestChartOption" style="height:300px" autoresize />
                </el-card>
            </el-col>
            <el-col :span="12">
                <el-card shadow="hover">
                    <template #header><span style="font-weight:bold;color:#1d3a5f">Token 消耗趋势</span></template>
                    <v-chart :option="tokenChartOption" style="height:300px" autoresize />
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="20" style="margin-top:20px">
            <el-col :span="12">
                <el-card shadow="hover">
                    <template #header><span style="font-weight:bold;color:#1d3a5f">平均延迟趋势</span></template>
                    <v-chart :option="latencyChartOption" style="height:300px" autoresize />
                </el-card>
            </el-col>
            <el-col :span="12">
                <el-card shadow="hover">
                    <template #header><span style="font-weight:bold;color:#1d3a5f">访问用户数趋势</span></template>
                    <v-chart :option="usersChartOption" style="height:300px" autoresize />
                </el-card>
            </el-col>
        </el-row>
    </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { gatewayApi, metricApi } from '../api'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent])

const gateways = ref([])
const queryForm = ref({ gateway_id: '' })
const dateRange = ref([])
const summary = ref({})
const trend = ref([])

// 默认查最近 7 天
const getDefaultRange = () => {
    const end = new Date()
    const start = new Date()
    start.setDate(start.getDate() - 6)
    const fmt = d => d.toISOString().split('T')[0]
    return [fmt(start), fmt(end)]
}

const summaryCards = computed(() => [
    { label: '总请求数', value: summary.value.total_requests || 0, color: '#409eff' },
    { label: '总 Token', value: summary.value.total_tokens || 0, color: '#67c23a' },
    { label: '平均延迟(ms)', value: (summary.value.avg_latency_ms || 0).toFixed(1), color: '#e6a23c' },
    { label: '错误数', value: summary.value.total_errors || 0, color: '#f56c6c' },
    { label: '独立用户', value: summary.value.unique_users || 0, color: '#909399' },
    { label: '错误率', value: ((summary.value.error_rate || 0) * 100).toFixed(1) + '%', color: '#f56c6c' }
])

const buildChartOption = (data, field, color, name) => ({
    tooltip: { trigger: 'axis' },
    grid: { left: 50, right: 20, top: 20, bottom: 30 },
    xAxis: { type: 'category', data: data.map(d => d.date) },
    yAxis: { type: 'value' },
    series: [{ name, type: 'line', data: data.map(d => d[field]), smooth: true, itemStyle: { color }, areaStyle: { color: color + '20' } }]
})

const requestChartOption = computed(() => buildChartOption(trend.value, 'requests', '#409eff', '请求量'))
const tokenChartOption = computed(() => buildChartOption(trend.value, 'tokens', '#67c23a', 'Token'))
const latencyChartOption = computed(() => buildChartOption(trend.value, 'avg_latency', '#e6a23c', '延迟(ms)'))
const usersChartOption = computed(() => buildChartOption(trend.value, 'users', '#909399', '用户数'))

const fetchData = async () => {
    const range = dateRange.value?.length === 2 ? dateRange.value : getDefaultRange()
    const params = { start_date: range[0], end_date: range[1] }
    if (queryForm.value.gateway_id) params.gateway_id = queryForm.value.gateway_id

    try {
        const [summaryRes, trendRes] = await Promise.all([
            metricApi.getSummary(params),
            metricApi.getTrend(params)
        ])
        summary.value = summaryRes.data || {}
        trend.value = trendRes.data || []
    } catch (e) {
        // 错误已在拦截器处理
    }
}

onMounted(async () => {
    dateRange.value = getDefaultRange()
    try {
        const res = await gatewayApi.list()
        gateways.value = res.data || []
    } catch (e) { }
    fetchData()
})
</script>

<style scoped>
.summary-card {
    text-align: center;
    padding: 10px 0;
}

.summary-value {
    font-size: 28px;
    font-weight: bold;
    margin-bottom: 6px;
}

.summary-label {
    font-size: 13px;
    color: #909399;
}
</style>