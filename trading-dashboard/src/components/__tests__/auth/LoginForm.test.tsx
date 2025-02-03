import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { LoginForm } from '../../auth/LoginForm'
import { useAuth } from '../../../hooks/auth/useAuth'

jest.mock('../../../hooks/auth/useAuth')

describe('LoginForm', () => {
  const mockLogin = jest.fn()
  
  beforeEach(() => {
    jest.clearAllMocks()
    ;(useAuth as jest.Mock).mockReturnValue({
      login: mockLogin,
      isLoading: false,
      error: null
    })
  })

  it('renders login form correctly', () => {
    render(<LoginForm />)
    
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /login/i })).toBeInTheDocument()
  })

  it('handles form submission correctly', async () => {
    render(<LoginForm />)
    
    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /login/i })

    fireEvent.change(usernameInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        username: 'testuser',
        password: 'password123'
      })
    })
  })

  it('displays loading state during submission', () => {
    ;(useAuth as jest.Mock).mockReturnValue({
      login: mockLogin,
      isLoading: true,
      error: null
    })

    render(<LoginForm />)
    
    expect(screen.getByRole('button', { name: /logging in/i })).toBeDisabled()
  })

  it('displays error message when login fails', () => {
    const errorMessage = 'Invalid credentials'
    ;(useAuth as jest.Mock).mockReturnValue({
      login: mockLogin,
      isLoading: false,
      error: errorMessage
    })

    render(<LoginForm />)
    
    expect(screen.getByRole('alert')).toHaveTextContent(errorMessage)
  })

  it('validates required fields', async () => {
    render(<LoginForm />)
    
    fireEvent.click(screen.getByRole('button', { name: /login/i }))
    
    await waitFor(() => {
      expect(screen.getByText(/username is required/i)).toBeInTheDocument()
      expect(screen.getByText(/password is required/i)).toBeInTheDocument()
    })
  })
})
