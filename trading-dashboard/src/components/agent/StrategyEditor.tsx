import { useState } from 'react';
import { AgentStrategy } from '../../types/agent';
import { cn } from '../../lib/utils';

interface StrategyEditorProps {
  strategy: Omit<AgentStrategy, 'id'>;
  onSave: (strategy: Omit<AgentStrategy, 'id'>) => void;
  onCancel: () => void;
}

export function StrategyEditor({ strategy: initialStrategy, onSave, onCancel }: StrategyEditorProps) {
  const [strategy, setStrategy] = useState<Omit<AgentStrategy, 'id'>>(initialStrategy);

  const updateEntryCondition = (index: number, field: 'type' | 'value', value: string | number) => {
    setStrategy(prev => ({
      ...prev,
      parameters: {
        ...prev.parameters,
        entryConditions: prev.parameters.entryConditions.map((condition, i) =>
          i === index ? { ...condition, [field]: value } : condition
        )
      }
    }));
  };

  const updateExitCondition = (index: number, field: 'type' | 'value', value: string | number) => {
    setStrategy(prev => ({
      ...prev,
      parameters: {
        ...prev.parameters,
        exitConditions: prev.parameters.exitConditions.map((condition, i) =>
          i === index ? { ...condition, [field]: value } : condition
        )
      }
    }));
  };

  const updateRiskManagement = (field: keyof AgentStrategy['parameters']['riskManagement'], value: number) => {
    setStrategy(prev => ({
      ...prev,
      parameters: {
        ...prev.parameters,
        riskManagement: {
          ...prev.parameters.riskManagement,
          [field]: value
        }
      }
    }));
  };

  const addCondition = (type: 'entry' | 'exit') => {
    setStrategy(prev => ({
      ...prev,
      parameters: {
        ...prev.parameters,
        [type === 'entry' ? 'entryConditions' : 'exitConditions']: [
          ...(type === 'entry' ? prev.parameters.entryConditions : prev.parameters.exitConditions),
          { type: 'price', value: 0 }
        ]
      }
    }));
  };

  const removeCondition = (type: 'entry' | 'exit', index: number) => {
    setStrategy(prev => ({
      ...prev,
      parameters: {
        ...prev.parameters,
        [type === 'entry' ? 'entryConditions' : 'exitConditions']:
          (type === 'entry' ? prev.parameters.entryConditions : prev.parameters.exitConditions)
            .filter((_, i) => i !== index)
      }
    }));
  };

  return (
    <div className={cn("bg-white rounded-lg shadow-lg p-6 max-w-2xl mx-auto")}>
      <div className={cn("flex justify-between items-center mb-6")}>
        <h2 className={cn("text-xl font-bold text-primary")}>Strategy Configuration</h2>
        <div className={cn("space-x-2")}>
          <button onClick={onCancel} className={cn("btn-secondary")}>
            Cancel
          </button>
          <button onClick={() => onSave(strategy)} className={cn("btn-primary")}>
            Save Changes
          </button>
        </div>
      </div>

      <div className={cn("space-y-6")}>
        <div>
          <h3 className={cn("text-lg font-semibold mb-3")}>Entry Conditions</h3>
          {strategy.parameters.entryConditions.map((condition, index) => (
            <div key={index} className={cn("flex items-center space-x-2 mb-2")}>
              <select
                value={condition.type}
                onChange={(e) => updateEntryCondition(index, 'type', e.target.value)}
                className={cn("input")}
              >
                <option value="price">Price</option>
                <option value="volume">Volume</option>
                <option value="trend">Trend</option>
              </select>
              <input
                type="number"
                value={condition.value}
                onChange={(e) => updateEntryCondition(index, 'value', parseFloat(e.target.value))}
                className={cn("input")}
              />
              <button
                onClick={() => removeCondition('entry', index)}
                className={cn("text-destructive hover:text-destructive/90")}
              >
                Remove
              </button>
            </div>
          ))}
          <button onClick={() => addCondition('entry')} className={cn("btn-ghost mt-2")}>
            Add Entry Condition
          </button>
        </div>

        <div>
          <h3 className={cn("text-lg font-semibold mb-3")}>Exit Conditions</h3>
          {strategy.parameters.exitConditions.map((condition, index) => (
            <div key={index} className={cn("flex items-center space-x-2 mb-2")}>
              <select
                value={condition.type}
                onChange={(e) => updateExitCondition(index, 'type', e.target.value)}
                className={cn("input")}
              >
                <option value="price">Price</option>
                <option value="volume">Volume</option>
                <option value="trend">Trend</option>
              </select>
              <input
                type="number"
                value={condition.value}
                onChange={(e) => updateExitCondition(index, 'value', parseFloat(e.target.value))}
                className="input"
              />
              <button
                onClick={() => removeCondition('exit', index)}
                className={cn("text-destructive hover:text-destructive/90")}
              >
                Remove
              </button>
            </div>
          ))}
          <button onClick={() => addCondition('exit')} className={cn("btn-ghost mt-2")}>
            Add Exit Condition
          </button>
        </div>

        <div>
          <h3 className={cn("text-lg font-semibold mb-3")}>Risk Management</h3>
          <div className={cn("grid grid-cols-3 gap-4")}>
            <div>
              <label className={cn("block text-sm font-medium mb-1")}>Stop Loss (%)</label>
              <input
                type="number"
                value={strategy.parameters.riskManagement.stopLoss}
                onChange={(e) => updateRiskManagement('stopLoss', parseFloat(e.target.value))}
                className={cn("input w-full")}
              />
            </div>
            <div>
              <label className={cn("block text-sm font-medium mb-1")}>Take Profit (%)</label>
              <input
                type="number"
                value={strategy.parameters.riskManagement.takeProfit}
                onChange={(e) => updateRiskManagement('takeProfit', parseFloat(e.target.value))}
                className={cn("input w-full")}
              />
            </div>
            <div>
              <label className={cn("block text-sm font-medium mb-1")}>Max Position Size</label>
              <input
                type="number"
                value={strategy.parameters.riskManagement.maxPositionSize}
                onChange={(e) => updateRiskManagement('maxPositionSize', parseFloat(e.target.value))}
                className={cn("input w-full")}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
