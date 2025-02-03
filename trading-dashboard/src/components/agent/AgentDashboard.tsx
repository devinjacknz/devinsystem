import { useState } from 'react';
import { useAgent } from '../../hooks/agent/useAgent';
import { Agent, TradingMode, AgentStrategy } from '../../types/agent';
import { cn } from '../../lib/utils';
import { StrategyEditor } from './StrategyEditor';

interface AgentDashboardProps {
  mode: TradingMode;
}

export function AgentDashboard({ mode }: AgentDashboardProps) {
  const { agents, isLoading, error, createAgent, updateAgent, toggleAgent } = useAgent();
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);
  const [editingStrategy, setEditingStrategy] = useState<{ agentId: string; strategy: Omit<AgentStrategy, 'id'> } | null>(null);

  return (
    <div className={cn("w-full max-w-4xl mx-auto p-4 space-y-6")}>
      <div className={cn("flex justify-between items-center")}>
        <h2 className={cn("text-2xl font-bold text-primary")}>Trading Agents</h2>
        <div className={cn("space-x-4")}>
          <button
            onClick={() => createAgent({
              name: 'New Agent',
              mode,
              strategy: {
                name: 'Default Strategy',
                mode,
                parameters: {
                  entryConditions: [],
                  exitConditions: [],
                  riskManagement: {
                    stopLoss: 5,
                    takeProfit: 10,
                    maxPositionSize: 100
                  }
                }
              }
            })}
            className={cn("px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90")}
          >
            Create Agent
          </button>
          <button
            onClick={() => selectedAgent && setEditingStrategy({
              agentId: selectedAgent.id,
              strategy: {
                name: selectedAgent.strategy.name,
                mode: selectedAgent.strategy.mode,
                parameters: selectedAgent.strategy.parameters
              }
            })}
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
        <div className={cn("p-4 bg-destructive/10 text-destructive rounded-md")}>
          {error}
        </div>
      )}

      {editingStrategy && (
        <div className={cn("fixed inset-0 bg-black/50 flex items-center justify-center z-50")}>
          <div className={cn("bg-white rounded-lg shadow-lg w-full max-w-2xl mx-4 overflow-auto max-h-[90vh]")}>
            <StrategyEditor
              strategy={editingStrategy.strategy}
              onSave={(updatedStrategy) => {
                updateAgent({
                  id: editingStrategy.agentId,
                  strategy: updatedStrategy
                });
                setEditingStrategy(null);
              }}
              onCancel={() => setEditingStrategy(null)}
            />
          </div>
        </div>
      )}

      <div className={cn("space-y-4")}>
        {isLoading ? (
          <div className={cn("text-center p-4 text-muted-foreground")}>Loading agents...</div>
        ) : agents.length === 0 ? (
          <div className={cn("text-center p-4 text-muted-foreground")}>
            No agents created. Click "Create Agent" to get started.
          </div>
        ) : (
          <div className={cn("grid gap-4")}>
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
                  <div className={cn("flex justify-between items-center")}>
                    <div>
                      <h3 className={cn("font-semibold")}>{agent.name}</h3>
                      <p className={cn("text-sm text-muted-foreground")}>
                        Last executed: {agent.status.lastExecuted}
                      </p>
                    </div>
                    <div className={cn("flex items-center space-x-4")}>
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
                      <div className={cn("text-right text-sm")}>
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
