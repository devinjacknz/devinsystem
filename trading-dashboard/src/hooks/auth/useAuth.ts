import { useState, useCallback } from 'react'
import { AuthCredentials, AuthResponse, AuthState } from '../../types/auth'

const API_URL = process.env.VITE_API_URL || 'http://localhost:8080'

export function useAuth() {
  const [state, setState] = useState<AuthState>({
    isAuthenticated: false,
    user: null,
    token: null,
    error: null,
    isLoading: false
  })

  const login = useCallback(async (credentials: AuthCredentials) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }))
    try {
      const response = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(credentials)
      })
      
      if (!response.ok) {
        throw new Error('Login failed')
      }

      const data: AuthResponse = await response.json()
      setState({
        isAuthenticated: true,
        user: data.user,
        token: data.token,
        error: null,
        isLoading: false
      })
      localStorage.setItem('auth_token', data.token)
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
      
      if (!response.ok) {
        throw new Error('Registration failed')
      }

      const data: AuthResponse = await response.json()
      setState({
        isAuthenticated: true,
        user: data.user,
        token: data.token,
        error: null,
        isLoading: false
      })
      localStorage.setItem('auth_token', data.token)
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
    setState({
      isAuthenticated: false,
      user: null,
      token: null,
      error: null,
      isLoading: false
    })
  }, [])

  return {
    ...state,
    login,
    logout,
    register
  }
}
