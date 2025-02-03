import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { WalletManager } from '../../wallet/WalletManager'
import { useWallet } from '../../../hooks/wallet/useWallet'
import { WalletType } from '../../../types/wallet'

jest.mock('../../../hooks/wallet/useWallet')

describe('WalletManager', () => {
  const defaultProps = {
    tradingWallet: {
      address: 'trading-wallet-address',
      balance: 1000,
      type: WalletType.Trading
    },
    profitWallet: {
      address: 'profit-wallet-address',
      balance: 500,
      type: WalletType.Profit
    },
    onTransfer: jest.fn(),
    onConnect: jest.fn(),
    isConnecting: false,
    error: null
  }

  const mockWallet = {
    tradingWallet: defaultProps.tradingWallet,
    profitWallet: defaultProps.profitWallet,
    isConnected: true,
    error: null,
    transfer: jest.fn()
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useWallet as jest.Mock).mockReturnValue(mockWallet)
  })

  it('renders wallet manager interface', () => {
    render(<WalletManager {...defaultProps} />)
    
    expect(screen.getByText(/trading wallet/i)).toBeInTheDocument()
    expect(screen.getByText(/profit wallet/i)).toBeInTheDocument()
    expect(screen.getByText(/balance: 1000/i)).toBeInTheDocument()
    expect(screen.getByText(/balance: 500/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /transfer/i })).toBeInTheDocument()
  })

  it('handles wallet transfer', async () => {
    const mockTransfer = jest.fn().mockResolvedValue({ hash: 'tx-hash' })
    ;(useWallet as jest.Mock).mockReturnValue({
      ...mockWallet,
      transfer: mockTransfer
    })

    render(<WalletManager {...defaultProps} />)
    
    const fromSelect = screen.getByLabelText(/from wallet/i)
    const toSelect = screen.getByLabelText(/to wallet/i)
    const amountInput = screen.getByLabelText(/amount/i)
    const transferButton = screen.getByRole('button', { name: /transfer/i })

    fireEvent.change(fromSelect, { target: { value: WalletType.Trading } })
    fireEvent.change(toSelect, { target: { value: WalletType.Profit } })
    fireEvent.change(amountInput, { target: { value: '100' } })
    fireEvent.click(transferButton)

    expect(screen.getByRole('button', { name: /transferring/i })).toBeDisabled()

    await waitFor(() => {
      expect(mockTransfer).toHaveBeenCalledWith(
        WalletType.Trading,
        WalletType.Profit,
        100
      )
      expect(screen.getByText(/transaction successful/i)).toBeInTheDocument()
      expect(screen.getByText(/tx-hash/i)).toBeInTheDocument()
    })
  })

  it('shows insufficient balance error', () => {
    render(<WalletManager {...defaultProps} />)
    
    const fromSelect = screen.getByLabelText(/from wallet/i)
    const amountInput = screen.getByLabelText(/amount/i)

    fireEvent.change(fromSelect, { target: { value: WalletType.Trading } })
    fireEvent.change(amountInput, { target: { value: '2000' } })

    expect(screen.getByText(/insufficient balance/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /transfer/i })).toBeDisabled()
  })

  it('handles transfer error', async () => {
    const mockTransfer = jest.fn().mockRejectedValue(new Error('Transfer failed'))
    ;(useWallet as jest.Mock).mockReturnValue({
      ...mockWallet,
      transfer: mockTransfer
    })

    render(<WalletManager {...defaultProps} />)
    
    const fromSelect = screen.getByLabelText(/from wallet/i)
    const toSelect = screen.getByLabelText(/to wallet/i)
    const amountInput = screen.getByLabelText(/amount/i)
    const transferButton = screen.getByRole('button', { name: /transfer/i })

    fireEvent.change(fromSelect, { target: { value: WalletType.Trading } })
    fireEvent.change(toSelect, { target: { value: WalletType.Profit } })
    fireEvent.change(amountInput, { target: { value: '100' } })
    fireEvent.click(transferButton)

    await waitFor(() => {
      expect(screen.getByText(/transfer failed/i)).toBeInTheDocument()
      expect(transferButton).not.toBeDisabled()
    })
  })

  it('validates transfer inputs', () => {
    render(<WalletManager {...defaultProps} />)
    
    const transferButton = screen.getByRole('button', { name: /transfer/i })
    fireEvent.click(transferButton)

    expect(screen.getByText(/amount is required/i)).toBeInTheDocument()
    expect(screen.getByText(/select source wallet/i)).toBeInTheDocument()
    expect(screen.getByText(/select destination wallet/i)).toBeInTheDocument()
  })

  it('prevents same wallet transfer', () => {
    render(<WalletManager {...defaultProps} />)
    
    const fromSelect = screen.getByLabelText(/from wallet/i)
    const toSelect = screen.getByLabelText(/to wallet/i)
    
    fireEvent.change(fromSelect, { target: { value: WalletType.Trading } })
    fireEvent.change(toSelect, { target: { value: WalletType.Trading } })

    expect(screen.getByText(/cannot transfer to same wallet/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /transfer/i })).toBeDisabled()
  })
})
