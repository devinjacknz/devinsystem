import { render, screen, fireEvent } from '@testing-library/react'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select'
import '@testing-library/jest-dom'
// React is used implicitly by JSX

describe('Select Components', () => {
  const options = [
    { value: 'SOL/USD', label: 'SOL/USD' },
    { value: 'SOL/USDC', label: 'SOL/USDC' },
    { value: 'BONK/SOL', label: 'BONK/SOL' }
  ]

  it('should render select with options', () => {
    render(
      <Select defaultValue="SOL/USD">
        <SelectTrigger>
          <SelectValue placeholder="Select trading pair" />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toBeInTheDocument()
    expect(screen.getByText('SOL/USD')).toBeInTheDocument()
  })

  it('should handle option selection', () => {
    const onValueChange = jest.fn()

    render(
      <Select defaultValue="SOL/USD" onValueChange={onValueChange}>
        <SelectTrigger>
          <SelectValue placeholder="Select trading pair" />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    )

    fireEvent.click(screen.getByRole('combobox'))
    fireEvent.click(screen.getByText('BONK/SOL'))
    
    expect(onValueChange).toHaveBeenCalledWith('BONK/SOL')
  })

  it('should show placeholder when no value selected', () => {
    render(
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="Select trading pair" />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    )

    expect(screen.getByText('Select trading pair')).toBeInTheDocument()
  })

  it('should handle disabled state', () => {
    render(
      <Select disabled defaultValue="SOL/USD">
        <SelectTrigger>
          <SelectValue placeholder="Select trading pair" />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    )

    const trigger = screen.getByRole('combobox')
    expect(trigger).toBeDisabled()
  })

  it('should handle custom className on components', () => {
    render(
      <Select defaultValue="SOL/USD">
        <SelectTrigger className="custom-trigger">
          <SelectValue className="custom-value" placeholder="Select trading pair" />
        </SelectTrigger>
        <SelectContent className="custom-content">
          {options.map((option) => (
            <SelectItem 
              key={option.value} 
              value={option.value}
              className="custom-item"
            >
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toHaveClass('custom-trigger')
  })
})
