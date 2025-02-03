import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { TradingDashboard } from '../TradingDashboard';
import { useAuth } from '../../hooks/auth/useAuth';
import { TradingMode } from '../../types/agent';
import '@testing-library/jest-dom';

jest.mock('../../hooks/auth/useAuth');
jest.mock('../ModeSelection', () => ({
  ModeSelection: ({ selectedMode, onModeSelect }: any) => (
    <div data-testid="mode-selection">
      <button onClick={() => onModeSelect(TradingMode.DEX)}>DEX</button>
      <button onClick={() => onModeSelect(TradingMode.PUMPFUN)}>PUMP</button>
      <span>Selected: {selectedMode}</span>
    </div>
  ),
}));

jest.mock('../agent/AgentDashboard', () => ({
  AgentDashboard: ({ mode }: any) => (
    <div data-testid="agent-dashboard">
      Mode: {mode}
    </div>
  ),
}));

describe('TradingDashboard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('shows loading state when authenticating', () => {
    (useAuth as jest.Mock).mockReturnValue({
      isLoading: true,
      isAuthenticated: false,
    });

    render(<TradingDashboard />);
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('shows authentication required message when not authenticated', () => {
    (useAuth as jest.Mock).mockReturnValue({
      isLoading: false,
      isAuthenticated: false,
    });

    render(<TradingDashboard />);
    expect(screen.getByText('Authentication Required')).toBeInTheDocument();
    expect(screen.getByText('Please log in to access the trading dashboard.')).toBeInTheDocument();
  });

  it('renders trading dashboard when authenticated', () => {
    (useAuth as jest.Mock).mockReturnValue({
      isLoading: false,
      isAuthenticated: true,
    });

    render(<TradingDashboard />);
    expect(screen.getByText('Trading Dashboard')).toBeInTheDocument();
    expect(screen.getByTestId('mode-selection')).toBeInTheDocument();
    expect(screen.getByTestId('agent-dashboard')).toBeInTheDocument();
  });

  it('updates mode when selection changes', () => {
    (useAuth as jest.Mock).mockReturnValue({
      isLoading: false,
      isAuthenticated: true,
    });

    render(<TradingDashboard />);
    
    // Initial mode should be DEX
    expect(screen.getByText('Mode: dex')).toBeInTheDocument();
    
    // Change mode to PUMP.fun
    fireEvent.click(screen.getByText('PUMP'));
    expect(screen.getByText('Mode: pumpfun')).toBeInTheDocument();
  });

  it('passes correct mode to AgentDashboard', () => {
    (useAuth as jest.Mock).mockReturnValue({
      isLoading: false,
      isAuthenticated: true,
    });

    render(<TradingDashboard />);
    
    // Initial mode
    expect(screen.getByText('Mode: dex')).toBeInTheDocument();
    
    // Change mode
    fireEvent.click(screen.getByText('PUMP'));
    expect(screen.getByText('Mode: pumpfun')).toBeInTheDocument();
  });
});
