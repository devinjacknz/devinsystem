import { Component, ErrorInfo, ReactNode } from 'react'
import { Card, CardContent } from './ui/card'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false
  }

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught error:', error, errorInfo)
    this.setState({ hasError: true, error })
  }

  public render() {
    if (this.state.hasError) {
      return (
        <Card className="p-4">
          <CardContent className="space-y-4">
            <div className="flex flex-col items-center space-y-2" role="alert">
              <div className="w-2 h-2 rounded-full bg-red-500" data-testid="status-indicator" aria-hidden="true" />
              <h1 className="text-lg font-semibold text-red-500" role="heading" aria-level={1} data-testid="error-title">
                An error occurred in the trading dashboard
              </h1>
              <div className="text-sm text-muted-foreground" data-testid="error-description">
                {this.state.error?.message || 'Critical trading system error'}
              </div>
              <button
                onClick={() => window.location.reload()}
                className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90"
                aria-label="Reload Application"
              >
                Reload Application
              </button>
            </div>
          </CardContent>
        </Card>
      )
    }

    return this.props.children
  }
}
