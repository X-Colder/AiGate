<template>
    <div class="login-container">
        <div class="login-card">
            <div class="login-header">
                <img src="" alt="" style="display:none" />
                <h1>AiGate</h1>
                <p>AI 网关管理平台</p>
            </div>
            <el-form ref="formRef" :model="form" :rules="rules" label-width="0">
                <el-form-item prop="username">
                    <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" size="large" />
                </el-form-item>
                <el-form-item prop="password">
                    <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" size="large"
                        show-password @keyup.enter="handleLogin" />
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" size="large" style="width:100%" :loading="loading" @click="handleLogin">登
                        录</el-button>
                </el-form-item>
            </el-form>
            <div class="login-footer">
                <span style="color:#909399;font-size:12px">默认账号: admin / admin123</span>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authApi } from '../api'

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const form = ref({ username: '', password: '' })

const rules = {
    username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
    password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    loading.value = true
    try {
        const res = await authApi.login(form.value)
        localStorage.setItem('token', res.data.token)
        localStorage.setItem('username', res.data.username)
        localStorage.setItem('tenant_id', res.data.tenant_id)
        localStorage.setItem('role', res.data.role)
        // 保存 RBAC 权限
        if (res.data.permissions) {
            localStorage.setItem('permissions', JSON.stringify(res.data.permissions))
        }
        ElMessage.success('登录成功')
        router.push('/')
    } catch (e) {
        // 错误已在拦截器处理
    } finally {
        loading.value = false
    }
}
</script>

<style scoped>
.login-container {
    width: 100%;
    height: 100vh;
    display: flex;
    justify-content: center;
    align-items: center;
    background: linear-gradient(135deg, #1d3a5f 0%, #409eff 50%, #ecf5ff 100%);
}

.login-card {
    width: 400px;
    padding: 40px;
    background: #fff;
    border-radius: 12px;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.15);
}

.login-header {
    text-align: center;
    margin-bottom: 30px;
}

.login-header h1 {
    font-size: 28px;
    color: #1d3a5f;
    margin-bottom: 8px;
    letter-spacing: 2px;
}

.login-header p {
    color: #909399;
    font-size: 14px;
}

.login-footer {
    text-align: center;
    margin-top: 10px;
}

.el-button--primary {
    background-color: #409eff;
    border-color: #409eff;
}

.el-button--primary:hover {
    background-color: #66b1ff;
    border-color: #66b1ff;
}
</style>