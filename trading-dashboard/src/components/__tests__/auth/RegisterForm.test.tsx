import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { RegisterForm } from '../../auth/RegisterForm'
import { useAuth } from '../../../hooks/auth/useAuth'

jest.mock('../../../hooks/auth/useAuth')

describe('RegisterForm', () => {
  const mockRegister = jest.fn()
  
  beforeEach(() => {
    (useAuth as jest.Mock).mockReturnValue({
      register: mockRegister,
      isLoading: false,
      error: null
    })
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('renders register form correctly', () => {
    render(<RegisterForm />)
    
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /register/i })).toBeInTheDocument()
  })

  it('validates password confirmation', async () => {
    render(<RegisterForm />)
    
    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/^password$/i)
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    fireEvent.change(usernameInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })
    fireEvent.change(confirmPasswordInput, { target: { value: 'password124' } })
    fireEvent.click(submitButton)

    expect(screen.getByRole('alert')).toHaveTextContent(/passwords do not match/i)
    expect(mockRegister).not.toHaveBeenCalled()
  })

  it('handles form submission correctly', async () => {
    render(<RegisterForm />)
    
    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/^password$/i)
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    fireEvent.change(usernameInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })
    fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        username: 'testuser',
        password: 'password123'
      })
    })
  })

  it('displays loading state during submission', () => {
    (useAuth as jest.Mock).mockReturnValue({
      register: mockRegister,
      isLoading: true,
      error: null
    })

    render(<RegisterForm />)
    
    expect(screen.getByRole('button', { name: /creating account/i })).toBeDisabled()
  })

  it('displays error message when registration fails', () => {
    const errorMessage = 'Username already taken'
    ;(useAuth as jest.Mock).mockReturnValue({
      register: mockRegister,
      isLoading: false,
      error: errorMessage
    })

    render(<RegisterForm />)
    
    expect(screen.getByRole('alert')).toHaveTextContent(errorMessage)
  })
})
