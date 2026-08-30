import { describe, it, expect } from 'vitest'
import { gradientFor } from '@/lib/gradient'

describe('gradientFor', () => {
  it('返回合法的渐变类名', () => {
    const result = gradientFor('test-repo')
    expect(result).toMatch(/^from-\S+\s+to-\S+$/)
  })

  it('相同名称返回相同结果', () => {
    expect(gradientFor('react')).toBe(gradientFor('react'))
  })

  it('不同名称可能返回不同结果', () => {
    const results = new Set(['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'].map(gradientFor))
    expect(results.size).toBeGreaterThan(1)
  })

  it('空字符串不报错', () => {
    expect(() => gradientFor('')).not.toThrow()
  })
})
