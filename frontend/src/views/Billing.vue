<template>
  <div class="page-container">
    <h2>计费管理</h2>

    <el-tabs v-model="activeTab">
      <!-- 模型成本/收入 -->
      <el-tab-pane label="模型财务" name="models">
        <el-row :gutter="16">
          <el-col :span="8" v-for="m in modelFinance" :key="m.model_id">
            <el-card shadow="hover" class="model-finance-card" :class="{ 'alert-card': m.upstream_balance <= m.alert_threshold && m.upstream_balance >= 0 }">
              <template #header>
                <div class="card-header">
                  <span class="model-name">{{ m.model_name }}</span>
                  <el-tag size="small">{{ m.provider }}</el-tag>
                </div>
              </template>
              <el-descriptions :column="1" size="small" border>
                <el-descriptions-item label="使用用户">{{ m.user_count }} 人</el-descriptions-item>
                <el-descriptions-item label="收入">
                  <span style="color:#67c23a;font-weight:bold">¥{{ m.total_revenue.toFixed(4) }}</span>
                </el-descriptions-item>
                <el-descriptions-item label="上游成本">
                  <span style="color:#f56c6c">¥{{ m.upstream_total_cost.toFixed(4) }}</span>
                </el-descriptions-item>
                <el-descriptions-item label="利润">
                  <span :style="{ color: m.profit >= 0 ? '#67c23a' : '#f56c6c', fontWeight: 'bold' }">¥{{ m.profit.toFixed(4) }}</span>
                </el-descriptions-item>
                <el-descriptions-item label="上游余额">
                  <span :style="{ color: m.upstream_balance <= m.alert_threshold ? '#f56c6c' : '#303133', fontWeight: 'bold' }">
                    ¥{{ m.upstream_balance.toFixed(2) }}
                  </span>
                  <el-tag v-if="m.upstream_balance <= m.alert_threshold" type="danger" size="small" style="margin-left:6px">余额不足</el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="累计充值">¥{{ m.upstream_total_recharge.toFixed(2) }}</el-descriptions-item>
              </el-descriptions>
              <div style="margin-top:12px;text-align:right">
                <el-button type="primary" size="small" @click="showModelRecharge(m)">充值</el-button>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-empty v-if="modelFinance.length===0" description="暂无模型数据" />
      </el-tab-pane>

      <!-- 用户余额 -->
      <el-tab-pane label="用户余额" name="users">
        <div style="margin-bottom:12px;text-align:right">
          <el-button type="primary" @click="showUserRecharge">用户充值</el-button>
        </div>
        <el-table :data="userBalances" stripe>
          <el-table-column prop="username" label="用户名" width="150" />
          <el-table-column prop="tenant_name" label="团队" width="150" />
          <el-table-column label="余额" width="120"><template #default="{row}">¥{{ row.balance.toFixed(4) }}</template></el-table-column>
          <el-table-column label="免费额度" width="120"><template #default="{row}">¥{{ row.free_balance.toFixed(4) }}</template></el-table-column>
          <el-table-column label="累计充值" width="120"><template #default="{row}">¥{{ row.total_recharged.toFixed(2) }}</template></el-table-column>
          <el-table-column label="累计消费" width="120"><template #default="{row}">¥{{ row.total_consumed.toFixed(4) }}</template></el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{row}"><el-button size="small" type="primary" @click="showUserRechargeFor(row)">充值</el-button></template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 模型充值对话框 -->
    <el-dialog v-model="modelRechargeVisible" title="模型上游充值" width="400px">
      <p style="margin-bottom:12px;color:#606266">模型: <strong>{{ modelRechargeForm.model_name }}</strong></p>
      <el-form :model="modelRechargeForm" label-width="80px">
        <el-form-item label="充值金额"><el-input-number v-model="modelRechargeForm.amount" :min="0.01" :precision="2" style="width:100%" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="modelRechargeForm.description" placeholder="如: 充值OpenAI账户$100" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modelRechargeVisible = false">取消</el-button>
        <el-button type="primary" @click="handleModelRecharge">确认充值</el-button>
      </template>
    </el-dialog>

    <!-- 用户充值对话框 -->
    <el-dialog v-model="userRechargeVisible" title="用户充值" width="400px">
      <el-form :model="userRechargeForm" label-width="80px">
        <el-form-item label="用户ID"><el-input v-model="userRechargeForm.user_id" :disabled="!!userRechargeForm.username" /></el-form-item>
        <el-form-item label="用户名" v-if="userRechargeForm.username"><el-input :model-value="userRechargeForm.username" disabled /></el-form-item>
        <el-form-item label="金额"><el-input-number v-model="userRechargeForm.amount" :min="0.01" :precision="2" style="width:100%" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="userRechargeForm.description" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="userRechargeVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUserRecharge">确认充值</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { billingApi } from '../api/index.js'

const activeTab = ref('models')
const modelFinance = ref([])
const userBalances = ref([])

const modelRechargeVisible = ref(false)
const modelRechargeForm = ref({ model_id: '', model_name: '', amount: 100, description: '' })

const userRechargeVisible = ref(false)
const userRechargeForm = ref({ user_id: '', amount: 10, description: '', username: '' })

const loadModels = async () => { const res = await billingApi.getModelFinance(); modelFinance.value = res.data || [] }
const loadUsers = async () => { const res = await billingApi.listUsers(); userBalances.value = res.data || [] }

const showModelRecharge = (m) => {
  modelRechargeForm.value = { model_id: m.model_id, model_name: m.model_name, amount: 100, description: '' }
  modelRechargeVisible.value = true
}
const handleModelRecharge = async () => {
  await billingApi.rechargeModel({ model_id: modelRechargeForm.value.model_id, amount: modelRechargeForm.value.amount, description: modelRechargeForm.value.description })
  modelRechargeVisible.value = false; ElMessage.success('充值成功'); loadModels()
}

const showUserRecharge = () => { userRechargeForm.value = { user_id: '', amount: 10, description: '', username: '' }; userRechargeVisible.value = true }
const showUserRechargeFor = (row) => { userRechargeForm.value = { user_id: row.user_id, amount: 10, description: '', username: row.username }; userRechargeVisible.value = true }
const handleUserRecharge = async () => {
  await billingApi.recharge({ user_id: userRechargeForm.value.user_id, amount: userRechargeForm.value.amount, description: userRechargeForm.value.description })
  userRechargeVisible.value = false; ElMessage.success('充值成功'); loadUsers()
}

onMounted(() => { loadModels(); loadUsers() })
</script>

<style scoped>
.page-container { padding: 20px; }
.model-finance-card { margin-bottom: 16px; }
.model-finance-card .card-header { display: flex; justify-content: space-between; align-items: center; }
.model-finance-card .model-name { font-weight: bold; font-size: 15px; }
.alert-card { border-color: #f56c6c; }
</style>
