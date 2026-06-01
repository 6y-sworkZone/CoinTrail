<template>
  <el-container class="main-layout">
    <el-aside width="220px" class="sidebar">
      <div class="sidebar-logo">💰 CoinTrail</div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <el-menu-item index="/">
          <el-icon><DataAnalysis /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/accounts">
          <el-icon><Wallet /></el-icon>
          <span>账户管理</span>
        </el-menu-item>
        <el-menu-item index="/transactions">
          <el-icon><Document /></el-icon>
          <span>记账记录</span>
        </el-menu-item>
        <el-menu-item index="/categories">
          <el-icon><Collection /></el-icon>
          <span>分类管理</span>
        </el-menu-item>
        <el-menu-item index="/stats">
          <el-icon><TrendCharts /></el-icon>
          <span>统计报表</span>
        </el-menu-item>
        <el-menu-item index="/budgets">
          <el-icon><Tickets /></el-icon>
          <span>预算管理</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    
    <el-container>
      <el-header class="page-header">
        <span class="page-title">{{ pageTitle }}</span>
        <div style="display: flex; align-items: center; gap: 16px;">
          <span style="color: #606266;">欢迎，{{ userStore.user?.username }}</span>
          <el-button type="danger" size="small" @click="handleLogout">退出</el-button>
        </div>
      </el-header>
      
      <el-main class="page-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const activeMenu = computed(() => route.path)

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/': '仪表盘',
    '/accounts': '账户管理',
    '/transactions': '记账记录',
    '/categories': '分类管理',
    '/stats': '统计报表',
    '/budgets': '预算管理',
    '/settings': '设置'
  }
  return titles[route.path] || 'CoinTrail'
})

const handleLogout = () => {
  ElMessageBox.confirm('确定要退出登录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    userStore.logout()
    ElMessage.success('已退出登录')
    router.push('/login')
  }).catch(() => {})
}
</script>
