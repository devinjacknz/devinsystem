import { useState, useCallback } from 'react'
import { AuthCredentials, AuthState } from '../../types/auth'

const API_URL = process.env.VITE_API_URL || 'http://localhost:8080'

export function useAuth() {
  const [state, setState] = useState<AuthState>(() => ({
    isAuthenticated: !!localStorage.getItem('auth_token'),
    user: null,
    token: localStorage.getItem('auth_token'),
    error: null,
    isLoading: false
  }))

  const login = useCallback(async (credentials: AuthCredentials) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }))
    try {
      const response = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(credentials)
      })
      
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || 'Invalid credentials')
      }

      const token = data.token
      localStorage.setItem('auth_token', token)
      setState(prev => ({
        ...prev,
        isAuthenticated: true,
        user: data.user,
        token: token,
        error: null,
        isLoading: false
      }))
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Login failed',
        isLoading: false
      }))
    }
  }, [])

  const register = useCallback(async (credentials: AuthCredentials) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }))
    try {
      const response = await fetch(`${API_URL}/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(credentials)
      })
      
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || 'Registration failed')
      }

      const token = data.token
      localStorage.setItem('auth_token', token)
      setState(prev => ({
        ...prev,
        isAuthenticated: true,
        user: data.user,
        token: token,
        error: null,
        isLoading: false
      }))
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Registration failed',
        isLoading: false
      }))
    }
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('auth_token')
    setState(prev => ({
      ...prev,
      isAuthenticated: false,
      user: null,
      token: null,
      error: null,
      isLoading: false
    }))
  }, [])

  return {
    ...state,
    login,
    logout,
    register
  }
}
