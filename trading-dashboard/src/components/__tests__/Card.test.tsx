import { render, screen } from '@testing-library/react'
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card'
import '@testing-library/jest-dom'

describe('Card Components', () => {
  it('should render Card with children', () => {
    render(
      <Card>
        <div>Test Content</div>
      </Card>
    )
    expect(screen.getByText('Test Content')).toBeInTheDocument()
  })

  it('should render CardHeader with title', () => {
    render(
      <Card>
        <CardHeader>
          <CardTitle>Test Title</CardTitle>
        </CardHeader>
      </Card>
    )
    expect(screen.getByText('Test Title')).toBeInTheDocument()
  })

  it('should render CardContent with content', () => {
    render(
      <Card>
        <CardContent>
          <p>Test Content</p>
        </CardContent>
      </Card>
    )
    expect(screen.getByText('Test Content')).toBeInTheDocument()
  })

  it('should apply custom className to Card', () => {
    render(
      <Card className="custom-class">
        <div>Test Content</div>
      </Card>
    )
    expect(screen.getByText('Test Content').parentElement).toHaveClass('custom-class')
  })

  it('should render nested card components', () => {
    render(
      <Card>
        <CardHeader>
          <CardTitle>Header Title</CardTitle>
        </CardHeader>
        <CardContent>
          <p>Content Text</p>
        </CardContent>
      </Card>
    )
    expect(screen.getByText('Header Title')).toBeInTheDocument()
    expect(screen.getByText('Content Text')).toBeInTheDocument()
  })

  it('should handle empty card components', () => {
    render(
      <Card>
        <CardHeader />
        <CardContent />
      </Card>
    )
    const cardElement = screen.getByTestId('card')
    expect(cardElement).toBeInTheDocument()
  })
})
