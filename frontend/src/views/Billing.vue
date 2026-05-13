<template>
  <div class="page-container">
    <div class="page-header">
      <h2>计费管理</h2>
      <el-button type="primary" @click="showRecharge">充值</el-button>
    </div>
    <el-table :data="balances" stripe>
      <el-table-column prop="username" label="用户名" width="150" />
      <el-table-column prop="tenant_name" label="租户" width="150" />
      <el-table-column label="余额" width="120">
        <template #default="{ row }">¥{{ row.balance.toFixed(4) }}</template>
      </el-table-column>
      <el-table-column label="免费额度" width="120">
        <template #default="{ row }">¥{{ row.free_balance.toFixed(4) }}</template>
      </el-table-column>
      <el-table-column label="累计充值" width="120">
        <template #default="{ row }">¥{{ row.total_recharged.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="累计消费" width="120">
        <template #default="{ row }">¥{{ row.total_consumed.toFixed(4) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="showRechargeFor(row)">充值</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="rechargeVisible" title="充值" width="400px">
      <el-form :model="rechargeForm" label-width="80px">
        <el-form-item label="用户ID"><el-input v-model="rechargeForm.user_id" :disabled="!!rechargeForm.username" /></el-form-item>
        <el-form-item label="用户名" v-if="rechargeForm.username"><el-input :model-value="rechargeForm.username" disabled /></el-form-item>
        <el-form-item label="金额"><el-input-number v-model="rechargeForm.amount" :min="0.01" :precision="2" style="width:100%" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="rechargeForm.description" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rechargeVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRecharge">确认充值</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { billingApi } from '../api/index.js'

const balances = ref([])
const rechargeVisible = ref(false)
const rechargeForm = ref({ user_id: '', amount: 10, description: '', username: '' })

const loadBalances = async () => { const res = await billingApi.listUsers(); balances.value = res.data || [] }
const showRecharge = () => { rechargeForm.value = { user_id: '', amount: 10, description: '', username: '' }; rechargeVisible.value = true }
const showRechargeFor = (row) => { rechargeForm.value = { user_id: row.user_id, amount: 10, description: '', username: row.username }; rechargeVisible.value = true }
const handleRecharge = async () => {
  await billingApi.recharge({ user_id: rechargeForm.value.user_id, amount: rechargeForm.value.amount, description: rechargeForm.value.description })
  rechargeVisible.value = false; ElMessage.success('充值成功'); loadBalances()
}
onMounted(loadBalances)
</script>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
</style>
