<template>
  <div>
    <el-row :gutter="20">
      <el-col :span="12">
        <div class="card" style="margin-bottom: 20px;">
          <h3 style="margin-bottom: 20px;">修改密码</h3>
          <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef" label-width="100px">
            <el-form-item label="旧密码" prop="old_password">
              <el-input v-model="passwordForm.old_password" type="password" show-password style="width: 100%;" />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input v-model="passwordForm.new_password" type="password" show-password style="width: 100%;" />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm_password">
              <el-input v-model="passwordForm.confirm_password" type="password" show-password style="width: 100%;" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleChangePassword" :loading="changingPassword">
                确认修改
              </el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-col>
      
      <el-col :span="12">
        <div class="card" style="margin-bottom: 20px;">
          <h3 style="margin-bottom: 20px;">数据导出</h3>
          <p style="color: #606266; margin-bottom: 16px;">导出指定日期范围的记账记录为 CSV 文件</p>
          <el-form label-width="100px">
            <el-form-item label="开始日期">
              <el-date-picker
                v-model="exportStartDate"
                type="date"
                value-format="YYYY-MM-DD"
                style="width: 100%;"
              />
            </el-form-item>
            <el-form-item label="结束日期">
              <el-date-picker
                v-model="exportEndDate"
                type="date"
                value-format="YYYY-MM-DD"
                style="width: 100%;"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleExport" :loading="exporting">
                <el-icon><Download /></el-icon>
                导出 CSV
              </el-button>
            </el-form-item>
          </el-form>
        </div>

        <div class="card">
          <h3 style="margin-bottom: 20px;">数据导入</h3>
          <p style="color: #606266; margin-bottom: 16px;">从 CSV 文件导入记账记录</p>
          <el-upload
            :show-file-list="false"
            :before-upload="beforeUpload"
            :http-request="handleImport"
            accept=".csv"
          >
            <el-button type="success" :loading="importing">
              <el-icon><Upload /></el-icon>
              选择文件导入
            </el-button>
          </el-upload>
          <div style="margin-top: 12px; font-size: 12px; color: #909399;">
            <p>CSV 文件格式要求：</p>
            <p>日期, 类型(收入/支出), 分类, 账户, 金额, 备注</p>
          </div>
          
          <div v-if="importResult" style="margin-top: 16px; padding: 12px; background: #f5f7fa; border-radius: 4px;">
            <p><strong>导入结果：</strong></p>
            <p>成功：{{ importResult.success }} 条</p>
            <p>跳过：{{ importResult.skipped }} 条</p>
            <p>失败：{{ importResult.errors }} 条</p>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, type FormInstance, type FormRules, type UploadRequestOptions } from 'vue'
import { ElMessage } from 'element-plus'
import { changePassword, type ChangePasswordData } from '@/api/auth'
import { exportTransactions, importTransactions } from '@/api/importExport'

const passwordFormRef = ref<FormInstance>()
const changingPassword = ref(false)
const exporting = ref(false)
const importing = ref(false)

const exportStartDate = ref('')
const exportEndDate = ref('')
const importResult = ref<{ success: number; skipped: number; errors: number } | null>(null)

const passwordForm = reactive<ChangePasswordData & { confirm_password: string }>({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const validateConfirmPassword = (rule: any, value: string, callback: any) => {
  if (value !== passwordForm.new_password) {
    callback(new Error('两次输入的新密码不一致'))
  } else {
    callback()
  }
}

const passwordRules: FormRules = {
  old_password: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleChangePassword = async () => {
  if (!passwordFormRef.value) return
  
  try {
    await passwordFormRef.value.validate()
    changingPassword.value = true
    
    await changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })
    
    ElMessage.success('密码修改成功')
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
  } catch (error) {
    // Error handled by interceptor
  } finally {
    changingPassword.value = false
  }
}

const handleExport = async () => {
  try {
    exporting.value = true
    
    const blob = await exportTransactions(exportStartDate.value, exportEndDate.value) as unknown as Blob
    
    const url = window.URL.createObjectURL(new Blob([blob]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `transactions_${new Date().getTime()}.csv`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

const beforeUpload = (file: File) => {
  const isCSV = file.name.endsWith('.csv')
  if (!isCSV) {
    ElMessage.error('只能上传 CSV 文件！')
  }
  return isCSV
}

const handleImport = async (options: UploadRequestOptions) => {
  try {
    importing.value = true
    importResult.value = null
    
    const result = await importTransactions(options.file as File)
    importResult.value = {
      success: result.success,
      skipped: result.skipped,
      errors: result.errors
    }
    
    ElMessage.success('导入完成')
  } catch (error) {
    ElMessage.error('导入失败')
  } finally {
    importing.value = false
  }
}
</script>
