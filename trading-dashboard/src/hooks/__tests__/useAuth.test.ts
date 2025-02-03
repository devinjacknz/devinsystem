import { renderHook, act } from '@testing-library/react'
import { useAuth } from '../auth/useAuth'

describe('useAuth', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    global.fetch = jest.fn()
    localStorage.clear()
  })

  it('initializes with correct default state', () => {
    const { result } = renderHook(() => useAuth())

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('handles successful login', async () => {
    const mockToken = 'mock-jwt-token'
    ;(global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ token: mockToken })
    })

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.login({
        username: 'testuser',
        password: 'password123'
      })
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.error).toBeNull()
    expect(localStorage.getItem('auth_token')).toBe(mockToken)
  })

  it('handles failed login', async () => {
    const errorMessage = 'Invalid credentials'
    ;(global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      json: () => Promise.resolve({ error: errorMessage })
    })

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.login({
        username: 'testuser',
        password: 'wrongpassword'
      })
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.error).toBe(errorMessage)
    expect(localStorage.getItem('auth_token')).toBeNull()
  })

  it('handles successful registration', async () => {
    const mockToken = 'mock-jwt-token'
    ;(global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ token: mockToken })
    })

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.register({
        username: 'testuser',
        password: 'password123'
      })
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.error).toBeNull()
    expect(localStorage.getItem('auth_token')).toBe(mockToken)
  })

  it('handles logout', async () => {
    localStorage.setItem('auth_token', 'existing-token')
    const { result } = renderHook(() => useAuth())

    await act(async () => {
      result.current.logout()
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(localStorage.getItem('auth_token')).toBeNull()
  })

  it('restores authentication state from localStorage', () => {
    localStorage.setItem('auth_token', 'existing-token')
    const { result } = renderHook(() => useAuth())

    expect(result.current.isAuthenticated).toBe(true)
  })

  it('handles network errors during login', async () => {
    const networkError = new Error('Network error')
    ;(global.fetch as jest.Mock).mockRejectedValueOnce(networkError)

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.login({
        username: 'testuser',
        password: 'password123'
      })
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.error).toBe('Network error')
    expect(localStorage.getItem('auth_token')).toBeNull()
  })

  it('handles token persistence', () => {
    const mockToken = 'mock-jwt-token'
    localStorage.setItem('auth_token', mockToken)
    
    const { result } = renderHook(() => useAuth())

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.error).toBeNull()
  })
})
