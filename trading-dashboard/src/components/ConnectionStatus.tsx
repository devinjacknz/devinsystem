import React from 'react'

interface ConnectionStatusProps {
  isPriceConnected?: boolean
  isPositionsConnected?: boolean
}

export const ConnectionStatus: React.FC<ConnectionStatusProps> = ({
  isPriceConnected,
  isPositionsConnected
}) => {
  const isLoading = typeof isPriceConnected === 'undefined' || typeof isPositionsConnected === 'undefined'
  const isFullyConnected = isPriceConnected && isPositionsConnected
  const isPartiallyConnected = isPriceConnected || isPositionsConnected

  const getStatusColor = () => {
    if (isLoading) return 'bg-gray-500 animate-pulse'
    if (isFullyConnected) return 'bg-green-500'
    if (isPartiallyConnected) return 'bg-yellow-500'
    return 'bg-red-500'
  }

  const getStatusText = () => {
    if (isLoading) return 'Connecting'
    if (isFullyConnected) return 'Connected'
    if (isPartiallyConnected) return 'Partial Connection'
    return 'Disconnected'
  }

  const getTooltipText = () => {
    return `Price feed: ${isPriceConnected ? 'Connected' : 'Disconnected'}\nPosition tracking: ${isPositionsConnected ? 'Connected' : 'Disconnected'}`
  }

  return (
    <div className="flex items-center space-x-2">
      <div
        data-testid="status-indicator"
        className={`w-2 h-2 rounded-full ${getStatusColor()}`}
        title={getTooltipText()}
      />
      <span className="text-sm font-medium">{getStatusText()}</span>
      {isPartiallyConnected && !isFullyConnected && (
        <span className="text-sm text-muted-foreground">
          {!isPriceConnected && 'Price feed disconnected'}
          {!isPositionsConnected && 'Position tracking disconnected'}
        </span>
      )}
    </div>
  )
}
