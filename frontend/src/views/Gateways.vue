<template>
    <div class="page-container">
        <el-card shadow="hover">
            <template #header>
                <div class="card-header">
                    <span style="font-size:18px;font-weight:bold;color:#1d3a5f">网关管理</span>
                    <el-button type="primary" @click="showCreateDialog">
                        <el-icon>
                            <Plus />
                        </el-icon> 新增网关
                    </el-button>
                </div>
            </template>
            <!-- 网关列表 -->
            <el-table :data="gateways" stripe style="width:100%" v-loading="loading">
                <el-table-column prop="name" label="网关名称" min-width="120" />
                <el-table-column prop="provider" label="AI 服务" width="120">
                    <template #default="{ row }">
                        <el-tag type="primary">{{ row.provider }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="model" label="模型" width="150" />
                <el-table-column prop="base_url" label="API 地址" min-width="200" show-overflow-tooltip />
                <el-table-column label="状态" width="80" align="center">
                    <template #default="{ row }">
                        <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                            {{ row.status === 1 ? '启用' : '禁用' }}
                        </el-tag>
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
        <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑网关' : '新增网关'" width="520px" destroy-on-close>
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
                <el-form-item label="API Key">
                    <el-input v-model="gwForm.api_key" placeholder="sk-..." show-password />
                </el-form-item>
                <el-form-item label="模型">
                    <el-input v-model="gwForm.model" placeholder="模型名称" />
                </el-form-item>
                <el-form-item label="超时(秒)">
                    <el-input-number v-model="gwForm.timeout" :min="5" :max="300" />
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
                    <el-slider v-model="policyForm.circuit_breaker_threshold" :min="0" :max="1" :step="0.05"
                        show-input />
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
                <el-form-item label="降级服务" v-if="policyForm.fallback_enabled">
                    <el-select v-model="policyForm.fallback_provider" style="width:100%">
                        <el-option v-for="p in providers" :key="p" :label="p" :value="p" />
                    </el-select>
                </el-form-item>
                <el-form-item label="降级模型" v-if="policyForm.fallback_enabled">
                    <el-input v-model="policyForm.fallback_model" placeholder="降级使用的模型" />
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
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { gatewayApi } from '../api'

const providers = ['openai', 'anthropic', 'deepseek', 'doubao', 'qwen', 'kimi']
const gateways = ref([])
const loading = ref(false)

// 网关表单
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref('')
const saving = ref(false)
const gwFormRef = ref()
const gwForm = ref({ name: '', provider: '', base_url: '', api_key: '', model: '', timeout: 60, statusBool: true })
const gwRules = {
    name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
    provider: [{ required: true, message: '请选择服务', trigger: 'change' }]
}

// 策略表单
const policyVisible = ref(false)
const policyGatewayId = ref('')
const savingPolicy = ref(false)
const policyForm = ref({
    rate_limit_enabled: false, rate_limit_qps: 100, rate_limit_burst: 200,
    circuit_breaker_enabled: false, circuit_breaker_threshold: 0.5, circuit_breaker_timeout: 30, circuit_breaker_min_reqs: 10,
    fallback_enabled: false, fallback_provider: '', fallback_model: ''
})

const fetchList = async () => {
    loading.value = true
    try {
        const res = await gatewayApi.list()
        gateways.value = res.data || []
    } finally {
        loading.value = false
    }
}

const showCreateDialog = () => {
    isEdit.value = false
    editId.value = ''
    gwForm.value = { name: '', provider: '', base_url: '', api_key: '', model: '', timeout: 60, statusBool: true }
    dialogVisible.value = true
}

const showEditDialog = (row) => {
    isEdit.value = true
    editId.value = row.id
    gwForm.value = { name: row.name, provider: row.provider, base_url: row.base_url, api_key: '', model: row.model, timeout: row.timeout, statusBool: row.status === 1 }
    dialogVisible.value = true
}

const handleSave = async () => {
    const valid = await gwFormRef.value.validate().catch(() => false)
    if (!valid) return

    saving.value = true
    try {
        const data = { ...gwForm.value }
        if (isEdit.value) {
            data.status = data.statusBool ? 1 : 0
            delete data.statusBool
            await gatewayApi.update(editId.value, data)
            ElMessage.success('更新成功')
        } else {
            delete data.statusBool
            await gatewayApi.create(data)
            ElMessage.success('创建成功')
        }
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
        fallback_provider: p.fallback_provider || '',
        fallback_model: p.fallback_model || ''
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