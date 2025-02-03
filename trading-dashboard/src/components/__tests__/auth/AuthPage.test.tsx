import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { AuthPage } from '../../auth/AuthPage'
import { useAuth } from '../../../hooks/auth/useAuth'

jest.mock('../../../hooks/auth/useAuth')

describe('AuthPage', () => {
  const mockAuth = {
    isAuthenticated: false,
    isLoading: false,
    error: null,
    login: jest.fn(),
    register: jest.fn(),
    logout: jest.fn()
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useAuth as jest.Mock).mockReturnValue(mockAuth)
  })

  it('renders login form by default', () => {
    render(<AuthPage />)
    
    expect(screen.getByRole('heading', { name: /login/i })).toBeInTheDocument()
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /login/i })).toBeInTheDocument()
  })

  it('switches to registration form', () => {
    render(<AuthPage />)
    
    const registerLink = screen.getByRole('button', { name: /create account/i })
    fireEvent.click(registerLink)

    expect(screen.getByRole('heading', { name: /register/i })).toBeInTheDocument()
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument()
  })

  it('handles successful login', async () => {
    const mockLogin = jest.fn().mockResolvedValue(undefined)
    ;(useAuth as jest.Mock).mockReturnValue({
      ...mockAuth,
      login: mockLogin
    })

    render(<AuthPage />)
    
    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/password/i)
    
    fireEvent.change(usernameInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })
    
    const loginButton = screen.getByRole('button', { name: /login/i })
    fireEvent.click(loginButton)

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        username: 'testuser',
        password: 'password123'
      })
    })
  })

  it('handles successful registration', async () => {
    const mockRegister = jest.fn().mockResolvedValue(undefined)
    ;(useAuth as jest.Mock).mockReturnValue({
      ...mockAuth,
      register: mockRegister
    })

    render(<AuthPage />)
    
    const registerLink = screen.getByRole('button', { name: /create account/i })
    fireEvent.click(registerLink)

    const emailInput = screen.getByLabelText(/email/i)
    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } })
    fireEvent.change(usernameInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })
    fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } })
    
    const registerButton = screen.getByRole('button', { name: /register/i })
    fireEvent.click(registerButton)

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        email: 'test@example.com',
        username: 'testuser',
        password: 'password123'
      })
    })
  })

  it('displays loading state', () => {
    ;(useAuth as jest.Mock).mockReturnValue({
      ...mockAuth,
      isLoading: true
    })

    render(<AuthPage />)
    
    expect(screen.getByRole('button', { name: /loading/i })).toBeDisabled()
  })

  it('displays error messages', () => {
    ;(useAuth as jest.Mock).mockReturnValue({
      ...mockAuth,
      error: 'Invalid credentials'
    })

    render(<AuthPage />)
    
    expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument()
  })

  it('validates password match in registration', async () => {
    render(<AuthPage />)
    
    const registerLink = screen.getByRole('button', { name: /create account/i })
    fireEvent.click(registerLink)

    const passwordInput = screen.getByLabelText(/password/i)
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    
    fireEvent.change(passwordInput, { target: { value: 'password123' } })
    fireEvent.change(confirmPasswordInput, { target: { value: 'password456' } })
    
    const registerButton = screen.getByRole('button', { name: /register/i })
    fireEvent.click(registerButton)

    expect(screen.getByText(/passwords do not match/i)).toBeInTheDocument()
    expect(mockAuth.register).not.toHaveBeenCalled()
  })
})
