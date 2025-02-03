import { ThemeProvider } from './components/theme-provider'
import { TradingDashboard } from './components/TradingDashboard'
import { ErrorBoundary } from './components/ErrorBoundary'

export default function App() {
  return (
    <ThemeProvider>
      <div className="min-h-screen bg-background">
        <header className="border-b">
          <div className="container mx-auto p-4">
            <h1 className="text-2xl font-bold">Solana Trading System</h1>
          </div>
        </header>
        <main className="container mx-auto py-6">
          <ErrorBoundary>
            <TradingDashboard />
          </ErrorBoundary>
        </main>
      </div>
    </ThemeProvider>
  )
}
