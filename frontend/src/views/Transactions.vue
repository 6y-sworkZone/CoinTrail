<template>
  <div>
    <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center;">
      <div style="display: flex; gap: 12px;">
        <el-button type="primary" @click="openCreateDialog('expense')">
          <el-icon><Minus /></el-icon>
          记支出
        </el-button>
        <el-button type="success" @click="openCreateDialog('income')">
          <el-icon><Plus /></el-icon>
          记收入
        </el-button>
        <el-button type="danger" @click="handleBatchDelete" :disabled="selectedIds.length === 0">
          <el-icon><Delete /></el-icon>
          批量删除
        </el-button>
      </div>
      <div style="display: flex; gap: 12px;">
        <el-select v-model="filterType" placeholder="类型" clearable style="width: 100px;">
          <el-option label="支出" value="expense" />
          <el-option label="收入" value="income" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          style="width: 240px;"
        />
        <el-button @click="loadTransactions">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>

    <div class="card" style="padding: 0;">
      <el-table
        :data="transactions"
        v-loading="loading"
        @selection-change="handleSelectionChange"
        style="width: 100%"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column label="日期" width="120">
          <template #default="{ row }">
            {{ formatDate(row.transaction_date) }}
          </template>
        </el-table-column>
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.type === 'income' ? 'success' : 'danger'" size="small">
              {{ row.type === 'income' ? '收入' : '支出' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="120">
          <template #default="{ row }">
            <span>{{ row.category_icon || '📝' }} {{ row.category_name || '未分类' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="账户" width="120">
          <template #default="{ row }">
            {{ row.account_name }}
          </template>
        </el-table-column>
        <el-table-column label="金额" width="120">
          <template #default="{ row }">
            <span :style="{ color: row.type === 'income' ? '#67c23a' : '#f56c6c', fontWeight: '600' }">
              {{ row.type === 'income' ? '+' : '-' }}¥{{ formatNumber(row.amount) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="备注" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.note || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div style="padding: 16px; display: flex; justify-content: flex-end;">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="transactionForm" :rules="transactionRules" ref="transactionFormRef" label-width="80px">
        <el-form-item label="账户" prop="account_id">
          <el-select v-model="transactionForm.account_id" placeholder="请选择账户" style="width: 100%;">
            <el-option v-for="acc in accounts" :key="acc.id" :label="acc.name" :value="acc.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类" prop="category_id">
          <el-select v-model="transactionForm.category_id" placeholder="请选择分类" style="width: 100%;">
            <el-option
              v-for="cat in categories"
              :key="cat.id"
              :label="`${cat.icon} ${cat.name}`"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="金额" prop="amount">
          <el-input-number v-model="transactionForm.amount" :precision="2" :min="0.01" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="日期" prop="transaction_date">
          <el-date-picker
            v-model="transactionForm.transaction_date"
            type="date"
            value-format="YYYY-MM-DD"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="备注" prop="note">
          <el-input v-model="transactionForm.note" type="textarea" :rows="2" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveTransaction" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, type FormInstance, type FormRules } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import type { Transaction, Account, Category, TransactionType } from '@/types'
import { getTransactions, createTransaction, updateTransaction, deleteTransaction, batchDeleteTransactions, type CreateTransactionData } from '@/api/transaction'
import { getAccounts } from '@/api/account'
import { getCategories } from '@/api/category'

const transactions = ref<Transaction[]>([])
const accounts = ref<Account[]>([])
const categories = ref<Category[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingTransaction = ref<Transaction | null>(null)
const currentType = ref<TransactionType>('expense')
const selectedIds = ref<number[]>([])

const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filterType = ref<TransactionType | ''>('')
const dateRange = ref<string[]>([])

const transactionFormRef = ref<FormInstance>()

const transactionForm = reactive<CreateTransactionData & { category_id?: number }>({
  account_id: 0,
  category_id: undefined,
  type: 'expense',
  amount: 0,
  transaction_date: '',
  note: ''
})

const transactionRules: FormRules = {
  account_id: [{ required: true, message: '请选择账户', trigger: 'change' }],
  amount: [{ required: true, message: '请输入金额', trigger: 'blur' }],
  transaction_date: [{ required: true, message: '请选择日期', trigger: 'change' }]
}

const dialogTitle = computed(() => {
  if (editingTransaction.value) {
    return '编辑记录'
  }
  return currentType.value === 'expense' ? '记支出' : '记收入'
})

const formatNumber = (num: number) => {
  return num.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

const formatDate = (date: string) => {
  return dayjs(date).format('YYYY-MM-DD')
}

const loadTransactions = async () => {
  loading.value = true
  try {
    const params: any = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    
    if (filterType.value) {
      params.type = filterType.value
    }
    
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    
    const response = await getTransactions(params)
    transactions.value = response.data
    total.value = response.total
  } catch (error) {
    console.error('Failed to load transactions:', error)
  } finally {
    loading.value = false
  }
}

const loadAccounts = async () => {
  try {
    accounts.value = await getAccounts()
    if (accounts.value.length > 0) {
      transactionForm.account_id = accounts.value[0].id
    }
  } catch (error) {
    console.error('Failed to load accounts:', error)
  }
}

const loadCategories = async (type: CategoryType = 'expense') => {
  try {
    categories.value = await getCategories(type)
  } catch (error) {
    console.error('Failed to load categories:', error)
  }
}

const handleSelectionChange = (selection: Transaction[]) => {
  selectedIds.value = selection.map(t => t.id)
}

const openCreateDialog = (type: TransactionType) => {
  currentType.value = type
  editingTransaction.value = null
  transactionForm.type = type
  transactionForm.account_id = accounts.value[0]?.id || 0
  transactionForm.category_id = undefined
  transactionForm.amount = 0
  transactionForm.transaction_date = dayjs().format('YYYY-MM-DD')
  transactionForm.note = ''
  loadCategories(type as CategoryType)
  dialogVisible.value = true
}

const openEditDialog = (transaction: Transaction) => {
  editingTransaction.value = transaction
  currentType.value = transaction.type
  transactionForm.type = transaction.type
  transactionForm.account_id = transaction.account_id
  transactionForm.category_id = transaction.category_id
  transactionForm.amount = transaction.amount
  transactionForm.transaction_date = formatDate(transaction.transaction_date)
  transactionForm.note = transaction.note
  loadCategories(transaction.type as CategoryType)
  dialogVisible.value = true
}

const handleSaveTransaction = async () => {
  if (!transactionFormRef.value) return
  
  try {
    await transactionFormRef.value.validate()
    saving.value = true
    
    const data: CreateTransactionData = {
      account_id: transactionForm.account_id,
      type: transactionForm.type,
      amount: transactionForm.amount,
      transaction_date: transactionForm.transaction_date,
      note: transactionForm.note
    }
    
    if (transactionForm.category_id) {
      data.category_id = transactionForm.category_id
    }
    
    if (editingTransaction.value) {
      await updateTransaction(editingTransaction.value.id, data)
      ElMessage.success('记录更新成功')
    } else {
      await createTransaction(data)
      ElMessage.success('记录创建成功')
    }
    
    dialogVisible.value = false
    loadTransactions()
  } catch (error) {
    // Error handled by interceptor
  } finally {
    saving.value = false
  }
}

const handleDelete = async (transaction: Transaction) => {
  try {
    await ElMessageBox.confirm('确定要删除这条记录吗？', '提示', { type: 'warning' })
    
    await deleteTransaction(transaction.id)
    ElMessage.success('记录删除成功')
    loadTransactions()
  } catch (error) {
    // User cancelled or error
  }
}

const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedIds.value.length} 条记录吗？`,
      '提示',
      { type: 'warning' }
    )
    
    await batchDeleteTransactions(selectedIds.value)
    ElMessage.success('批量删除成功')
    selectedIds.value = []
    loadTransactions()
  } catch (error) {
    // User cancelled or error
  }
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  loadTransactions()
}

const handleCurrentChange = (page: number) => {
  currentPage.value = page
  loadTransactions()
}

onMounted(() => {
  loadAccounts()
  loadCategories()
  loadTransactions()
})
</script>
