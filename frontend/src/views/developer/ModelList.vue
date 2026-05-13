<template>
  <div class="page-container">
    <h2>可用模型</h2>
    <el-row :gutter="16">
      <el-col :span="8" v-for="m in models" :key="m.id">
        <el-card shadow="hover" class="model-card">
          <template #header>
            <div class="card-header">
              <span class="model-name">{{ m.name }}</span>
              <el-tag size="small">{{ m.provider }}</el-tag>
            </div>
          </template>
          <p class="model-desc">{{ m.description || '暂无描述' }}</p>
          <div class="model-pricing">
            <span>输入: ¥{{ m.input_price_per_1k }}/1K</span>
            <span>输出: ¥{{ m.output_price_per_1k }}/1K</span>
          </div>
          <div class="model-meta">
            <span>上下文: {{ m.max_context_length }} tokens</span>
            <span>计费: {{ billingLabel(m.billing_mode) }}</span>
          </div>
          <el-button type="primary" size="small" style="margin-top:12px" @click="$router.push(`/developer/models/${m.id}/doc`)">查看文档</el-button>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { developerApi } from '../../api/index.js'

const models = ref([])
const billingLabel = (m) => ({ prepaid: '预充值', quota: '月配额', per_request: '按次', free_tier: '免费额度' }[m] || m)
onMounted(async () => { const res = await developerApi.listModels(); models.value = res.data || [] })
</script>

<style scoped>
.page-container { padding: 24px; }
.model-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.model-name { font-weight: bold; font-size: 15px; }
.model-desc { color: #666; font-size: 13px; margin-bottom: 12px; min-height: 36px; }
.model-pricing { display: flex; gap: 16px; font-size: 13px; color: #409eff; margin-bottom: 8px; }
.model-meta { display: flex; gap: 16px; font-size: 12px; color: #999; }
</style>
