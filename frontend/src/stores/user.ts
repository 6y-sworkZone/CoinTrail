import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types'
import { login as apiLogin, register as apiRegister, getProfile } from '@/api/auth'
import type { LoginData, RegisterData } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUser = (newUser: User) => {
    user.value = newUser
    localStorage.setItem('user', JSON.stringify(newUser))
  }

  const login = async (data: LoginData) => {
    const response = await apiLogin(data)
    setToken(response.token)
    setUser(response.user)
    return response
  }

  const register = async (data: RegisterData) => {
    const response = await apiRegister(data)
    setToken(response.token)
    setUser(response.user)
    return response
  }

  const logout = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const fetchProfile = async () => {
    try {
      const profile = await getProfile()
      user.value = profile
      return profile
    } catch (error) {
      logout()
      throw error
    }
  }

  const initFromStorage = () => {
    const storedUser = localStorage.getItem('user')
    if (storedUser) {
      try {
        user.value = JSON.parse(storedUser)
      } catch {
        // ignore
      }
    }
  }

  return {
    token,
    user,
    isLoggedIn,
    login,
    register,
    logout,
    fetchProfile,
    initFromStorage
  }
})
