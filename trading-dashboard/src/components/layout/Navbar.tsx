import { TradingMode } from '../../types/agent';

interface NavbarProps {
  onModeChange: (mode: TradingMode) => void;
  currentMode: TradingMode;
}

export function Navbar({ onModeChange, currentMode }: NavbarProps) {
  return (
    <nav className="w-full bg-white border-b border-border px-4 py-3">
      <div className="container mx-auto flex justify-between items-center">
        <h1 className="text-xl font-bold text-primary">Trading System</h1>
        <div className="flex items-center space-x-4">
          <div className="flex rounded-md overflow-hidden border border-border">
            <button
              onClick={() => onModeChange(TradingMode.DEX)}
              className={`nav-item ${currentMode === TradingMode.DEX ? 'nav-item-active' : ''}`}
            >
              DEX
            </button>
            <button
              onClick={() => onModeChange(TradingMode.PUMPFUN)}
              className={`nav-item ${currentMode === TradingMode.PUMPFUN ? 'nav-item-active' : ''}`}
            >
              Pump.fun
            </button>
          </div>
        </div>
      </div>
    </nav>
  );
}
