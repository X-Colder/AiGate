<template>
    <el-container style="height: 100vh">
        <!-- 侧边栏 -->
        <el-aside width="220px" style="background-color: #1d3a5f">
            <div class="logo">
                <span>AiGate</span>
            </div>
            <el-menu :default-active="$route.path" router background-color="#1d3a5f" text-color="#ffffffcc"
                active-text-color="#409eff">
                <el-menu-item index="/tenants" v-if="isAdmin">
                    <el-icon>
                        <OfficeBuilding />
                    </el-icon>
                    <span>租户管理</span>
                </el-menu-item>
                <el-menu-item index="/gateways">
                    <el-icon>
                        <Connection />
                    </el-icon>
                    <span>网关管理</span>
                </el-menu-item>
                <el-menu-item index="/monitor">
                    <el-icon>
                        <DataLine />
                    </el-icon>
                    <span>监控面板</span>
                </el-menu-item>
            </el-menu>
        </el-aside>
        <!-- 主区域 -->
        <el-container>
            <el-header
                style="background:#fff;display:flex;justify-content:flex-end;align-items:center;box-shadow:0 1px 4px rgba(0,0,0,0.08)">
                <el-dropdown>
                    <span style="cursor:pointer;display:flex;align-items:center;color:#1d3a5f">
                        <el-icon style="margin-right:6px">
                            <User />
                        </el-icon>
                        {{ username }}
                        <el-icon style="margin-left:4px">
                            <ArrowDown />
                        </el-icon>
                    </span>
                    <template #dropdown>
                        <el-dropdown-menu>
                            <el-dropdown-item @click="handleLogout">
                                <el-icon>
                                    <SwitchButton />
                                </el-icon>
                                退出登录
                            </el-dropdown-item>
                        </el-dropdown-menu>
                    </template>
                </el-dropdown>
            </el-header>
            <el-main style="background:#f0f5ff;overflow-y:auto">
                <router-view />
            </el-main>
        </el-container>
    </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = computed(() => localStorage.getItem('username') || 'User')
const isAdmin = computed(() => localStorage.getItem('role') === 'admin')

const handleLogout = () => {
    localStorage.clear()
    router.push('/login')
}
</script>

<style scoped>
.logo {
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 22px;
    font-weight: bold;
    letter-spacing: 3px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.el-aside {
    transition: width 0.3s;
}

.el-menu-item {
    font-size: 14px;
}

.el-menu-item.is-active {
    background-color: rgba(64, 158, 255, 0.15) !important;
}

.el-menu-item:hover {
    background-color: rgba(255, 255, 255, 0.08) !important;
}
</style>