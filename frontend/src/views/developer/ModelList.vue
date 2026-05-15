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
            <span v-if="m.billing_mode === 'both' || m.billing_mode === 'monthly'" class="price-monthly">月租: ¥{{ m.monthly_price }}/月</span>
          </div>
          <div class="model-pricing" v-if="m.billing_mode === 'both' || m.billing_mode === 'token'">
            <span>Token计费: 输入 ¥{{ m.input_price_per_1k }}/1K 输出 ¥{{ m.output_price_per_1k }}/1K</span>
          </div>
          <div class="model-meta">
            <span>上下文: {{ m.max_context_length }} tokens</span>
            <span>计费: {{ billingLabel(m.billing_mode) }}</span>
          </div>
          <div style="margin-top:12px;display:flex;gap:8px">
            <el-button type="primary" size="small" @click="$router.push(`/developer/models/${m.id}/doc`)">查看文档</el-button>
            <el-button v-if="m.billing_mode === 'both' || m.billing_mode === 'monthly'" type="success" size="small" @click="openPurchase(m)">购买月租</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 购买月租对话框 -->
    <el-dialog v-model="purchaseVisible" title="购买月租订阅" width="440px">
      <el-form v-if="purchaseModel" label-width="90px">
        <el-form-item label="模型">
          <span>{{ purchaseModel.name }}</span>
        </el-form-item>
        <el-form-item label="月租价格">
          <span style="color:#e6a23c;font-weight:bold;font-size:16px">¥{{ purchaseModel.monthly_price }}/月</span>
        </el-form-item>
        <el-form-item label="支付方式">
          <el-radio-group v-model="purchaseForm.paid_by">
            <el-radio value="personal">个人余额</el-radio>
            <el-radio value="tenant" :disabled="!userHasTenant">团队余额</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="purchaseVisible = false">取消</el-button>
        <el-button type="primary" :loading="purchasing" @click="confirmPurchase">确认购买</el-button>
      </template>
    </el-dialog>

    <!-- 购买成功对话框 -->
    <el-dialog v-model="successVisible" title="购买成功" width="400px">
      <el-result icon="success" title="订阅成功">
        <template #sub-title>
          <p>模型: {{ purchaseResult.model_name }}</p>
          <p>订阅到期时间: {{ purchaseResult.end_date }}</p>
        </template>
      </el-result>
      <template #footer>
        <el-button type="primary" @click="successVisible = false">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { developerApi } from '../../api/index.js'

const models = ref([])
const billingLabel = (m) => ({ both: '月租+Token', monthly: '仅月租', token: '仅Token', prepaid: '预充值', quota: '月配额', per_request: '按次', free_tier: '免费额度' }[m] || m)

// 购买相关
const purchaseVisible = ref(false)
const purchaseModel = ref(null)
const purchaseForm = ref({ paid_by: 'personal' })
const purchasing = ref(false)
const successVisible = ref(false)
const purchaseResult = ref({ model_name: '', end_date: '' })

// 判断用户是否有团队
const userHasTenant = computed(() => {
  try {
    const token = localStorage.getItem('token')
    if (!token) return false
    const payload = JSON.parse(atob(token.split('.')[1]))
    return !!payload.tenant_id
  } catch { return false }
})

const openPurchase = (model) => {
  purchaseModel.value = model
  purchaseForm.value = { paid_by: 'personal' }
  purchaseVisible.value = true
}

const confirmPurchase = async () => {
  purchasing.value = true
  try {
    const res = await developerApi.purchaseSubscription({
      model_id: purchaseModel.value.id,
      paid_by: purchaseForm.value.paid_by
    })
    purchaseVisible.value = false
    purchaseResult.value = {
      model_name: purchaseModel.value.name,
      end_date: res.data?.end_date || '一个月后'
    }
    successVisible.value = true
    ElMessage.success('订阅成功')
  } catch (e) {
    // error already handled by interceptor
  } finally {
    purchasing.value = false
  }
}

onMounted(async () => { const res = await developerApi.listModels(); models.value = res.data || [] })
</script>

<style scoped>
.page-container { padding: 24px; }
.model-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.model-name { font-weight: bold; font-size: 15px; }
.model-desc { color: #666; font-size: 13px; margin-bottom: 12px; min-height: 36px; }
.model-pricing { display: flex; gap: 16px; font-size: 13px; color: #409eff; margin-bottom: 8px; }
.price-monthly { color: #e6a23c; font-weight: bold; }
.model-meta { display: flex; gap: 16px; font-size: 12px; color: #999; }
</style>
