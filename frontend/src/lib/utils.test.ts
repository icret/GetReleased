import { describe, it, expect } from 'vitest'
import { cn } from './utils'

describe('cn', () => {
  it('合并类名', () => {
    expect(cn('a', 'b')).toBe('a b')
  })

  it('处理条件类', () => {
    // eslint-disable-next-line no-constant-binary-expression
    expect(cn('a', false && 'b', 'c')).toBe('a c')
  })

  it('tailwind-merge 去重冲突', () => {
    expect(cn('px-2', 'px-4')).toBe('px-4')
  })
})
