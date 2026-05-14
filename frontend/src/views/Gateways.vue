<template>
    <div class="page-container">
        <el-card shadow="hover">
            <template #header>
                <div class="card-header">
                    <span style="font-size:18px;font-weight:bold;color:#1d3a5f">网关管理</span>
                    <el-button type="primary" @click="showCreateDialog">
                        <el-icon><Plus /></el-icon> 新增网关
                    </el-button>
                </div>
            </template>
            <el-table :data="gateways" stripe style="width:100%" v-loading="loading">
                <el-table-column prop="name" label="网关名称" min-width="120" />
                <el-table-column prop="provider" label="AI 服务" width="120">
                    <template #default="{ row }"><el-tag type="primary">{{ row.provider }}</el-tag></template>
                </el-table-column>
                <el-table-column label="绑定模型" min-width="180">
                    <template #default="{ row }">
                        <el-tag v-for="m in getBoundModels(row.id)" :key="m.id" size="small" style="margin:2px">{{ m.name }}</el-tag>
                        <span v-if="getBoundModels(row.id).length===0" style="color:#999">无</span>
                    </template>
                </el-table-column>
                <el-table-column prop="base_url" label="API 地址" min-width="200" show-overflow-tooltip />
                <el-table-column label="状态" width="80" align="center">
                    <template #default="{ row }">
                        <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column label="操作" width="260" align="center">
                    <template #default="{ row }">
                        <el-button size="small" @click="showEditDialog(row)">编辑</el-button>
                        <el-button size="small" type="warning" @click="showPolicyDialog(row)">策略</el-button>
                        <el-popconfirm title="确定删除该网关?" @confirm="handleDelete(row.id)">
                            <template #reference>
                                <el-button size="small" type="danger">删除</el-button>
                            </template>
                        </el-popconfirm>
                    </template>
                </el-table-column>
            </el-table>
        </el-card>

        <!-- 新增/编辑网关对话框 -->
        <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑网关' : '新增网关'" width="560px" destroy-on-close>
            <el-form ref="gwFormRef" :model="gwForm" :rules="gwRules" label-width="90px">
                <el-form-item label="名称" prop="name">
                    <el-input v-model="gwForm.name" placeholder="网关名称" />
                </el-form-item>
                <el-form-item label="AI 服务" prop="provider">
                    <el-select v-model="gwForm.provider" placeholder="选择 AI 服务" style="width:100%">
                        <el-option v-for="p in providers" :key="p" :label="p" :value="p" />
                    </el-select>
                </el-form-item>
                <el-form-item label="API 地址">
                    <el-input v-model="gwForm.base_url" placeholder="https://api.example.com/v1" />
                </el-form-item>
                <el-form-item label="上游密钥">
                    <el-input v-model="gwForm.api_key" placeholder="选填，自建服务可留空" show-password />
                    <div style="font-size:12px;color:#909399;margin-top:4px">第三方服务需填写提供商的 API Key，自建服务（vLLM/Ollama 等）可留空</div>
                </el-form-item>
                <el-form-item label="超时(秒)">
                    <el-input-number v-model="gwForm.timeout" :min="5" :max="300" />
                </el-form-item>
                <el-form-item label="绑定模型">
                    <el-select v-model="gwForm.modelIds" multiple style="width:100%" placeholder="选择要绑定的模型">
                        <el-option v-for="m in allModels" :key="m.id" :label="m.name + ' (' + m.model_id + ')'" :value="m.id" />
                    </el-select>
                    <div v-if="allModels.length===0" style="margin-top:6px">
                        <el-alert type="warning" :closable="false" show-icon>
                            <template #title>暂无模型，请先到 <router-link to="/models" style="color:#409eff">模型管理</router-link> 创建模型</template>
                        </el-alert>
                    </div>
                </el-form-item>
                <el-form-item label="状态" v-if="isEdit">
                    <el-switch v-model="gwForm.statusBool" active-text="启用" inactive-text="禁用" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="dialogVisible = false">取消</el-button>
                <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
            </template>
        </el-dialog>

        <!-- 策略配置对话框 -->
        <el-dialog v-model="policyVisible" title="网关策略配置" width="600px" destroy-on-close>
            <el-form :model="policyForm" label-width="140px">
                <el-divider content-position="left">限流策略</el-divider>
                <el-form-item label="启用限流">
                    <el-switch v-model="policyForm.rate_limit_enabled" />
                </el-form-item>
                <el-form-item label="QPS 上限" v-if="policyForm.rate_limit_enabled">
                    <el-input-number v-model="policyForm.rate_limit_qps" :min="1" :max="10000" />
                </el-form-item>
                <el-form-item label="突发上限" v-if="policyForm.rate_limit_enabled">
                    <el-input-number v-model="policyForm.rate_limit_burst" :min="1" :max="50000" />
                </el-form-item>

                <el-divider content-position="left">熔断策略</el-divider>
                <el-form-item label="启用熔断">
                    <el-switch v-model="policyForm.circuit_breaker_enabled" />
                </el-form-item>
                <el-form-item label="错误率阈值" v-if="policyForm.circuit_breaker_enabled">
                    <el-slider v-model="policyForm.circuit_breaker_threshold" :min="0" :max="1" :step="0.05" show-input />
                </el-form-item>
                <el-form-item label="恢复时间(秒)" v-if="policyForm.circuit_breaker_enabled">
                    <el-input-number v-model="policyForm.circuit_breaker_timeout" :min="5" :max="300" />
                </el-form-item>
                <el-form-item label="最小请求数" v-if="policyForm.circuit_breaker_enabled">
                    <el-input-number v-model="policyForm.circuit_breaker_min_reqs" :min="1" :max="1000" />
                </el-form-item>

                <el-divider content-position="left">降级策略</el-divider>
                <el-form-item label="启用降级">
                    <el-switch v-model="policyForm.fallback_enabled" />
                </el-form-item>
                <el-form-item label="降级网关" v-if="policyForm.fallback_enabled">
                    <el-select v-model="policyForm.fallback_gateway_id" style="width:100%" placeholder="选择降级网关">
                        <el-option v-for="g in otherGateways" :key="g.id" :label="g.name + ' (' + g.provider + ')'" :value="g.id" />
                    </el-select>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="policyVisible = false">取消</el-button>
                <el-button type="primary" :loading="savingPolicy" @click="handleSavePolicy">保存策略</el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { gatewayApi, modelApi } from '../api'

const providers = ['openai', 'anthropic', 'deepseek', 'doubao', 'qwen', 'kimi']
const gateways = ref([])
const allModels = ref([])
const loading = ref(false)

const getBoundModels = (gatewayId) => allModels.value.filter(m => m.gateway_id === gatewayId)

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref('')
const saving = ref(false)
const gwFormRef = ref()
const gwForm = ref({ name: '', provider: '', base_url: '', api_key: '', timeout: 60, statusBool: true, modelIds: [] })
const gwRules = {
    name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
    provider: [{ required: true, message: '请选择服务', trigger: 'change' }]
}

const policyVisible = ref(false)
const policyGatewayId = ref('')
const savingPolicy = ref(false)
const policyForm = ref({
    rate_limit_enabled: false, rate_limit_qps: 100, rate_limit_burst: 200,
    circuit_breaker_enabled: false, circuit_breaker_threshold: 0.5, circuit_breaker_timeout: 30, circuit_breaker_min_reqs: 10,
    fallback_enabled: false, fallback_gateway_id: ''
})

const otherGateways = computed(() => gateways.value.filter(g => g.id !== policyGatewayId.value))

const fetchList = async () => {
    loading.value = true
    try {
        const [gRes, mRes] = await Promise.all([gatewayApi.list(), modelApi.list()])
        gateways.value = gRes.data || []
        allModels.value = mRes.data || []
    } finally {
        loading.value = false
    }
}

const showCreateDialog = () => {
    isEdit.value = false
    editId.value = ''
    gwForm.value = { name: '', provider: '', base_url: '', api_key: '', timeout: 60, statusBool: true, modelIds: [] }
    dialogVisible.value = true
}

const showEditDialog = (row) => {
    isEdit.value = true
    editId.value = row.id
    const boundIds = allModels.value.filter(m => m.gateway_id === row.id).map(m => m.id)
    gwForm.value = { name: row.name, provider: row.provider, base_url: row.base_url, api_key: '', timeout: row.timeout, statusBool: row.status === 1, modelIds: boundIds }
    dialogVisible.value = true
}

const handleSave = async () => {
    const valid = await gwFormRef.value.validate().catch(() => false)
    if (!valid) return

    saving.value = true
    try {
        const data = { name: gwForm.value.name, provider: gwForm.value.provider, base_url: gwForm.value.base_url, api_key: gwForm.value.api_key, timeout: gwForm.value.timeout }
        let gatewayId = editId.value
        if (isEdit.value) {
            data.status = gwForm.value.statusBool ? 1 : 0
            await gatewayApi.update(editId.value, data)
        } else {
            const res = await gatewayApi.create(data)
            gatewayId = res.data.id
        }

        // 更新模型绑定：将选中的模型绑到此网关，将取消选中的模型解绑
        const selectedIds = new Set(gwForm.value.modelIds)
        for (const m of allModels.value) {
            if (selectedIds.has(m.id) && m.gateway_id !== gatewayId) {
                await modelApi.update(m.id, { gateway_id: gatewayId })
            } else if (!selectedIds.has(m.id) && m.gateway_id === gatewayId) {
                await modelApi.update(m.id, { gateway_id: '' })
            }
        }

        ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
        dialogVisible.value = false
        fetchList()
    } finally {
        saving.value = false
    }
}

const handleDelete = async (id) => {
    await gatewayApi.remove(id)
    ElMessage.success('删除成功')
    fetchList()
}

const showPolicyDialog = (row) => {
    policyGatewayId.value = row.id
    const p = row.policy || {}
    policyForm.value = {
        rate_limit_enabled: p.rate_limit_enabled || false,
        rate_limit_qps: p.rate_limit_qps || 100,
        rate_limit_burst: p.rate_limit_burst || 200,
        circuit_breaker_enabled: p.circuit_breaker_enabled || false,
        circuit_breaker_threshold: p.circuit_breaker_threshold || 0.5,
        circuit_breaker_timeout: p.circuit_breaker_timeout || 30,
        circuit_breaker_min_reqs: p.circuit_breaker_min_reqs || 10,
        fallback_enabled: p.fallback_enabled || false,
        fallback_gateway_id: p.fallback_gateway_id || ''
    }
    policyVisible.value = true
}

const handleSavePolicy = async () => {
    savingPolicy.value = true
    try {
        await gatewayApi.updatePolicy(policyGatewayId.value, policyForm.value)
        ElMessage.success('策略已更新')
        policyVisible.value = false
        fetchList()
    } finally {
        savingPolicy.value = false
    }
}

onMounted(fetchList)
</script>

<style scoped>
.page-container { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
