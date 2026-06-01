<template>
  <div>
    <div style="margin-bottom: 16px; display: flex; gap: 12px; align-items: center;">
      <el-radio-group v-model="activeTab" @change="loadCategories">
        <el-radio-button label="expense">支出分类</el-radio-button>
        <el-radio-button label="income">收入分类</el-radio-button>
      </el-radio-group>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        新建分类
      </el-button>
    </div>

    <el-row :gutter="16">
      <el-col :span="12" v-for="category in categories" :key="category.id">
        <div class="category-item">
          <div class="category-icon" :style="{ background: category.color + '20', color: category.color }">
            {{ category.icon }}
          </div>
          <span class="category-name">{{ category.name }}</span>
          <div style="display: flex; gap: 8px;">
            <el-button v-if="!category.is_default" size="small" @click="openEditDialog(category)">编辑</el-button>
            <el-button
              v-if="!category.is_default"
              size="small"
              type="danger"
              @click="handleDelete(category)"
            >
              删除
            </el-button>
            <el-tag v-if="category.is_default" type="info" size="small">默认</el-tag>
          </div>
        </div>
      </el-col>
    </el-row>

    <el-dialog v-model="dialogVisible" :title="editingCategory ? '编辑分类' : '新建分类'" width="500px">
      <el-form :model="categoryForm" :rules="categoryRules" ref="categoryFormRef" label-width="80px">
        <el-form-item label="分类名称" prop="name">
          <el-input v-model="categoryForm.name" placeholder="请输入分类名称" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="categoryForm.icon" placeholder="请输入图标emoji" maxlength="2" style="width: 100px;" />
          <span style="margin-left: 12px; color: #909399;">输入emoji图标，如：🍜 🚗 🛒</span>
        </el-form-item>
        <el-form-item label="颜色" prop="color">
          <el-color-picker v-model="categoryForm.color" show-alpha />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="categoryForm.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveCategory" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import type { Category, CategoryType } from '@/types'
import { getCategories, createCategory, updateCategory, deleteCategory, type CreateCategoryData } from '@/api/category'

const categories = ref<Category[]>([])
const activeTab = ref<CategoryType>('expense')
const dialogVisible = ref(false)
const editingCategory = ref<Category | null>(null)
const saving = ref(false)

const categoryFormRef = ref<FormInstance>()

const categoryForm = reactive<CreateCategoryData>({
  name: '',
  type: 'expense',
  icon: '📝',
  color: '#409eff',
  sort: 0
})

const categoryRules: FormRules = {
  name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
  icon: [{ required: true, message: '请输入图标', trigger: 'blur' }],
  color: [{ required: true, message: '请选择颜色', trigger: 'change' }]
}

const loadCategories = async () => {
  try {
    categories.value = await getCategories(activeTab.value)
  } catch (error) {
    console.error('Failed to load categories:', error)
  }
}

const openCreateDialog = () => {
  editingCategory.value = null
  categoryForm.name = ''
  categoryForm.type = activeTab.value
  categoryForm.icon = '📝'
  categoryForm.color = '#409eff'
  categoryForm.sort = 0
  dialogVisible.value = true
}

const openEditDialog = (category: Category) => {
  editingCategory.value = category
  categoryForm.name = category.name
  categoryForm.type = category.type
  categoryForm.icon = category.icon
  categoryForm.color = category.color
  categoryForm.sort = category.sort
  dialogVisible.value = true
}

const handleSaveCategory = async () => {
  if (!categoryFormRef.value) return
  
  try {
    await categoryFormRef.value.validate()
    saving.value = true
    
    if (editingCategory.value) {
      await updateCategory(editingCategory.value.id, categoryForm)
      ElMessage.success('分类更新成功')
    } else {
      await createCategory(categoryForm)
      ElMessage.success('分类创建成功')
    }
    
    dialogVisible.value = false
    loadCategories()
  } catch (error) {
    // Error handled by interceptor
  } finally {
    saving.value = false
  }
}

const handleDelete = async (category: Category) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除分类"${category.name}"吗？`,
      '提示',
      { type: 'warning' }
    )
    
    await deleteCategory(category.id)
    ElMessage.success('分类删除成功')
    loadCategories()
  } catch (error) {
    // User cancelled or error
  }
}

onMounted(() => {
  loadCategories()
})
</script>
