import { render, screen, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs'

describe('Tabs Component', () => {
  const user = userEvent.setup()

  afterEach(() => {
    cleanup()
  })

  it('should render basic tabs', () => {
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
      </Tabs>
    )
    
    expect(screen.getByRole('tab', { name: /tab 1/i })).toBeInTheDocument()
    expect(screen.getByText('Content 1')).toBeInTheDocument()
  })

  it('should switch tabs when clicked', async () => {
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
      </Tabs>
    )
    
    const tab2 = screen.getByRole('tab', { name: /tab 2/i })
    await user.click(tab2)
    
    expect(screen.getByText('Content 2')).toBeVisible()
  })

  it('should support keyboard navigation with ARIA updates', async () => {
    const user = userEvent.setup({ delay: null })
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
          <TabsTrigger value="tab3">Tab 3</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
        <TabsContent value="tab3">Content 3</TabsContent>
      </Tabs>
    )

    const tab1 = screen.getByRole('tab', { name: /tab 1/i })
    const tab2 = screen.getByRole('tab', { name: /tab 2/i })
    const tab3 = screen.getByRole('tab', { name: /tab 3/i })

    // Test keyboard navigation
    await user.tab()
    expect(tab1).toHaveFocus()

    await user.keyboard('[ArrowRight]')
    expect(tab2).toHaveFocus()

    await user.keyboard('[End]')
    expect(tab3).toHaveFocus()

    await user.keyboard('[Home]')
    expect(tab1).toHaveFocus()
  })

  it('should handle disabled tabs', () => {
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2" disabled>Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
      </Tabs>
    )

    const disabledTab = screen.getByRole('tab', { name: /tab 2/i })
    expect(disabledTab).toBeDisabled()
    expect(screen.getByText('Content 1')).toBeVisible()
  })

  it('should handle forceMount behavior', () => {
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1" forceMount>Content 1</TabsContent>
        <TabsContent value="tab2" forceMount>Content 2</TabsContent>
      </Tabs>
    )

    const tab1Content = screen.getByText('Content 1')
    const tab2Content = screen.getByText('Content 2')

    expect(tab1Content).toBeVisible()
    expect(tab2Content).toBeInTheDocument()
  })


  it('should preserve ARIA attributes with custom className', () => {
    render(
      <Tabs defaultValue="tab1" className="custom-tabs" aria-label="Custom sections">
        <TabsList className="custom-list" aria-label="Custom views">
          <TabsTrigger value="tab1" aria-controls="tab1-content" className="custom-trigger">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2" aria-controls="tab2-content" className="custom-trigger">Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1" id="tab1-content" role="tabpanel" aria-label="Content 1" className="custom-content">
          Content 1
        </TabsContent>
        <TabsContent value="tab2" id="tab2-content" role="tabpanel" aria-label="Content 2" className="custom-content">
          Content 2
        </TabsContent>
      </Tabs>
    )
    
    // Verify custom classes are applied while maintaining ARIA attributes
    const tablist = screen.getByRole('tablist', { name: 'Custom views' })
    expect(tablist).toHaveClass('custom-list')
    
    const tabs = screen.getAllByRole('tab')
    tabs.forEach(tab => {
      expect(tab).toHaveClass('custom-trigger')
      expect(tab).toHaveAttribute('aria-controls')
    })
    
    const panels = screen.getAllByRole('tabpanel')
    panels.forEach(panel => {
      expect(panel).toHaveClass('custom-content')
      expect(panel).toHaveAttribute('aria-label')
    })
    
    // Verify tab container maintains its ARIA label
    expect(screen.getByRole('tablist').parentElement).toHaveAttribute('aria-label', 'Custom sections')
  })

})
