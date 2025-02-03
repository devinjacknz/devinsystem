import { useState } from 'react';
import { useAgent } from '../../hooks/agent/useAgent';
import { Agent, TradingMode } from '../../types/agent';
import { cn } from '../../lib/utils';

interface AgentDashboardProps {
  mode: TradingMode;
}

export function AgentDashboard({ mode }: AgentDashboardProps) {
  const { agents, isLoading, error, createAgent, updateAgent, toggleAgent } = useAgent();
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);

  return (
    <div className="w-full max-w-4xl mx-auto p-4 space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Trading Agents</h2>
        <div className="space-x-4">
          <button
            onClick={() => createAgent({ name: 'New Agent', mode })}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
          >
            Create Agent
          </button>
          <button
            onClick={() => selectedAgent && updateAgent({ id: selectedAgent.id })}
            disabled={!selectedAgent}
            className={cn(
              "px-4 py-2 rounded-md",
              selectedAgent
                ? "bg-primary text-primary-foreground hover:bg-primary/90"
                : "bg-muted text-muted-foreground cursor-not-allowed"
            )}
          >
            Edit Strategy
          </button>
          <button
            onClick={() => selectedAgent && toggleAgent(selectedAgent.id, selectedAgent.status.status !== 'active')}
            disabled={!selectedAgent}
            className={cn(
              "px-4 py-2 rounded-md",
              selectedAgent
                ? "bg-primary text-primary-foreground hover:bg-primary/90"
                : "bg-muted text-muted-foreground cursor-not-allowed"
            )}
          >
            {selectedAgent?.status.status === 'active' ? 'Stop' : 'Start'}
          </button>
        </div>
      </div>

      {error && (
        <div className="p-4 bg-destructive/10 text-destructive rounded-md">
          {error}
        </div>
      )}

      <div className="space-y-4">
        {isLoading ? (
          <div className="text-center p-4">Loading agents...</div>
        ) : agents.length === 0 ? (
          <div className="text-center p-4 text-muted-foreground">
            No agents created. Click "Create Agent" to get started.
          </div>
        ) : (
          <div className="grid gap-4">
            {agents
              .filter(agent => agent.mode === mode)
              .map(agent => (
                <div
                  key={agent.id}
                  onClick={() => setSelectedAgent(agent)}
                  className={cn(
                    "p-4 rounded-lg border-2 cursor-pointer transition-all",
                    selectedAgent?.id === agent.id
                      ? "border-primary bg-primary/5"
                      : "border-border hover:border-primary/50"
                  )}
                >
                  <div className="flex justify-between items-center">
                    <div>
                      <h3 className="font-semibold">{agent.name}</h3>
                      <p className="text-sm text-muted-foreground">
                        Last executed: {agent.status.lastExecuted}
                      </p>
                    </div>
                    <div className="flex items-center space-x-4">
                      <span className={cn(
                        "px-2 py-1 rounded-full text-xs",
                        agent.status.status === 'active'
                          ? "bg-success/10 text-success"
                          : agent.status.status === 'error'
                          ? "bg-destructive/10 text-destructive"
                          : "bg-muted text-muted-foreground"
                      )}>
                        {agent.status.status}
                      </span>
                      <div className="text-right text-sm">
                        <div>Total Trades: {agent.status.performance.totalTrades}</div>
                        <div>Success Rate: {agent.status.performance.successRate}%</div>
                        <div>PnL: {agent.status.performance.pnl}</div>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
          </div>
        )}
      </div>
    </div>
  );
}
