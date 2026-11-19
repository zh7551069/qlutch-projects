import { describe, expect, it } from 'vitest'
import { envLabel } from './env'

describe('envLabel', () => {
  it('uses different labels for local, staging, and production', () => {
    expect(envLabel('local')).toBe('本地')
    expect(envLabel('staging')).toBe('预发')
    expect(envLabel('production')).toBe('生产')
    expect(
      new Set([envLabel('local'), envLabel('staging'), envLabel('production')])
        .size,
    ).toBe(3)
  })
})
