<template>
    <div>
        <el-card>
            <template #header>
                <div style="display:flex;justify-content:space-between;align-items:center">
                    <span style="font-size:18px;font-weight:bold">用户管理</span>
                    <el-button type="primary" @click="openCreate">新增用户</el-button>
                </div>
            </template>
            <el-table :data="users" stripe v-loading="loading">
                <el-table-column prop="username" label="用户名" width="150" />
                <el-table-column prop="tenant_name" label="所属租户" width="150" />
                <el-table-column prop="role_name" label="角色" width="150" />
                <el-table-column prop="role" label="类型" width="100">
                    <template #default="{ row }">
                        <el-tag :type="row.role === 'admin' ? 'danger' : 'info'" size="small">{{ row.role }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="100">
                    <template #default="{ row }">
                        <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                            {{ row.status === 1 ? '启用' : '禁用' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="created_at" label="创建时间" width="180" />
                <el-table-column label="操作" min-width="200">
                    <template #default="{ row }">
                        <el-button size="small" @click="openEdit(row)">编辑</el-button>
                        <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
                    </template>
                </el-table-column>
            </el-table>
        </el-card>

        <!-- 创建用户对话框 -->
        <el-dialog v-model="createVisible" title="新增用户" width="500px">
            <el-form :model="createForm" label-width="100px">
                <el-form-item label="用户名" required>
                    <el-input v-model="createForm.username" placeholder="请输入用户名" />
                </el-form-item>
                <el-form-item label="密码" required>
                    <el-input v-model="createForm.password" type="password" placeholder="请输入密码" show-password />
                </el-form-item>
                <el-form-item label="所属租户" required>
                    <el-select v-model="createForm.tenant_id" placeholder="选择租户" style="width:100%">
                        <el-option v-for="t in tenants" :key="t.id" :label="t.name" :value="t.id" />
                    </el-select>
                </el-form-item>
                <el-form-item label="角色" required>
                    <el-select v-model="createForm.role_id" placeholder="选择角色" style="width:100%">
                        <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
                    </el-select>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="createVisible = false">取消</el-button>
                <el-button type="primary" @click="handleCreate" :loading="saving">确定</el-button>
            </template>
        </el-dialog>

        <!-- 编辑用户对话框 -->
        <el-dialog v-model="editVisible" title="编辑用户" width="500px">
            <el-form :model="editForm" label-width="100px">
                <el-form-item label="用户名">
                    <el-input :value="editForm.username" disabled />
                </el-form-item>
                <el-form-item label="新密码">
                    <el-input v-model="editForm.password" type="password" placeholder="留空则不修改" show-password />
                </el-form-item>
                <el-form-item label="角色">
                    <el-select v-model="editForm.role_id" placeholder="选择角色" style="width:100%">
                        <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
                    </el-select>
                </el-form-item>
                <el-form-item label="状态">
                    <el-switch v-model="editForm.statusBool" active-text="启用" inactive-text="禁用" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="editVisible = false">取消</el-button>
                <el-button type="primary" @click="handleEdit" :loading="saving">确定</el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { userApi, roleApi, tenantApi } from '../api'

const users = ref([])
const roles = ref([])
const tenants = ref([])
const loading = ref(false)
const saving = ref(false)

const createVisible = ref(false)
const editVisible = ref(false)
const createForm = ref({ username: '', password: '', tenant_id: '', role_id: '' })
const editForm = ref({ id: '', username: '', password: '', role_id: '', statusBool: true })

const loadUsers = async () => {
    loading.value = true
    try {
        const res = await userApi.list()
        users.value = res.data || []
    } catch (e) { /* */ } finally {
        loading.value = false
    }
}

const loadRoles = async () => {
    try {
        const res = await roleApi.list()
        roles.value = res.data || []
    } catch (e) { /* */ }
}

const loadTenants = async () => {
    try {
        const res = await tenantApi.list()
        tenants.value = (res.data || []).map(t => ({ id: t.id, name: t.name }))
    } catch (e) { /* */ }
}

const openCreate = () => {
    createForm.value = { username: '', password: '', tenant_id: '', role_id: '' }
    createVisible.value = true
}

const handleCreate = async () => {
    if (!createForm.value.username || !createForm.value.password || !createForm.value.tenant_id || !createForm.value.role_id) {
        ElMessage.warning('请填写完整信息')
        return
    }
    saving.value = true
    try {
        await userApi.create(createForm.value)
        ElMessage.success('创建成功')
        createVisible.value = false
        loadUsers()
    } catch (e) { /* */ } finally {
        saving.value = false
    }
}

const openEdit = (row) => {
    editForm.value = {
        id: row.id,
        username: row.username,
        password: '',
        role_id: row.role_id,
        statusBool: row.status === 1
    }
    editVisible.value = true
}

const handleEdit = async () => {
    saving.value = true
    try {
        const data = {
            role_id: editForm.value.role_id,
            status: editForm.value.statusBool ? 1 : 0
        }
        if (editForm.value.password) {
            data.password = editForm.value.password
        }
        await userApi.update(editForm.value.id, data)
        ElMessage.success('更新成功')
        editVisible.value = false
        loadUsers()
    } catch (e) { /* */ } finally {
        saving.value = false
    }
}

const handleDelete = async (row) => {
    try {
        await ElMessageBox.confirm(`确定删除用户「${row.username}」吗？`, '删除确认', { type: 'warning' })
        await userApi.remove(row.id)
        ElMessage.success('删除成功')
        loadUsers()
    } catch (e) { /* */ }
}

onMounted(() => {
    loadUsers()
    loadRoles()
    loadTenants()
})
</script>