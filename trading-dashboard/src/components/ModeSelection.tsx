import { TradingMode } from '../types/agent';
import { cn } from '../lib/utils';

interface ModeSelectionProps {
  selectedMode: TradingMode;
  onModeSelect: (mode: TradingMode) => void;
  disabled?: boolean;
  className?: string;
}

export function ModeSelection({ selectedMode, onModeSelect, disabled, className }: ModeSelectionProps) {
  return (
    <div className={cn("w-full max-w-2xl mx-auto p-4", className)}>
      <div className="flex rounded-md overflow-hidden border border-border">
        <button
          onClick={() => onModeSelect(TradingMode.DEX)}
          disabled={disabled}
          className={cn(
            "px-4 py-2 transition-colors",
            selectedMode === TradingMode.DEX
              ? "bg-primary text-primary-foreground"
              : "hover:bg-secondary",
            disabled && "opacity-50 cursor-not-allowed"
          )}
        >
          DEX
        </button>

        <button
          onClick={() => onModeSelect(TradingMode.PUMPFUN)}
          disabled={disabled}
          className={cn(
            "px-4 py-2 transition-colors",
            selectedMode === TradingMode.PUMPFUN
              ? "bg-primary text-primary-foreground"
              : "hover:bg-secondary",
            disabled && "opacity-50 cursor-not-allowed"
          )}
        >
          Pump.fun
        </button>
      </div>
    </div>
  );
}
