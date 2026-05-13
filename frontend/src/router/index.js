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
            },
            {
                path: 'models',
                name: 'Models',
                component: () => import('../views/Models.vue'),
                meta: { permission: 'tenant_access' }
            },
            {
                path: 'billing',
                name: 'Billing',
                component: () => import('../views/Billing.vue'),
                meta: { permission: 'tenant_access' }
            }
        ]
    },
    {
        path: '/developer',
        name: 'DevLayout',
        component: () => import('../views/developer/DevLayout.vue'),
        redirect: '/developer/dashboard',
        children: [
            {
                path: 'dashboard',
                name: 'DevDashboard',
                component: () => import('../views/developer/Dashboard.vue'),
                meta: { permission: 'api_access' }
            },
            {
                path: 'apikeys',
                name: 'DevAPIKeys',
                component: () => import('../views/developer/APIKeys.vue'),
                meta: { permission: 'api_access' }
            },
            {
                path: 'models',
                name: 'DevModels',
                component: () => import('../views/developer/ModelList.vue'),
                meta: { permission: 'api_access' }
            },
            {
                path: 'models/:id/doc',
                name: 'DevModelDoc',
                component: () => import('../views/developer/ModelDoc.vue'),
                meta: { permission: 'api_access' }
            },
            {
                path: 'usage',
                name: 'DevUsage',
                component: () => import('../views/developer/Usage.vue'),
                meta: { permission: 'api_access' }
            },
            {
                path: 'balance',
                name: 'DevBalance',
                component: () => import('../views/developer/Balance.vue'),
                meta: { permission: 'api_access' }
            }
        ]
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

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
                if (permissions.api_access) {
                    next('/developer/dashboard')
                } else {
                    next('/gateways')
                }
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
