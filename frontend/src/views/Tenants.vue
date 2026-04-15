<template>
    <div>
        <el-card shadow="never">
            <template #header>
                <div style="display:flex;justify-content:space-between;align-items:center">
                    <span style="font-size:18px;font-weight:600;color:#1d3a5f">租户管理</span>
                    <el-button type="primary" @click="showCreate">
                        <el-icon>
                            <Plus />
                        </el-icon> 新增租户
                    </el-button>
                </div>
            </template>
            <el-table :data="tenants" stripe v-loading="loading" style="width:100%">
                <el-table-column prop="name" label="租户名称" min-width="130" />
                <el-table-column prop="email" label="邮箱" min-width="160" />
                <el-table-column prop="phone" label="电话" width="130" />
                <el-table-column prop="user_count" label="用户数" width="80" align="center" />
                <el-table-column prop="gateway_count" label="网关数" width="80" align="center" />
                <el-table-column prop="status" label="状态" width="80" align="center">
                    <template #default="{ row }">
                        <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                            {{ row.status === 1 ? '启用' : '禁用' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="created_at" label="创建时间" width="180">
                    <template #default="{ row }">
                        {{ formatTime(row.created_at) }}
                    </template>
                </el-table-column>
                <el-table-column label="操作" width="260" align="center">
                    <template #default="{ row }">
                        <el-button size="small" type="primary" link @click="viewUsage(row)">
                            <el-icon>
                                <View />
                            </el-icon> 详情
                        </el-button>
                        <el-button size="small" type="warning" link @click="showEdit(row)">
                            <el-icon>
                                <Edit />
                            </el-icon> 编辑
                        </el-button>
                        <el-popconfirm title="确定删除该租户？" @confirm="handleDelete(row.id)">
                            <template #reference>
                                <el-button size="small" type="danger" link>
                                    <el-icon>
                                        <Delete />
                                    </el-icon> 删除
                                </el-button>
                            </template>
                        </el-popconfirm>
                    </template>
                </el-table-column>
            </el-table>
        </el-card>

        <!-- 新增/编辑对话框 -->
        <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑租户' : '新增租户'" width="500px" destroy-on-close>
            <el-form :model="form" label-width="80px">
                <el-form-item label="租户名称">
                    <el-input v-model="form.name" placeholder="请输入租户名称" />
                </el-form-item>
                <el-form-item label="管理账号" v-if="!isEdit">
                    <el-input v-model="form.admin_user" placeholder="租户管理员登录用户名" />
                </el-form-item>
                <el-form-item label="密码" v-if="!isEdit">
                    <el-input v-model="form.password" type="password" placeholder="租户管理员密码" show-password />
                </el-form-item>
                <el-form-item label="邮箱">
                    <el-input v-model="form.email" placeholder="请输入邮箱" />
                </el-form-item>
                <el-form-item label="电话">
                    <el-input v-model="form.phone" placeholder="请输入电话" />
                </el-form-item>
                <el-form-item label="状态" v-if="isEdit">
                    <el-switch v-model="form.statusBool" active-text="启用" inactive-text="禁用" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="dialogVisible = false">取消</el-button>
                <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
            </template>
        </el-dialog>

        <!-- 使用详情对话框 -->
        <el-dialog v-model="usageVisible" title="租户使用详情" width="700px" destroy-on-close>
            <div v-loading="usageLoading">
                <div v-if="usage" style="margin-bottom:16px">
                    <el-descriptions :column="2" border size="small">
                        <el-descriptions-item label="租户名称">{{ usage.tenant_name }}</el-descriptions-item>
                        <el-descriptions-item label="总请求数">{{ usage.total_requests }}</el-descriptions-item>
                        <el-descriptions-item label="总Token消耗">{{ usage.total_tokens }}</el-descriptions-item>
                        <el-descriptions-item label="网关数量">{{ usage.gateways?.length || 0 }}</el-descriptions-item>
                    </el-descriptions>
                </div>
                <el-table :data="usage?.gateways || []" stripe size="small" style="width:100%">
                    <el-table-column prop="gateway_name" label="网关名称" min-width="120" />
                    <el-table-column prop="provider" label="服务商" width="100" />
                    <el-table-column prop="status" label="状态" width="80" align="center">
                        <template #default="{ row }">
                            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                                {{ row.status === 1 ? '在线' : '离线' }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column prop="tokens_used" label="Token消耗" width="100" align="right" />
                    <el-table-column prop="request_count" label="请求数" width="90" align="right" />
                    <el-table-column prop="avg_latency_ms" label="平均延迟(ms)" width="110" align="right">
                        <template #default="{ row }">
                            {{ row.avg_latency_ms?.toFixed(1) || '0.0' }}
                        </template>
                    </el-table-column>
                    <el-table-column prop="error_rate" label="错误率" width="90" align="right">
                        <template #default="{ row }">
                            {{ (row.error_rate * 100).toFixed(1) }}%
                        </template>
                    </el-table-column>
                </el-table>
            </div>
        </el-dialog>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { tenantApi } from '../api'

const tenants = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref('')
const submitting = ref(false)
const form = ref({ name: '', admin_user: '', password: '', email: '', phone: '', statusBool: true })

const usageVisible = ref(false)
const usageLoading = ref(false)
const usage = ref(null)

const formatTime = (t) => {
    if (!t) return '-'
    return new Date(t).toLocaleString('zh-CN')
}

const fetchTenants = async () => {
    loading.value = true
    try {
        const res = await tenantApi.list()
        tenants.value = res.data || []
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const showCreate = () => {
    isEdit.value = false
    form.value = { name: '', admin_user: '', password: '', email: '', phone: '', statusBool: true }
    dialogVisible.value = true
}

const showEdit = (row) => {
    isEdit.value = true
    editId.value = row.id
    form.value = { name: row.name, password: '', email: row.email || '', phone: row.phone || '', statusBool: row.status === 1 }
    dialogVisible.value = true
}

const handleSubmit = async () => {
    if (!form.value.name) {
        ElMessage.warning('请输入租户名称')
        return
    }
    if (!isEdit.value && !form.value.admin_user) {
        ElMessage.warning('请输入管理员用户名')
        return
    }
    if (!isEdit.value && !form.value.password) {
        ElMessage.warning('请输入密码')
        return
    }
    submitting.value = true
    try {
        if (isEdit.value) {
            await tenantApi.update(editId.value, {
                name: form.value.name,
                email: form.value.email,
                phone: form.value.phone,
                status: form.value.statusBool ? 1 : 0
            })
            ElMessage.success('更新成功')
        } else {
            await tenantApi.create({
                name: form.value.name,
                admin_user: form.value.admin_user,
                password: form.value.password,
                email: form.value.email,
                phone: form.value.phone
            })
            ElMessage.success(`创建成功，管理员账号: ${form.value.admin_user}`)
        }
        dialogVisible.value = false
        fetchTenants()
    } catch (e) {
        console.error(e)
    } finally {
        submitting.value = false
    }
}

const handleDelete = async (id) => {
    try {
        await tenantApi.remove(id)
        ElMessage.success('删除成功')
        fetchTenants()
    } catch (e) {
        console.error(e)
    }
}

const viewUsage = async (row) => {
    usageVisible.value = true
    usageLoading.value = true
    try {
        const res = await tenantApi.getUsage(row.id)
        usage.value = res.data
    } catch (e) {
        console.error(e)
    } finally {
        usageLoading.value = false
    }
}

onMounted(fetchTenants)
</script>

<style scoped>
.el-card {
    border-radius: 8px;
}
</style>