import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { EnvBanner } from './EnvBanner'

describe('EnvBanner', () => {
  it('keeps local and production visually distinct', () => {
    const { rerender } = render(<EnvBanner env="local" />)
    const local = screen.getByRole('status')
    expect(local.getAttribute('data-env')).toBe('local')
    expect(local.textContent).toContain('本地')

    rerender(<EnvBanner env="production" />)
    const production = screen.getByRole('status')
    expect(production.getAttribute('data-env')).toBe('production')
    expect(production.textContent).toContain('生产')
    expect(production.textContent).not.toContain('本地')
  })
})
