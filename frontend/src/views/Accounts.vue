<template>
  <div>
    <div style="margin-bottom: 16px; display: flex; gap: 12px;">
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        新建账户
      </el-button>
      <el-button @click="openTransferDialog">
        <el-icon><Switch /></el-icon>
        账户转账
      </el-button>
    </div>

    <div v-if="accounts.length === 0" class="card" style="text-align: center; padding: 60px;">
      <el-empty description="暂无账户，点击上方按钮创建第一个账户吧" />
    </div>

    <div v-else>
      <div v-for="account in accounts" :key="account.id" class="account-item">
        <div class="account-icon" :style="{ background: getAccountTypeBg(account.type), color: getAccountTypeColor(account.type) }">
          {{ getAccountIcon(account.type) }}
        </div>
        <div class="account-info">
          <div class="account-name">{{ account.name }}</div>
          <div class="account-type">{{ getAccountTypeName(account.type) }}</div>
        </div>
        <div class="account-balance">¥{{ formatNumber(account.balance) }}</div>
        <div style="margin-left: 16px; display: flex; gap: 8px;">
          <el-button size="small" @click="openEditDialog(account)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(account)">删除</el-button>
        </div>
      </div>
    </div>

    <el-dialog v-model="createDialogVisible" title="新建账户" width="500px">
      <el-form :model="accountForm" :rules="accountRules" ref="accountFormRef" label-width="80px">
        <el-form-item label="账户名称" prop="name">
          <el-input v-model="accountForm.name" placeholder="请输入账户名称" />
        </el-form-item>
        <el-form-item label="账户类型" prop="type">
          <el-select v-model="accountForm.type" placeholder="请选择账户类型" style="width: 100%;">
            <el-option label="现金" value="cash" />
            <el-option label="银行卡" value="bank" />
            <el-option label="微信" value="wechat" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="信用卡" value="credit" />
            <el-option label="投资账户" value="invest" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="初始余额" prop="balance">
          <el-input-number v-model="accountForm.balance" :precision="2" :min="0" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="备注" prop="note">
          <el-input v-model="accountForm.note" type="textarea" :rows="3" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveAccount" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="transferDialogVisible" title="账户转账" width="500px">
      <el-form :model="transferForm" :rules="transferRules" ref="transferFormRef" label-width="80px">
        <el-form-item label="转出账户" prop="from_account_id">
          <el-select v-model="transferForm.from_account_id" placeholder="请选择转出账户" style="width: 100%;">
            <el-option v-for="acc in accounts" :key="acc.id" :label="acc.name" :value="acc.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="转入账户" prop="to_account_id">
          <el-select v-model="transferForm.to_account_id" placeholder="请选择转入账户" style="width: 100%;">
            <el-option v-for="acc in accounts" :key="acc.id" :label="acc.name" :value="acc.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="转账金额" prop="amount">
          <el-input-number v-model="transferForm.amount" :precision="2" :min="0.01" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="备注" prop="note">
          <el-input v-model="transferForm.note" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="transferDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleTransfer" :loading="transferring">确认转账</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, type FormInstance, type FormRules } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { Account, AccountType } from '@/types'
import { getAccounts, createAccount, updateAccount, deleteAccount, transfer, type CreateAccountData, type TransferData } from '@/api/account'

const accounts = ref<Account[]>([])
const createDialogVisible = ref(false)
const transferDialogVisible = ref(false)
const editingAccount = ref<Account | null>(null)
const saving = ref(false)
const transferring = ref(false)

const accountFormRef = ref<FormInstance>()
const transferFormRef = ref<FormInstance>()

const accountForm = reactive<CreateAccountData>({
  name: '',
  type: 'cash' as AccountType,
  balance: 0,
  note: ''
})

const transferForm = reactive<TransferData>({
  from_account_id: 0,
  to_account_id: 0,
  amount: 0,
  note: ''
})

const accountRules: FormRules = {
  name: [{ required: true, message: '请输入账户名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择账户类型', trigger: 'change' }]
}

const transferRules: FormRules = {
  from_account_id: [{ required: true, message: '请选择转出账户', trigger: 'change' }],
  to_account_id: [{ required: true, message: '请选择转入账户', trigger: 'change' }],
  amount: [{ required: true, message: '请输入转账金额', trigger: 'blur' }]
}

const formatNumber = (num: number) => {
  return num.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

const loadAccounts = async () => {
  try {
    accounts.value = await getAccounts()
  } catch (error) {
    console.error('Failed to load accounts:', error)
  }
}

const getAccountIcon = (type: AccountType) => {
  const icons: Record<AccountType, string> = {
    cash: '💵',
    bank: '🏦',
    wechat: '💚',
    alipay: '💙',
    credit: '💳',
    invest: '📈',
    other: '💰'
  }
  return icons[type] || '💰'
}

const getAccountTypeName = (type: AccountType) => {
  const names: Record<AccountType, string> = {
    cash: '现金',
    bank: '银行卡',
    wechat: '微信',
    alipay: '支付宝',
    credit: '信用卡',
    invest: '投资账户',
    other: '其他'
  }
  return names[type] || '其他'
}

const getAccountTypeBg = (type: AccountType) => {
  const colors: Record<AccountType, string> = {
    cash: 'rgba(103, 194, 58, 0.1)',
    bank: 'rgba(64, 158, 255, 0.1)',
    wechat: 'rgba(7, 193, 96, 0.1)',
    alipay: 'rgba(22, 119, 255, 0.1)',
    credit: 'rgba(155, 89, 182, 0.1)',
    invest: 'rgba(230, 126, 34, 0.1)',
    other: 'rgba(144, 147, 153, 0.1)'
  }
  return colors[type] || 'rgba(144, 147, 153, 0.1)'
}

const getAccountTypeColor = (type: AccountType) => {
  const colors: Record<AccountType, string> = {
    cash: '#67c23a',
    bank: '#409eff',
    wechat: '#07c160',
    alipay: '#1677ff',
    credit: '#9b59b6',
    invest: '#e67e22',
    other: '#909399'
  }
  return colors[type] || '#909399'
}

const openCreateDialog = () => {
  editingAccount.value = null
  accountForm.name = ''
  accountForm.type = 'cash'
  accountForm.balance = 0
  accountForm.note = ''
  createDialogVisible.value = true
}

const openEditDialog = (account: Account) => {
  editingAccount.value = account
  accountForm.name = account.name
  accountForm.type = account.type
  accountForm.balance = account.balance
  accountForm.note = account.note
  createDialogVisible.value = true
}

const openTransferDialog = () => {
  transferForm.from_account_id = accounts.value[0]?.id || 0
  transferForm.to_account_id = accounts.value[1]?.id || 0
  transferForm.amount = 0
  transferForm.note = ''
  transferDialogVisible.value = true
}

const handleSaveAccount = async () => {
  if (!accountFormRef.value) return
  
  try {
    await accountFormRef.value.validate()
    saving.value = true
    
    if (editingAccount.value) {
      await updateAccount(editingAccount.value.id, accountForm)
      ElMessage.success('账户更新成功')
    } else {
      await createAccount(accountForm)
      ElMessage.success('账户创建成功')
    }
    
    createDialogVisible.value = false
    loadAccounts()
  } catch (error) {
    // Error handled by interceptor
  } finally {
    saving.value = false
  }
}

const handleDelete = async (account: Account) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除账户"${account.name}"吗？相关的交易记录也会被删除。`,
      '提示',
      { type: 'warning' }
    )
    
    await deleteAccount(account.id)
    ElMessage.success('账户删除成功')
    loadAccounts()
  } catch (error) {
    // User cancelled or error
  }
}

const handleTransfer = async () => {
  if (!transferFormRef.value) return
  
  try {
    await transferFormRef.value.validate()
    
    if (transferForm.from_account_id === transferForm.to_account_id) {
      ElMessage.error('转出账户和转入账户不能相同')
      return
    }
    
    transferring.value = true
    await transfer(transferForm)
    ElMessage.success('转账成功')
    transferDialogVisible.value = false
    loadAccounts()
  } catch (error) {
    // Error handled by interceptor
  } finally {
    transferring.value = false
  }
}

onMounted(() => {
  loadAccounts()
})
</script>
