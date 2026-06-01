<template>
  <div>
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="6">
        <div class="stat-card">
          <div style="display: flex; align-items: center; justify-content: space-between;">
            <div>
              <div class="stat-card-title">总资产</div>
              <div class="stat-card-value">¥{{ formatNumber(stats.total_balance) }}</div>
            </div>
            <div class="stat-card-icon" style="background: rgba(64, 158, 255, 0.1); color: #409eff;">
              💎
            </div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div style="display: flex; align-items: center; justify-content: space-between;">
            <div>
              <div class="stat-card-title">本月收入</div>
              <div class="stat-card-value" style="color: #67c23a;">¥{{ formatNumber(stats.monthly_income) }}</div>
            </div>
            <div class="stat-card-icon" style="background: rgba(103, 194, 58, 0.1); color: #67c23a;">
              📈
            </div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div style="display: flex; align-items: center; justify-content: space-between;">
            <div>
              <div class="stat-card-title">本月支出</div>
              <div class="stat-card-value" style="color: #f56c6c;">¥{{ formatNumber(stats.monthly_expense) }}</div>
            </div>
            <div class="stat-card-icon" style="background: rgba(245, 108, 108, 0.1); color: #f56c6c;">
              📉
            </div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div style="display: flex; align-items: center; justify-content: space-between;">
            <div>
              <div class="stat-card-title">本月结余</div>
              <div class="stat-card-value" :style="{ color: stats.monthly_balance >= 0 ? '#67c23a' : '#f56c6c' }">
                ¥{{ formatNumber(stats.monthly_balance) }}
              </div>
            </div>
            <div class="stat-card-icon" :style="{ background: stats.monthly_balance >= 0 ? 'rgba(103, 194, 58, 0.1)' : 'rgba(245, 108, 108, 0.1)', color: stats.monthly_balance >= 0 ? '#67c23a' : '#f56c6c' }">
              {{ stats.monthly_balance >= 0 ? '💰' : '⚠️' }}
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="16">
        <div class="card" style="margin-bottom: 20px;">
          <h3 style="margin-bottom: 16px;">近12个月收支趋势</h3>
          <div ref="trendChartRef" class="chart-container"></div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="card" style="margin-bottom: 20px;">
          <h3 style="margin-bottom: 16px;">今日收支</h3>
          <div style="text-align: center; padding: 20px 0;">
            <div style="margin-bottom: 16px;">
              <div style="color: #909399; font-size: 14px;">今日收入</div>
              <div style="font-size: 24px; font-weight: 600; color: #67c23a;">¥{{ formatNumber(stats.today_income) }}</div>
            </div>
            <div>
              <div style="color: #909399; font-size: 14px;">今日支出</div>
              <div style="font-size: 24px; font-weight: 600; color: #f56c6c;">¥{{ formatNumber(stats.today_expense) }}</div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="24">
        <div class="card">
          <h3 style="margin-bottom: 16px;">预算使用进度</h3>
          <div v-if="budgets.length === 0" style="text-align: center; padding: 40px; color: #909399;">
            暂无预算数据，快去设置预算吧！
          </div>
          <div v-else>
            <div v-for="budget in budgets" :key="budget.id" class="budget-item" style="margin-bottom: 16px;">
              <div class="budget-header">
                <div class="budget-info">
                  <span class="budget-icon">{{ budget.category_icon }}</span>
                  <span style="font-weight: 500;">{{ budget.category_name }}</span>
                </div>
                <div class="budget-amounts">
                  <span>¥{{ formatNumber(budget.used_amount) }} / ¥{{ formatNumber(budget.amount) }}</span>
                  <span :class="{ 'over-budget': budget.is_over_budget }">
                    {{ budget.percentage.toFixed(1) }}%
                  </span>
                </div>
              </div>
              <el-progress
                :percentage="budget.percentage"
                :color="budget.is_over_budget ? '#f56c6c' : budget.category_color"
                :stroke-width="8"
              />
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import type { DashboardStats, BudgetWithUsage, MonthlySummary } from '@/types'
import { getDashboardStats } from '@/api/stats'
import { getDashboardBudgets } from '@/api/budget'

const trendChartRef = ref<HTMLElement>()
let trendChart: echarts.ECharts | null = null

const stats = ref<DashboardStats>({
  total_balance: 0,
  monthly_income: 0,
  monthly_expense: 0,
  monthly_balance: 0,
  today_income: 0,
  today_expense: 0
})

const budgets = ref<BudgetWithUsage[]>([])

const formatNumber = (num: number) => {
  return num.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

const loadData = async () => {
  try {
    const [statsData, budgetsData] = await Promise.all([
      getDashboardStats(),
      getDashboardBudgets()
    ])
    stats.value = statsData
    budgets.value = budgetsData
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
  }
}

const initTrendChart = async () => {
  if (!trendChartRef.value) return
  
  await nextTick()
  
  trendChart = echarts.init(trendChartRef.value)
  
  const months = []
  const now = new Date()
  for (let i = 11; i >= 0; i--) {
    const date = new Date(now.getFullYear(), now.getMonth() - i, 1)
    months.push(`${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`)
  }
  
  const option = {
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['收入', '支出', '结余']
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: months
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '收入',
        type: 'line',
        smooth: true,
        data: Array(12).fill(0),
        itemStyle: { color: '#67c23a' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(103, 194, 58, 0.3)' },
            { offset: 1, color: 'rgba(103, 194, 58, 0.05)' }
          ])
        }
      },
      {
        name: '支出',
        type: 'line',
        smooth: true,
        data: Array(12).fill(0),
        itemStyle: { color: '#f56c6c' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(245, 108, 108, 0.3)' },
            { offset: 1, color: 'rgba(245, 108, 108, 0.05)' }
          ])
        }
      }
    ]
  }
  
  trendChart.setOption(option)
}

onMounted(() => {
  loadData()
  initTrendChart()
})
</script>
