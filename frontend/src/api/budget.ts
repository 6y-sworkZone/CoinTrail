import request from '@/utils/request'
import type { Budget, BudgetWithUsage } from '@/types'

export interface CreateBudgetData {
  category_id: number
  amount: number
  month: string
}

export interface UpdateBudgetData {
  amount: number
}

export const getBudgets = (month?: string): Promise<BudgetWithUsage[]> => {
  const params = month ? { month } : {}
  return request.get('/budgets', { params })
}

export const getBudget = (id: number): Promise<BudgetWithUsage> => {
  return request.get(`/budgets/${id}`)
}

export const createBudget = (data: CreateBudgetData): Promise<Budget> => {
  return request.post('/budgets', data)
}

export const updateBudget = (id: number, data: UpdateBudgetData): Promise<Budget> => {
  return request.patch(`/budgets/${id}`, data)
}

export const deleteBudget = (id: number): Promise<any> => {
  return request.delete(`/budgets/${id}`)
}

export const getDashboardBudgets = (): Promise<BudgetWithUsage[]> => {
  return request.get('/stats/dashboard-budgets')
}
