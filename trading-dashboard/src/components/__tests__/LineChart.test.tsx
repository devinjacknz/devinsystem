import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import type { ReactElement } from 'react'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'

interface RechartsProps {
  children?: ReactElement | ReactElement[]
  data?: Array<Record<string, any>>
  width?: string | number
  height?: string | number
  type?: string
  dataKey?: string
  stroke?: string
  dot?: boolean | object
  strokeWidth?: number
  tickFormatter?: (value: any) => string
  strokeDasharray?: string
  className?: string
  [key: string]: any
}
jest.mock('recharts', () => ({
  LineChart: ({ children, data, ...props }: RechartsProps) => (
    <div data-testid="line-chart" {...props}>
      <div data-testid="chart-container">
        {children}
        <div data-testid="x-axis" />
        <div data-testid="y-axis" />
        <div data-testid="tooltip" />
        <div data-testid="legend" />
        <div data-testid="cartesian-grid" />
        <div data-testid="line" />
      </div>
    </div>
  ),
  Line: ({ type, dataKey, stroke, dot, strokeWidth, ...props }: RechartsProps) => (
    <div data-testid="line" data-type={type} data-datakey={dataKey} data-stroke={stroke} data-dot={String(dot)} data-strokewidth={String(strokeWidth)} {...props} />
  ),
  XAxis: ({ dataKey, tickFormatter, ...props }: RechartsProps) => (
    <div data-testid="x-axis" data-datakey={dataKey} data-tickformatter={tickFormatter?.toString()} {...props} />
  ),
  YAxis: ({ tickFormatter, ...props }: RechartsProps) => (
    <div data-testid="y-axis" data-tickformatter={tickFormatter?.toString()} {...props} />
  ),
  CartesianGrid: ({ strokeDasharray, ...props }: RechartsProps) => (
    <div data-testid="cartesian-grid" data-strokedasharray={strokeDasharray} {...props} />
  ),
  Tooltip: (props: RechartsProps) => <div data-testid="tooltip" {...props} />,
  ResponsiveContainer: ({ children, width, height, ...props }: RechartsProps) => (
    <div data-testid="responsive-container" style={{ width, height }} {...props}>{children}</div>
  ),
  Legend: (props: RechartsProps) => <div data-testid="legend" {...props} />
}))

const mockPriceData = [
  { timestamp: 1641024000000, price: 100 },
  { timestamp: 1641027600000, price: 105 },
  { timestamp: 1641031200000, price: 102 }
]

describe('LineChart Component', () => {
  it('should render chart with all components', () => {
    render(
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={mockPriceData}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="timestamp" />
          <YAxis />
          <Tooltip />
          <Legend />
          <Line type="monotone" dataKey="price" stroke="#8884d8" />
        </LineChart>
      </ResponsiveContainer>
    )

    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
    expect(screen.getByTestId('line-chart')).toBeInTheDocument()
    expect(screen.getAllByTestId('cartesian-grid')[0]).toBeInTheDocument()
    expect(screen.getAllByTestId('x-axis')[0]).toBeInTheDocument()
    expect(screen.getAllByTestId('y-axis')[0]).toBeInTheDocument()
    expect(screen.getAllByTestId('tooltip')[0]).toBeInTheDocument()
    expect(screen.getAllByTestId('legend')[0]).toBeInTheDocument()
    expect(screen.getAllByTestId('line')[0]).toBeInTheDocument()
  })

  it('should pass correct props to Line component', () => {
    render(
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={mockPriceData}>
          <Line 
            type="monotone" 
            dataKey="price" 
            stroke="#8884d8"
            dot={false}
            strokeWidth={2}
          />
        </LineChart>
      </ResponsiveContainer>
    )

    const line = screen.getAllByTestId('line')[0]
    expect(line).toHaveAttribute('data-type', 'monotone')
    expect(line).toHaveAttribute('data-datakey', 'price')
    expect(line).toHaveAttribute('data-stroke', '#8884d8')
    expect(line).toHaveAttribute('data-strokewidth', '2')
    expect(line).toHaveAttribute('data-dot', 'false')
  })

  it('should handle empty data', () => {
    render(
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={[]}>
          <Line type="monotone" dataKey="price" stroke="#8884d8" />
        </LineChart>
      </ResponsiveContainer>
    )

    expect(screen.getByTestId('line-chart')).toBeInTheDocument()
  })

  it('should format axis ticks correctly', () => {
    const formatTimestamp = (timestamp: number) => 
      new Date(timestamp).toLocaleString()
    
    const formatCurrency = (value: number) =>
      new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value)

    render(
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={mockPriceData}>
          <XAxis 
            dataKey="timestamp" 
            tickFormatter={formatTimestamp}
          />
          <YAxis 
            tickFormatter={formatCurrency}
          />
        </LineChart>
      </ResponsiveContainer>
    )

    const xAxis = screen.getAllByTestId('x-axis')[0]
    const yAxis = screen.getAllByTestId('y-axis')[0]
    
    expect(xAxis).toHaveAttribute('data-tickformatter')
    expect(yAxis).toHaveAttribute('data-tickformatter')
  })
})
