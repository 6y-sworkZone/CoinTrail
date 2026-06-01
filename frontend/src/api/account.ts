import request from '@/utils/request'
import type { Account, AccountType } from '@/types'

export interface CreateAccountData {
  name: string
  type: AccountType
  balance?: number
  currency?: string
  icon?: string
  note?: string
}

export interface UpdateAccountData {
  name?: string
  type?: AccountType
  balance?: number
  currency?: string
  icon?: string
  note?: string
}

export interface TransferData {
  from_account_id: number
  to_account_id: number
  amount: number
  note?: string
}

export const getAccounts = (): Promise<Account[]> => {
  return request.get('/accounts')
}

export const getAccount = (id: number): Promise<Account> => {
  return request.get(`/accounts/${id}`)
}

export const createAccount = (data: CreateAccountData): Promise<Account> => {
  return request.post('/accounts', data)
}

export const updateAccount = (id: number, data: UpdateAccountData): Promise<Account> => {
  return request.patch(`/accounts/${id}`, data)
}

export const deleteAccount = (id: number): Promise<any> => {
  return request.delete(`/accounts/${id}`)
}

export const transfer = (data: TransferData): Promise<any> => {
  return request.post('/accounts/transfer', data)
}
