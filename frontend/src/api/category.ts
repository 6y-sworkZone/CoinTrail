import request from '@/utils/request'
import type { Category, CategoryType } from '@/types'

export interface CreateCategoryData {
  name: string
  type: CategoryType
  icon?: string
  color?: string
  sort?: number
}

export interface UpdateCategoryData {
  name?: string
  type?: CategoryType
  icon?: string
  color?: string
  sort?: number
}

export const getCategories = (type?: CategoryType): Promise<Category[]> => {
  const params = type ? { type } : {}
  return request.get('/categories', { params })
}

export const getCategory = (id: number): Promise<Category> => {
  return request.get(`/categories/${id}`)
}

export const createCategory = (data: CreateCategoryData): Promise<Category> => {
  return request.post('/categories', data)
}

export const updateCategory = (id: number, data: UpdateCategoryData): Promise<Category> => {
  return request.patch(`/categories/${id}`, data)
}

export const deleteCategory = (id: number): Promise<any> => {
  return request.delete(`/categories/${id}`)
}
