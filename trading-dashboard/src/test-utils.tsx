import React from 'react'
import { render, RenderOptions } from '@testing-library/react'
import '@testing-library/jest-dom'

const customRender = (
  ui: React.ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => render(ui, { ...options })

export * from '@testing-library/react'
export { customRender as render }
