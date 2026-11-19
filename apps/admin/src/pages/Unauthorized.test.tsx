import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Unauthorized } from './Unauthorized'

describe('Unauthorized', () => {
  it('shows the empty unauthorized state', () => {
    render(<Unauthorized />)
    expect(
      screen.getByRole('heading', { name: '这个位置没有对当前角色开放' }),
    ).toBeTruthy()
    expect(screen.getByText('这是空态占位，没有业务数据。')).toBeTruthy()
  })
})
