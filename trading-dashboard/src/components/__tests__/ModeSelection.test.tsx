import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { ModeSelection } from '../ModeSelection';
import { TradingMode } from '../../types/agent';
import type { RenderResult } from '@testing-library/react';

describe('ModeSelection', () => {
  const defaultProps = {
    selectedMode: TradingMode.DEX,
    onModeSelect: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders both trading mode options', () => {
    render(<ModeSelection {...defaultProps} />);
    
    expect(screen.getByText('DEX Trading')).toBeInTheDocument();
    expect(screen.getByText('Pump.fun Trading')).toBeInTheDocument();
  });

  it('highlights selected mode', () => {
    render(<ModeSelection {...defaultProps} />);
    
    const dexButton = screen.getByText('DEX Trading').closest('button');
    const pumpButton = screen.getByText('Pump.fun Trading').closest('button');
    
    expect(dexButton).toHaveClass('border-primary');
    expect(pumpButton).not.toHaveClass('border-primary');
  });

  it('calls onModeSelect when mode is changed', () => {
    render(<ModeSelection {...defaultProps} />);
    
    fireEvent.click(screen.getByText('Pump.fun Trading'));
    expect(defaultProps.onModeSelect).toHaveBeenCalledWith(TradingMode.PUMPFUN);
  });

  it('disables buttons when disabled prop is true', () => {
    render(<ModeSelection {...defaultProps} disabled />);
    
    const buttons = screen.getAllByRole('button');
    buttons.forEach(button => {
      expect(button).toBeDisabled();
      expect(button).toHaveClass('opacity-50');
      expect(button).toHaveClass('cursor-not-allowed');
    });
  });
});
