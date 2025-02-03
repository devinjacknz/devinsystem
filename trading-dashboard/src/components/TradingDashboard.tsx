import React, { useState, useEffect } from 'react'
import { ErrorBoundary } from './ErrorBoundary'
import { AgentDashboard } from './agent/AgentDashboard'
import { useAuth } from '../hooks/auth/useAuth'
import { TradingMode } from '../types/agent'

interface TradingDashboardProps {
  initialMode: TradingMode;
}

export const TradingDashboard: React.FC<TradingDashboardProps> = ({ initialMode }) => {
  const [mode, setMode] = useState<TradingMode>(initialMode)
  
  useEffect(() => {
    setMode(initialMode)
  }, [initialMode])
  const { isAuthenticated, isLoading: authLoading } = useAuth()

  if (authLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-4">Authentication Required</h2>
          <p className="text-muted-foreground">Please log in to access the trading dashboard.</p>
        </div>
      </div>
    );
  }

  return (
    <ErrorBoundary>
      <div className="container mx-auto py-6 space-y-8">
        <header className="text-center mb-8">
          <h1 className="text-3xl font-bold">Trading Dashboard</h1>
          <p className="text-muted-foreground mt-2">Select your trading mode and manage your agents</p>
        </header>

        <AgentDashboard mode={mode} />
      </div>
    </ErrorBoundary>
  );
}
