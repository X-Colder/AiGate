import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
    baseURL: '/api/v1',
    timeout: 30000
})

// 请求拦截：注入 JWT Token
api.interceptors.request.use(config => {
    const token = localStorage.getItem('token')
    if (token) {
        config.headers.Authorization = `Bearer ${token}`
    }
    return config
})

// 响应拦截：统一错误处理
api.interceptors.response.use(
    response => response.data,
    error => {
        const msg = error.response?.data?.message || error.message
        if (error.response?.status === 401) {
            localStorage.removeItem('token')
            window.location.href = '/login'
        } else {
            ElMessage.error(msg)
        }
        return Promise.reject(error)
    }
)

// 认证
export const authApi = {
    login: (data) => api.post('/auth/login', data),
    register: (data) => api.post('/auth/register', data)
}

// 网关管理
export const gatewayApi = {
    list: () => api.get('/gateways'),
    getById: (id) => api.get(`/gateways/${id}`),
    create: (data) => api.post('/gateways', data),
    update: (id, data) => api.put(`/gateways/${id}`, data),
    remove: (id) => api.delete(`/gateways/${id}`),
    updatePolicy: (id, data) => api.put(`/gateways/${id}/policy`, data)
}

// 监控
export const metricApi = {
    getSummary: (params) => api.get('/metrics/summary', { params }),
    getTrend: (params) => api.get('/metrics/trend', { params })
}

// 租户管理（管理员专用）
export const tenantApi = {
    list: () => api.get('/admin/tenants'),
    getById: (id) => api.get(`/admin/tenants/${id}`),
    create: (data) => api.post('/admin/tenants', data),
    update: (id, data) => api.put(`/admin/tenants/${id}`, data),
    remove: (id) => api.delete(`/admin/tenants/${id}`),
    getUsage: (id) => api.get(`/admin/tenants/${id}/usage`)
}

export default api