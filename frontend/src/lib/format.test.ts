import { describe, it, expect } from 'vitest'
import { formatBytes, formatDate, formatRelativeTime } from './format'

const now = new Date('2026-08-29T12:00:00Z')

describe('formatRelativeTime', () => {
  it('1 分钟内显示刚刚', () => {
    expect(formatRelativeTime('2026-08-29T11:59:40Z', now)).toBe('刚刚')
  })

  it('1 小时内显示分钟前', () => {
    expect(formatRelativeTime('2026-08-29T11:30:00Z', now)).toBe('30 分钟前')
  })

  it('24 小时内显示小时前', () => {
    expect(formatRelativeTime('2026-08-29T09:00:00Z', now)).toBe('3 小时前')
  })

  it('30 天内显示天前', () => {
    expect(formatRelativeTime('2026-08-15T12:00:00Z', now)).toBe('14 天前')
  })

  it('超过 30 天显示绝对日期', () => {
    expect(formatRelativeTime('2026-07-19T17:00:26Z', now)).toBe(formatDate('2026-07-19T17:00:26Z'))
  })

  it('未来 1 天显示 1 天前', () => {
    expect(formatRelativeTime('2026-08-30T12:00:00Z', now)).toBe('1 天前')
  })

  it('非法时间回退为绝对日期', () => {
    expect(formatRelativeTime('not-a-date', now)).toBe(formatDate('not-a-date'))
  })
})

describe('formatDate', () => {
  it('按 UTC 格式化为中文长日期', () => {
    expect(formatDate('2026-07-19T17:00:26Z')).toBe('2026年7月19日')
  })

  it('一小时后的本地时区差异不改变结果', () => {
    expect(formatDate('2026-07-19T23:59:00Z')).toBe('2026年7月19日')
  })
})

describe('formatBytes', () => {
  it('小于 1024 显示字节', () => {
    expect(formatBytes(0)).toBe('0 B')
    expect(formatBytes(512)).toBe('512 B')
  })

  it('KB 以上带单位', () => {
    expect(formatBytes(2048)).toBe('2 KB')
    expect(formatBytes(1048576)).toBe('1 MB')
  })

  it('小于 10 的值保留一位小数', () => {
    expect(formatBytes(1536)).toBe('1.5 KB')
  })

  it('非法值回退 0 B', () => {
    expect(formatBytes(-1)).toBe('0 B')
    expect(formatBytes(Number.NaN)).toBe('0 B')
  })
})
