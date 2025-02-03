import { TradingMode } from '../types/agent';
import { cn } from '../lib/utils';

interface ModeSelectionProps {
  selectedMode: TradingMode;
  onModeSelect: (mode: TradingMode) => void;
  disabled?: boolean;
}

export function ModeSelection({ selectedMode, onModeSelect, disabled }: ModeSelectionProps) {
  return (
    <div className="w-full max-w-2xl mx-auto p-4">
      <div className="grid grid-cols-2 gap-4">
        <button
          onClick={() => onModeSelect(TradingMode.DEX)}
          disabled={disabled}
          className={cn(
            "p-6 rounded-lg border-2 transition-all",
            "flex flex-col items-center justify-center space-y-2",
            selectedMode === TradingMode.DEX
              ? "border-primary bg-primary/10 text-primary"
              : "border-border hover:border-primary/50",
            disabled && "opacity-50 cursor-not-allowed"
          )}
        >
          <h3 className="text-lg font-semibold">DEX Trading</h3>
          <p className="text-sm text-muted-foreground">
            Trade tokens on Solana DEX
          </p>
        </button>

        <button
          onClick={() => onModeSelect(TradingMode.PUMPFUN)}
          disabled={disabled}
          className={cn(
            "p-6 rounded-lg border-2 transition-all",
            "flex flex-col items-center justify-center space-y-2",
            selectedMode === TradingMode.PUMPFUN
              ? "border-primary bg-primary/10 text-primary"
              : "border-border hover:border-primary/50",
            disabled && "opacity-50 cursor-not-allowed"
          )}
        >
          <h3 className="text-lg font-semibold">Pump.fun Trading</h3>
          <p className="text-sm text-muted-foreground">
            Trade meme coins on Pump.fun
          </p>
        </button>
      </div>
    </div>
  );
}
