import { createRouter, createWebHistory } from 'vue-router'

const routes = [
    {
        path: '/login',
        name: 'Login',
        component: () => import('../views/Login.vue')
    },
    {
        path: '/',
        name: 'Layout',
        component: () => import('../views/Layout.vue'),
        redirect: '/gateways',
        children: [
            {
                path: 'tenants',
                name: 'Tenants',
                component: () => import('../views/Tenants.vue'),
                meta: { permission: 'tenant_access' }
            },
            {
                path: 'users',
                name: 'Users',
                component: () => import('../views/Users.vue'),
                meta: { permission: 'tenant_access' }
            },
            {
                path: 'roles',
                name: 'Roles',
                component: () => import('../views/Roles.vue'),
                meta: { permission: 'tenant_access' }
            },
            {
                path: 'gateways',
                name: 'Gateways',
                component: () => import('../views/Gateways.vue'),
                meta: { permission: 'gateway_access' }
            },
            {
                path: 'monitor',
                name: 'Monitor',
                component: () => import('../views/Monitor.vue'),
                meta: { permission: 'monitor_access' }
            }
        ]
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

// 路由守卫：按 RBAC permissions 控制页面访问
router.beforeEach((to, from, next) => {
    const token = localStorage.getItem('token')
    if (to.path !== '/login' && !token) {
        next('/login')
        return
    }

    if (to.meta.permission) {
        try {
            const permissions = JSON.parse(localStorage.getItem('permissions') || '{}')
            if (!permissions[to.meta.permission]) {
                // 无权限则跳转到第一个有权限的页面
                next('/gateways')
                return
            }
        } catch {
            next('/gateways')
            return
        }
    }

    next()
})

export default router