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
                meta: { requiresAdmin: true }
            },
            {
                path: 'gateways',
                name: 'Gateways',
                component: () => import('../views/Gateways.vue')
            },
            {
                path: 'monitor',
                name: 'Monitor',
                component: () => import('../views/Monitor.vue')
            }
        ]
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

// 路由守卫：未登录跳转到登录页，非管理员禁止访问管理页
router.beforeEach((to, from, next) => {
    const token = localStorage.getItem('token')
    if (to.path !== '/login' && !token) {
        next('/login')
    } else if (to.meta.requiresAdmin && localStorage.getItem('role') !== 'admin') {
        next('/gateways')
    } else {
        next()
    }
})

export default router