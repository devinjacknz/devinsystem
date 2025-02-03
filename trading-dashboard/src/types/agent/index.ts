export enum TradingMode {
  DEX = 'dex',
  PUMPFUN = 'pumpfun'
}

export interface AgentStrategy {
  id: string;
  name: string;
  mode: TradingMode;
  parameters: {
    entryConditions: {
      type: string;
      value: number;
    }[];
    exitConditions: {
      type: string;
      value: number;
    }[];
    riskManagement: {
      stopLoss: number;
      takeProfit: number;
      maxPositionSize: number;
    };
  };
}

export interface AgentStatus {
  id: string;
  status: 'active' | 'paused' | 'error';
  lastExecuted: string;
  performance: {
    totalTrades: number;
    successRate: number;
    pnl: number;
  };
  error?: {
    code: string;
    message: string;
    timestamp: string;
  };
}

export interface Agent {
  id: string;
  name: string;
  mode: TradingMode;
  strategy: AgentStrategy;
  status: AgentStatus;
  createdAt: string;
  updatedAt: string;
}

export interface CreateAgentRequest {
  name: string;
  mode: TradingMode;
  strategy: Omit<AgentStrategy, 'id'>;
}

export interface UpdateAgentRequest {
  id: string;
  name?: string;
  strategy?: Partial<AgentStrategy['parameters']>;
}

export interface AgentExecutionRecord {
  id: string;
  agentId: string;
  action: 'buy' | 'sell';
  symbol: string;
  amount: number;
  price: number;
  timestamp: string;
  status: 'pending' | 'completed' | 'failed';
  error?: string;
}
