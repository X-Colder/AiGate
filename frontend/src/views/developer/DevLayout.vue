<template>
  <div class="dev-layout">
    <el-container style="height: 100vh">
      <el-aside width="200px" class="dev-aside">
        <div class="dev-logo">AiGate 开发者</div>
        <el-menu :default-active="$route.path" router class="dev-menu">
          <el-menu-item index="/developer/dashboard"><el-icon><Odometer /></el-icon><span>概览</span></el-menu-item>
          <el-menu-item index="/developer/apikeys"><el-icon><Key /></el-icon><span>API Keys</span></el-menu-item>
          <el-menu-item index="/developer/models"><el-icon><Cpu /></el-icon><span>模型列表</span></el-menu-item>
          <el-menu-item index="/developer/usage"><el-icon><DataAnalysis /></el-icon><span>用量统计</span></el-menu-item>
          <el-menu-item index="/developer/balance"><el-icon><Wallet /></el-icon><span>余额账单</span></el-menu-item>
        </el-menu>
      </el-aside>
      <el-main class="dev-main">
        <div class="dev-topbar">
          <span>{{ username }}</span>
          <el-button text @click="logout">退出</el-button>
        </div>
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Odometer, Key, Cpu, DataAnalysis, Wallet } from '@element-plus/icons-vue'

const router = useRouter()
const username = computed(() => localStorage.getItem('username') || '')
const logout = () => { localStorage.clear(); router.push('/login') }
</script>

<style scoped>
.dev-layout { height: 100vh; }
.dev-aside { background: #1d1e2c; }
.dev-logo { color: #fff; font-size: 16px; font-weight: bold; padding: 20px; text-align: center; border-bottom: 1px solid #2d2e3c; }
.dev-menu { background: #1d1e2c; border: none; }
.dev-menu .el-menu-item { color: #a0a3bd; }
.dev-menu .el-menu-item.is-active { color: #409eff; background: #252636; }
.dev-main { background: #f5f7fa; padding: 0; }
.dev-topbar { display: flex; justify-content: flex-end; align-items: center; gap: 12px; padding: 12px 24px; background: #fff; border-bottom: 1px solid #e4e7ed; }
</style>
