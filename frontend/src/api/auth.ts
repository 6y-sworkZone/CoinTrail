import request from '@/utils/request'
import type { User, LoginResponse } from '@/types'

export interface LoginData {
  email: string
  password: string
}

export interface RegisterData {
  username: string
  email: string
  password: string
}

export interface ChangePasswordData {
  old_password: string
  new_password: string
}

export const login = (data: LoginData): Promise<LoginResponse> => {
  return request.post('/auth/login', data)
}

export const register = (data: RegisterData): Promise<LoginResponse> => {
  return request.post('/auth/register', data)
}

export const changePassword = (data: ChangePasswordData): Promise<any> => {
  return request.patch('/auth/password', data)
}

export const getProfile = (): Promise<User> => {
  return request.get('/auth/profile')
}
