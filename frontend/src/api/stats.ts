import request from '@/utils/request'
import type { DashboardStats, MonthlySummary, CategorySummary } from '@/types'

export const getDashboardStats = (): Promise<DashboardStats> => {
  return request.get('/stats/dashboard')
}

export const getMonthlySummary = (startMonth?: string, endMonth?: string): Promise<MonthlySummary[]> => {
  const params: any = {}
  if (startMonth) params.start_month = startMonth
  if (endMonth) params.end_month = endMonth
  return request.get('/stats/monthly', { params })
}

export const getCategorySummary = (type?: string, month?: string): Promise<{ total: number; details: CategorySummary[] }> => {
  const params: any = {}
  if (type) params.type = type
  if (month) params.month = month
  return request.get('/stats/category', { params })
}

export const getTrendStats = (): Promise<MonthlySummary[]> => {
  return request.get('/stats/trend')
}
