import { describe, it, expect } from 'vitest'
import { extractExcerpt } from '@/lib/excerpt'

describe('extractExcerpt', () => {
  it('跳过标题行取首个正文行', () => {
    expect(extractExcerpt('## 标题\n首个正文行\n第二行')).toBe('首个正文行')
  })

  it('跳过引用与 HTML 行', () => {
    expect(extractExcerpt('> 引用\n<div>html</div>\n正文')).toBe('正文')
  })

  it('跳过图片行', () => {
    expect(extractExcerpt('![图片](url)\n正文')).toBe('正文')
  })

  it('无匹配行时截取前 120 字符', () => {
    const body = '# 标题\n## 另一个标题'
    expect(extractExcerpt(body)).toBe(body.trim())
  })

  it('空 body 返回占位文本', () => {
    expect(extractExcerpt('')).toBe('暂无 Release Notes')
  })

  it('仅有空白行返回占位文本', () => {
    expect(extractExcerpt('   \n  \n')).toBe('暂无 Release Notes')
  })
})
