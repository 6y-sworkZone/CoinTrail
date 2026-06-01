import request from '@/utils/request'
import type { Transaction, TransactionType, TransactionListResponse } from '@/types'

export interface CreateTransactionData {
  account_id: number
  category_id?: number
  type: TransactionType
  amount: number
  note?: string
  transaction_date: string
}

export interface UpdateTransactionData {
  account_id?: number
  category_id?: number
  type?: TransactionType
  amount?: number
  note?: string
  transaction_date?: string
}

export interface TransactionQueryParams {
  page?: number
  page_size?: number
  type?: TransactionType
  category_id?: number
  account_id?: number
  start_date?: string
  end_date?: string
}

export const getTransactions = (params?: TransactionQueryParams): Promise<TransactionListResponse> => {
  return request.get('/transactions', { params })
}

export const getTransaction = (id: number): Promise<Transaction> => {
  return request.get(`/transactions/${id}`)
}

export const createTransaction = (data: CreateTransactionData): Promise<Transaction> => {
  return request.post('/transactions', data)
}

export const updateTransaction = (id: number, data: UpdateTransactionData): Promise<Transaction> => {
  return request.patch(`/transactions/${id}`, data)
}

export const deleteTransaction = (id: number): Promise<any> => {
  return request.delete(`/transactions/${id}`)
}

export const batchDeleteTransactions = (ids: number[]): Promise<any> => {
  return request.delete('/transactions/batch', { data: { ids } })
}
