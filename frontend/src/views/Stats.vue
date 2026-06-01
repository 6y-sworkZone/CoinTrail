<template>
  <div>
    <div style="margin-bottom: 16px; display: flex; gap: 12px; align-items: center;">
      <el-date-picker
        v-model="selectedMonth"
        type="month"
        value-format="YYYY-MM"
        placeholder="选择月份"
        style="width: 160px;"
        @change="loadData"
      />
      <span style="color: #909399;">当前统计月份：{{ selectedMonth }}</span>
    </div>

    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="8">
        <div class="stat-card" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white;">
          <div class="stat-card-title" style="color: rgba(255,255,255,0.8);">本月收入</div>
          <div class="stat-card-value" style="color: white;">¥{{ formatNumber(monthlyIncome) }}</div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stat-card" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white;">
          <div class="stat-card-title" style="color: rgba(255,255,255,0.8);">本月支出</div>
          <div class="stat-card-value" style="color: white;">¥{{ formatNumber(monthlyExpense) }}</div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stat-card" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); color: white;">
          <div class="stat-card-title" style="color: rgba(255,255,255,0.8);">本月结余</div>
          <div class="stat-card-value" style="color: white;">¥{{ formatNumber(monthlyIncome - monthlyExpense) }}</div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="12">
        <div class="card" style="margin-bottom: 20px;">
          <h3 style="margin-bottom: 16px;">收支趋势（近12个月）</h3>
          <div ref="trendChartRef" class="chart-container"></div>
        </div>
      </el-col>
      <el-col :span="12">
        <div class="card" style="margin-bottom: 20px;">
          <h3 style="margin-bottom: 16px;">月度收支对比</h3>
          <div ref="barChartRef" class="chart-container"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="12">
        <div class="card" style="margin-bottom: 20px;">
          <h3 style="margin-bottom: 16px;">
            <el-radio-group v-model="pieType" size="small" @change="loadCategoryData">
              <el-radio-button label="expense">支出分类占比</el-radio-button>
              <el-radio-button label="income">收入分类占比</el-radio-button>
            </el-radio-group>
          </h3>
          <div ref="pieChartRef" class="chart-container"></div>
        </div>
      </el-col>
      <el-col :span="12">
        <div class="card" style="margin-bottom: 20px;">
          <h3 style="margin-bottom: 16px;">分类明细</h3>
          <div v-if="categoryDetails.length === 0" style="text-align: center; padding: 40px; color: #909399;">
            暂无数据
          </div>
          <div v-else>
            <div v-for="item in categoryDetails" :key="item.category_id" style="margin-bottom: 16px;">
              <div style="display: flex; justify-content: space-between; margin-bottom: 8px;">
                <span>{{ item.category_icon || '📝' }} {{ item.category_name || '未分类' }}</span>
                <span style="font-weight: 600;">¥{{ formatNumber(item.amount) }} ({{ item.percentage.toFixed(1) }}%)</span>
              </div>
              <el-progress
                :percentage="item.percentage"
                :color="item.category_color || '#409eff'"
                :stroke-width="8"
                :show-text="false"
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
import dayjs from 'dayjs'
import type { CategorySummary } from '@/types'
import { getMonthlySummary, getCategorySummary } from '@/api/stats'

const trendChartRef = ref<HTMLElement>()
const barChartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()

let trendChart: echarts.ECharts | null = null
let barChart: echarts.ECharts | null = null
let pieChart: echarts.ECharts | null = null

const selectedMonth = ref(dayjs().format('YYYY-MM'))
const pieType = ref<'expense' | 'income'>('expense')
const monthlyIncome = ref(0)
const monthlyExpense = ref(0)
const categoryDetails = ref<CategorySummary[]>([])

const formatNumber = (num: number) => {
  return num.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

const loadTrendData = async () => {
  try {
    const data = await getMonthlySummary()
    
    const months = data.map(d => d.month)
    const incomes = data.map(d => d.income)
    const expenses = data.map(d => d.expense)

    if (trendChart) {
      trendChart.setOption({
        tooltip: { trigger: 'axis' },
        legend: { data: ['收入', '支出'] },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: { type: 'category', boundaryGap: false, data: months },
        yAxis: { type: 'value' },
        series: [
          {
            name: '收入',
            type: 'line',
            smooth: true,
            data: incomes,
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
            data: expenses,
            itemStyle: { color: '#f56c6c' },
            areaStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: 'rgba(245, 108, 108, 0.3)' },
                { offset: 1, color: 'rgba(245, 108, 108, 0.05)' }
              ])
            }
          }
        ]
      })
    }

    if (barChart) {
      barChart.setOption({
        tooltip: { trigger: 'axis' },
        legend: { data: ['收入', '支出'] },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: { type: 'category', data: months },
        yAxis: { type: 'value' },
        series: [
          { name: '收入', type: 'bar', data: incomes, itemStyle: { color: '#67c23a' } },
          { name: '支出', type: 'bar', data: expenses, itemStyle: { color: '#f56c6c' } }
        ]
      })
    }
  } catch (error) {
    console.error('Failed to load trend data:', error)
  }
}

const loadCategoryData = async () => {
  try {
    const result = await getCategorySummary(pieType.value, selectedMonth.value)
    categoryDetails.value = result.details

    if (pieChart) {
      const pieData = result.details.map(d => ({
        value: d.amount,
        name: d.category_name || '未分类',
        itemStyle: { color: d.category_color }
      }))

      pieChart.setOption({
        tooltip: { trigger: 'item', formatter: '{b}: ¥{c} ({d}%)' },
        legend: { orient: 'vertical', left: 'left' },
        series: [{
          type: 'pie',
          radius: ['40%', '70%'],
          avoidLabelOverlap: false,
          itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
          label: { show: false },
          emphasis: { label: { show: true, fontSize: 16, fontWeight: 'bold' } },
          labelLine: { show: false },
          data: pieData
        }]
      })
    }
  } catch (error) {
    console.error('Failed to load category data:', error)
  }
}

const loadMonthlyTotal = async () => {
  try {
    const result = await getCategorySummary('expense', selectedMonth.value)
    monthlyExpense.value = result.total
    
    const incomeResult = await getCategorySummary('income', selectedMonth.value)
    monthlyIncome.value = incomeResult.total
  } catch (error) {
    console.error('Failed to load monthly total:', error)
  }
}

const loadData = () => {
  loadTrendData()
  loadCategoryData()
  loadMonthlyTotal()
}

onMounted(async () => {
  await nextTick()
  
  if (trendChartRef.value) {
    trendChart = echarts.init(trendChartRef.value)
  }
  if (barChartRef.value) {
    barChart = echarts.init(barChartRef.value)
  }
  if (pieChartRef.value) {
    pieChart = echarts.init(pieChartRef.value)
  }
  
  loadData()
})
</script>
