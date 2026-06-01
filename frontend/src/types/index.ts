export interface User {
  id: number
  username: string
  email: string
  created_at: string
  updated_at: string
}

export interface LoginResponse {
  token: string
  user: User
}

export type AccountType = 'cash' | 'bank' | 'wechat' | 'alipay' | 'credit' | 'invest' | 'other'

export interface Account {
  id: number
  user_id: number
  name: string
  type: AccountType
  balance: number
  currency: string
  icon: string
  note: string
  created_at: string
  updated_at: string
}

export type CategoryType = 'expense' | 'income'

export interface Category {
  id: number
  user_id: number
  name: string
  type: CategoryType
  icon: string
  color: string
  sort: number
  is_default: boolean
  created_at: string
  updated_at: string
}

export type TransactionType = 'expense' | 'income' | 'transfer'

export interface Transaction {
  id: number
  user_id: number
  account_id: number
  target_account_id?: number
  category_id?: number
  type: TransactionType
  amount: number
  note: string
  transaction_date: string
  created_at: string
  updated_at: string
  account_name?: string
  category_name?: string
  category_icon?: string
  category_color?: string
}

export interface TransactionListResponse {
  data: Transaction[]
  total: number
  page: number
  page_size: number
  total_page: number
}

export interface Budget {
  id: number
  user_id: number
  category_id: number
  amount: number
  month: string
  created_at: string
  updated_at: string
}

export interface BudgetWithUsage extends Budget {
  used_amount: number
  remaining: number
  percentage: number
  is_over_budget: boolean
  category_name: string
  category_icon: string
  category_color: string
}

export interface MonthlySummary {
  month: string
  income: number
  expense: number
  balance: number
}

export interface CategorySummary {
  category_id: number
  category_name: string
  category_icon: string
  category_color: string
  amount: number
  percentage: number
}

export interface DashboardStats {
  total_balance: number
  monthly_income: number
  monthly_expense: number
  monthly_balance: number
  today_income: number
  today_expense: number
}
