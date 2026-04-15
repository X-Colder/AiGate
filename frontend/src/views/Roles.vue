<template>
    <div>
        <el-card>
            <template #header>
                <div style="display:flex;justify-content:space-between;align-items:center">
                    <span style="font-size:18px;font-weight:bold">角色管理</span>
                    <el-button type="primary" @click="openCreate">新增角色</el-button>
                </div>
            </template>
            <el-table :data="roles" stripe v-loading="loading">
                <el-table-column prop="name" label="角色名称" width="160" />
                <el-table-column prop="description" label="描述" width="200" />
                <el-table-column label="租户管理" width="110" align="center">
                    <template #default="{ row }">
                        <el-tag :type="row.tenant_access ? 'success' : 'info'" size="small">
                            {{ row.tenant_access ? '有' : '无' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column label="网关管理" width="110" align="center">
                    <template #default="{ row }">
                        <el-tag :type="row.gateway_access ? 'success' : 'info'" size="small">
                            {{ row.gateway_access ? '有' : '无' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column label="监控面板" width="110" align="center">
                    <template #default="{ row }">
                        <el-tag :type="row.monitor_access ? 'success' : 'info'" size="small">
                            {{ row.monitor_access ? '有' : '无' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="user_count" label="用户数" width="100" />
                <el-table-column label="系统角色" width="100" align="center">
                    <template #default="{ row }">
                        <el-tag v-if="row.is_system" type="warning" size="small">系统</el-tag>
                    </template>
                </el-table-column>
                <el-table-column label="操作" min-width="200">
                    <template #default="{ row }">
                        <el-button size="small" @click="openEdit(row)">编辑</el-button>
                        <el-button size="small" type="danger" @click="handleDelete(row)" :disabled="row.is_system">删除
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
        </el-card>

        <!-- 创建角色对话框 -->
        <el-dialog v-model="createVisible" title="新增角色" width="500px">
            <el-form :model="createForm" label-width="100px">
                <el-form-item label="角色名称" required>
                    <el-input v-model="createForm.name" placeholder="请输入角色名称" />
                </el-form-item>
                <el-form-item label="描述">
                    <el-input v-model="createForm.description" placeholder="请输入描述" />
                </el-form-item>
                <el-form-item label="租户管理">
                    <el-switch v-model="createForm.tenant_access" />
                </el-form-item>
                <el-form-item label="网关管理">
                    <el-switch v-model="createForm.gateway_access" />
                </el-form-item>
                <el-form-item label="监控面板">
                    <el-switch v-model="createForm.monitor_access" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="createVisible = false">取消</el-button>
                <el-button type="primary" @click="handleCreate" :loading="saving">确定</el-button>
            </template>
        </el-dialog>

        <!-- 编辑角色对话框 -->
        <el-dialog v-model="editVisible" title="编辑角色" width="500px">
            <el-form :model="editForm" label-width="100px">
                <el-form-item label="角色名称">
                    <el-input v-model="editForm.name" :disabled="editForm.is_system" />
                </el-form-item>
                <el-form-item label="描述">
                    <el-input v-model="editForm.description" />
                </el-form-item>
                <el-form-item label="租户管理">
                    <el-switch v-model="editForm.tenant_access" :disabled="editForm.is_system" />
                </el-form-item>
                <el-form-item label="网关管理">
                    <el-switch v-model="editForm.gateway_access" :disabled="editForm.is_system" />
                </el-form-item>
                <el-form-item label="监控面板">
                    <el-switch v-model="editForm.monitor_access" :disabled="editForm.is_system" />
                </el-form-item>
                <el-alert v-if="editForm.is_system" title="系统角色的权限设置不可修改" type="warning" :closable="false"
                    style="margin-top:10px" />
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
import { roleApi } from '../api'

const roles = ref([])
const loading = ref(false)
const saving = ref(false)

const createVisible = ref(false)
const editVisible = ref(false)
const createForm = ref({ name: '', description: '', tenant_access: false, gateway_access: true, monitor_access: true })
const editForm = ref({ id: '', name: '', description: '', tenant_access: false, gateway_access: false, monitor_access: false, is_system: false })

const loadRoles = async () => {
    loading.value = true
    try {
        const res = await roleApi.list()
        roles.value = res.data || []
    } catch (e) { /* */ } finally {
        loading.value = false
    }
}

const openCreate = () => {
    createForm.value = { name: '', description: '', tenant_access: false, gateway_access: true, monitor_access: true }
    createVisible.value = true
}

const handleCreate = async () => {
    if (!createForm.value.name) {
        ElMessage.warning('请输入角色名称')
        return
    }
    saving.value = true
    try {
        await roleApi.create(createForm.value)
        ElMessage.success('创建成功')
        createVisible.value = false
        loadRoles()
    } catch (e) { /* */ } finally {
        saving.value = false
    }
}

const openEdit = (row) => {
    editForm.value = {
        id: row.id,
        name: row.name,
        description: row.description,
        tenant_access: row.tenant_access,
        gateway_access: row.gateway_access,
        monitor_access: row.monitor_access,
        is_system: row.is_system
    }
    editVisible.value = true
}

const handleEdit = async () => {
    saving.value = true
    try {
        const data = { description: editForm.value.description }
        if (!editForm.value.is_system) {
            data.name = editForm.value.name
            data.tenant_access = editForm.value.tenant_access
            data.gateway_access = editForm.value.gateway_access
            data.monitor_access = editForm.value.monitor_access
        }
        await roleApi.update(editForm.value.id, data)
        ElMessage.success('更新成功')
        editVisible.value = false
        loadRoles()
    } catch (e) { /* */ } finally {
        saving.value = false
    }
}

const handleDelete = async (row) => {
    if (row.is_system) return
    try {
        await ElMessageBox.confirm(`确定删除角色「${row.name}」吗？`, '删除确认', { type: 'warning' })
        await roleApi.remove(row.id)
        ElMessage.success('删除成功')
        loadRoles()
    } catch (e) { /* */ }
}

onMounted(() => {
    loadRoles()
})
</script>