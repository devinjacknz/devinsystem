import { ThemeProvider } from './components/theme-provider'
import { TradingDashboard } from './components/TradingDashboard'
import { ErrorBoundary } from './components/ErrorBoundary'
import { Navbar } from './components/layout/Navbar'
import { TradingMode } from './types/agent'
import { useState } from 'react'
import { cn } from './lib/utils'

const DEFAULT_MODE = TradingMode.DEX

export default function App() {
  const [mode, setMode] = useState<TradingMode>(DEFAULT_MODE)
  return (
    <ThemeProvider>
      <div className={cn("min-h-screen bg-background")}>
        <Navbar currentMode={mode} onModeChange={setMode} />
        <main className={cn("container mx-auto py-6")}>
          <ErrorBoundary>
            <TradingDashboard initialMode={mode} />
          </ErrorBoundary>
        </main>
      </div>
    </ThemeProvider>
  )
}
