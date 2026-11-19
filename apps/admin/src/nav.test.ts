import { describe, expect, it } from 'vitest'
import { canAccess, visibleNav } from './nav'

describe('nav', () => {
  it('trims the empty frame by role', () => {
    expect(visibleNav('duty').map((item) => item.label)).toEqual(['工作台'])
    expect(visibleNav('admin').map((item) => item.label)).toEqual([
      '工作台',
      '权限占位',
    ])
    expect(canAccess('duty', '/access')).toBe(false)
    expect(canAccess('admin', '/access')).toBe(true)
  })
})
