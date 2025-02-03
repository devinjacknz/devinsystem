import { render, screen, fireEvent, within } from '@testing-library/react';
import { AgentDashboard } from '../../agent/AgentDashboard';
import { useAgent } from '../../../hooks/agent/useAgent';
import { TradingMode } from '../../../types/agent';
import '@testing-library/jest-dom';

jest.mock('../../../hooks/agent/useAgent');

describe('AgentDashboard', () => {
  const mockAgents = [
    {
      id: '1',
      name: 'Test Agent 1',
      mode: TradingMode.DEX,
      status: {
        status: 'active',
        lastExecuted: '2024-02-01T12:00:00Z',
        performance: {
          totalTrades: 10,
          successRate: 80,
          pnl: 1000
        }
      }
    },
    {
      id: '2',
      name: 'Test Agent 2',
      mode: TradingMode.PUMPFUN,
      status: {
        status: 'paused',
        lastExecuted: '2024-02-01T11:00:00Z',
        performance: {
          totalTrades: 5,
          successRate: 60,
          pnl: 500
        }
      }
    }
  ];

  const mockUseAgent = {
    agents: mockAgents,
    isLoading: false,
    error: null,
    createAgent: jest.fn(),
    updateAgent: jest.fn(),
    toggleAgent: jest.fn()
  };

  beforeEach(() => {
    jest.clearAllMocks();
    (useAgent as jest.Mock).mockReturnValue(mockUseAgent);
  });

  it('renders agent dashboard with correct mode filtering', () => {
    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    expect(screen.getByText('Test Agent 1')).toBeInTheDocument();
    expect(screen.queryByText('Test Agent 2')).not.toBeInTheDocument();
  });

  it('handles agent creation', () => {
    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    fireEvent.click(screen.getByText('Create Agent'));
    expect(mockUseAgent.createAgent).toHaveBeenCalledWith({
      name: 'New Agent',
      mode: TradingMode.DEX
    });
  });

  it('handles agent selection and strategy editing', () => {
    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    const agentCard = screen.getByText('Test Agent 1').closest('div');
    fireEvent.click(agentCard!);
    
    const editButton = screen.getByText('Edit Strategy');
    fireEvent.click(editButton);
    
    expect(mockUseAgent.updateAgent).toHaveBeenCalledWith({
      id: '1'
    });
  });

  it('handles agent status toggle', () => {
    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    const agentCard = screen.getByText('Test Agent 1').closest('div');
    fireEvent.click(agentCard!);
    
    const toggleButton = screen.getByText('Stop');
    fireEvent.click(toggleButton);
    
    expect(mockUseAgent.toggleAgent).toHaveBeenCalledWith('1', true);
  });

  it('displays loading state', () => {
    (useAgent as jest.Mock).mockReturnValue({
      ...mockUseAgent,
      isLoading: true
    });
    
    render(<AgentDashboard mode={TradingMode.DEX} />);
    expect(screen.getByText('Loading agents...')).toBeInTheDocument();
  });

  it('displays error state', () => {
    const errorMessage = 'Failed to load agents';
    (useAgent as jest.Mock).mockReturnValue({
      ...mockUseAgent,
      error: errorMessage
    });
    
    render(<AgentDashboard mode={TradingMode.DEX} />);
    expect(screen.getByText(errorMessage)).toBeInTheDocument();
  });

  it('displays empty state', () => {
    (useAgent as jest.Mock).mockReturnValue({
      ...mockUseAgent,
      agents: []
    });
    
    render(<AgentDashboard mode={TradingMode.DEX} />);
    expect(screen.getByText(/no agents created/i)).toBeInTheDocument();
  });

  it('displays correct status indicators', () => {
    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    const activeStatus = screen.getByText('active');
    expect(activeStatus).toHaveClass('bg-success/10', 'text-success');
  });

  it('handles error states in agent operations', () => {
    const errorMessage = 'Failed to update agent';
    (useAgent as jest.Mock).mockReturnValue({
      ...mockUseAgent,
      updateAgent: jest.fn().mockRejectedValue(new Error(errorMessage))
    });

    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    const agentCard = screen.getByText('Test Agent 1').closest('div');
    fireEvent.click(agentCard!);
    
    const editButton = screen.getByText('Edit Strategy');
    fireEvent.click(editButton);
    
    expect(screen.getByText(errorMessage)).toBeInTheDocument();
  });

  it('disables controls when agent operation is in progress', () => {
    (useAgent as jest.Mock).mockReturnValue({
      ...mockUseAgent,
      isLoading: true
    });

    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    const createButton = screen.getByText('Create Agent');
    expect(createButton).toBeDisabled();
  });

  it('updates agent list after successful operations', async () => {
    const newAgent = {
      id: '3',
      name: 'New Test Agent',
      mode: TradingMode.DEX,
      createdAt: '2024-02-01T13:00:00Z',
      updatedAt: '2024-02-01T13:00:00Z',
      strategy: {
        id: '3',
        name: 'New Strategy',
        mode: TradingMode.DEX,
        parameters: {
          entryConditions: [],
          exitConditions: [],
          riskManagement: {
            stopLoss: 5,
            takeProfit: 10,
            maxPositionSize: 1000
          }
        }
      },
      status: {
        status: 'active',
        lastExecuted: '',
        performance: {
          totalTrades: 0,
          successRate: 0,
          pnl: 0
        }
      }
    };

    const mockCreateAgent = jest.fn().mockResolvedValue(newAgent);
    (useAgent as jest.Mock).mockReturnValue({
      ...mockUseAgent,
      createAgent: mockCreateAgent,
      agents: [...mockAgents, newAgent]
    });

    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    fireEvent.click(screen.getByText('Create Agent'));
    expect(await screen.findByText('New Test Agent')).toBeInTheDocument();
  });

  it('displays agent performance metrics correctly', () => {
    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    const agentCard = screen.getByText('Test Agent 1').closest('div');
    expect(within(agentCard!).getByText('Total Trades: 10')).toBeInTheDocument();
    expect(within(agentCard!).getByText('Success Rate: 80%')).toBeInTheDocument();
    expect(within(agentCard!).getByText('PnL: 1000')).toBeInTheDocument();
  });

  it('handles mode switching correctly', () => {
    const { rerender } = render(<AgentDashboard mode={TradingMode.DEX} />);
    expect(screen.getByText('Test Agent 1')).toBeInTheDocument();
    expect(screen.queryByText('Test Agent 2')).not.toBeInTheDocument();

    rerender(<AgentDashboard mode={TradingMode.PUMPFUN} />);
    expect(screen.queryByText('Test Agent 1')).not.toBeInTheDocument();
    expect(screen.getByText('Test Agent 2')).toBeInTheDocument();
  });

  it('handles agent status updates correctly', async () => {
    const updatedAgent = {
      ...mockAgents[0],
      status: {
        ...mockAgents[0].status,
        status: 'paused'
      }
    };

    const mockToggleAgent = jest.fn().mockResolvedValue(updatedAgent);
    (useAgent as jest.Mock).mockReturnValue({
      ...mockUseAgent,
      toggleAgent: mockToggleAgent,
      agents: [updatedAgent, mockAgents[1]]
    });

    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    const agentCard = screen.getByText('Test Agent 1').closest('div');
    fireEvent.click(agentCard!);
    
    const toggleButton = screen.getByText('Stop');
    fireEvent.click(toggleButton);
    
    expect(mockToggleAgent).toHaveBeenCalledWith('1', true);
    expect(await screen.findByText('paused')).toBeInTheDocument();
  });

  it('displays agent performance metrics with correct formatting', () => {
    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    const agentCard = screen.getByText('Test Agent 1').closest('div');
    const performanceSection = within(agentCard!).getByText('Total Trades: 10').closest('div');
    
    expect(within(performanceSection!).getByText('Total Trades: 10')).toBeInTheDocument();
    expect(within(performanceSection!).getByText('Success Rate: 80%')).toBeInTheDocument();
    expect(within(performanceSection!).getByText('PnL: 1000')).toBeInTheDocument();
  });

  it('handles agent operation errors gracefully', async () => {
    const errorMessage = 'Failed to update agent status';
    const mockToggleAgent = jest.fn().mockRejectedValue(new Error(errorMessage));
    
    (useAgent as jest.Mock).mockReturnValue({
      ...mockUseAgent,
      toggleAgent: mockToggleAgent
    });

    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    const agentCard = screen.getByText('Test Agent 1').closest('div');
    fireEvent.click(agentCard!);
    
    const toggleButton = screen.getByText('Stop');
    fireEvent.click(toggleButton);
    
    expect(await screen.findByText(errorMessage)).toBeInTheDocument();
  });

  it('updates agent list after successful strategy update', async () => {
    const updatedAgent = {
      ...mockAgents[0],
      strategy: {
        ...mockAgents[0].strategy,
        parameters: {
          entryConditions: ['new condition'],
          exitConditions: ['new exit'],
          riskManagement: {
            stopLoss: 10,
            takeProfit: 20
          }
        }
      }
    };

    const mockUpdateAgent = jest.fn().mockResolvedValue(updatedAgent);
    (useAgent as jest.Mock).mockReturnValue({
      ...mockUseAgent,
      updateAgent: mockUpdateAgent,
      agents: [updatedAgent, mockAgents[1]]
    });

    render(<AgentDashboard mode={TradingMode.DEX} />);
    
    const agentCard = screen.getByText('Test Agent 1').closest('div');
    fireEvent.click(agentCard!);
    
    const editButton = screen.getByText('Edit Strategy');
    fireEvent.click(editButton);
    
    expect(mockUpdateAgent).toHaveBeenCalledWith({ id: '1' });
  });
});
