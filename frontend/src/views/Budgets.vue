<template>
  <div>
    <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center;">
      <el-date-picker
        v-model="selectedMonth"
        type="month"
        value-format="YYYY-MM"
        placeholder="选择月份"
        style="width: 160px;"
        @change="loadBudgets"
      />
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        新建预算
      </el-button>
    </div>

    <div v-if="budgets.length === 0" class="card" style="text-align: center; padding: 60px;">
      <el-empty description="暂无预算，点击上方按钮创建第一个预算吧" />
    </div>

    <div v-else>
      <div v-for="budget in budgets" :key="budget.id" class="budget-item">
        <div class="budget-header">
          <div class="budget-info">
            <span class="budget-icon" :style="{ background: budget.category_color + '20', color: budget.category_color }">
              {{ budget.category_icon }}
            </span>
            <div>
              <div style="font-weight: 500;">{{ budget.category_name }}</div>
              <div style="font-size: 12px; color: #909399;">{{ budget.month }}</div>
            </div>
          </div>
          <div class="budget-amounts">
            <div style="text-align: right;">
              <div>已用: ¥{{ formatNumber(budget.used_amount) }}</div>
              <div style="font-size: 12px; color: #909399;">
                预算: ¥{{ formatNumber(budget.amount) }}
              </div>
            </div>
            <div style="text-align: right; margin-left: 16px;">
              <div :class="{ 'over-budget': budget.is_over_budget }">
                {{ budget.percentage.toFixed(1) }}%
              </div>
              <div style="font-size: 12px; color: #909399;">
                剩余: ¥{{ formatNumber(budget.remaining) }}
              </div>
            </div>
            <div style="margin-left: 16px; display: flex; gap: 8px;">
              <el-button size="small" @click="openEditDialog(budget)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDelete(budget)">删除</el-button>
            </div>
          </div>
        </div>
        <el-progress
          :percentage="budget.percentage"
          :color="budget.is_over_budget ? '#f56c6c' : budget.category_color"
          :stroke-width="10"
        />
        <div v-if="budget.is_over_budget" style="margin-top: 8px; color: #f56c6c; font-size: 12px;">
          ⚠️ 已超出预算 ¥{{ formatNumber(budget.used_amount - budget.amount) }}
        </div>
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="editingBudget ? '编辑预算' : '新建预算'" width="500px">
      <el-form :model="budgetForm" :rules="budgetRules" ref="budgetFormRef" label-width="80px">
        <el-form-item label="月份" prop="month">
          <el-date-picker
            v-model="budgetForm.month"
            type="month"
            value-format="YYYY-MM"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="分类" prop="category_id">
          <el-select v-model="budgetForm.category_id" placeholder="请选择支出分类" style="width: 100%;">
            <el-option
              v-for="cat in expenseCategories"
              :key="cat.id"
              :label="`${cat.icon} ${cat.name}`"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="预算金额" prop="amount">
          <el-input-number v-model="budgetForm.amount" :precision="2" :min="0" style="width: 100%;" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveBudget" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, type FormInstance, type FormRules } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import type { BudgetWithUsage, Category } from '@/types'
import { getBudgets, createBudget, updateBudget, deleteBudget, type CreateBudgetData } from '@/api/budget'
import { getCategories } from '@/api/category'

const budgets = ref<BudgetWithUsage[]>([])
const expenseCategories = ref<Category[]>([])
const selectedMonth = ref(dayjs().format('YYYY-MM'))
const dialogVisible = ref(false)
const editingBudget = ref<BudgetWithUsage | null>(null)
const saving = ref(false)

const budgetFormRef = ref<FormInstance>()

const budgetForm = reactive<CreateBudgetData>({
  category_id: 0,
  amount: 0,
  month: dayjs().format('YYYY-MM')
})

const budgetRules: FormRules = {
  month: [{ required: true, message: '请选择月份', trigger: 'change' }],
  category_id: [{ required: true, message: '请选择分类', trigger: 'change' }],
  amount: [{ required: true, message: '请输入预算金额', trigger: 'blur' }]
}

const formatNumber = (num: number) => {
  return num.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

const loadBudgets = async () => {
  try {
    budgets.value = await getBudgets(selectedMonth.value)
  } catch (error) {
    console.error('Failed to load budgets:', error)
  }
}

const loadCategories = async () => {
  try {
    expenseCategories.value = await getCategories('expense')
    if (expenseCategories.value.length > 0) {
      budgetForm.category_id = expenseCategories.value[0].id
    }
  } catch (error) {
    console.error('Failed to load categories:', error)
  }
}

const openCreateDialog = () => {
  editingBudget.value = null
  budgetForm.month = selectedMonth.value
  budgetForm.category_id = expenseCategories.value[0]?.id || 0
  budgetForm.amount = 0
  dialogVisible.value = true
}

const openEditDialog = (budget: BudgetWithUsage) => {
  editingBudget.value = budget
  budgetForm.month = budget.month
  budgetForm.category_id = budget.category_id
  budgetForm.amount = budget.amount
  dialogVisible.value = true
}

const handleSaveBudget = async () => {
  if (!budgetFormRef.value) return
  
  try {
    await budgetFormRef.value.validate()
    saving.value = true
    
    if (editingBudget.value) {
      await updateBudget(editingBudget.value.id, { amount: budgetForm.amount })
      ElMessage.success('预算更新成功')
    } else {
      await createBudget(budgetForm)
      ElMessage.success('预算创建成功')
    }
    
    dialogVisible.value = false
    loadBudgets()
  } catch (error) {
    // Error handled by interceptor
  } finally {
    saving.value = false
  }
}

const handleDelete = async (budget: BudgetWithUsage) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除"${budget.category_name}"的预算吗？`,
      '提示',
      { type: 'warning' }
    )
    
    await deleteBudget(budget.id)
    ElMessage.success('预算删除成功')
    loadBudgets()
  } catch (error) {
    // User cancelled or error
  }
}

onMounted(() => {
  loadBudgets()
  loadCategories()
})
</script>
