import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { TagTable } from '@/features/admin/TagTable'
import { TagDialog } from '@/features/admin/TagDialog'
import type { Tag } from '@/types'

afterEach(cleanup)

describe('TagTable', () => {
  const tags: Tag[] = [
    { id: 1, name: 'go', type: 'category' },
    { id: 2, name: 'rust', type: 'category' },
  ]

  it('渲染标签名称', () => {
    render(<TagTable tags={tags} onEdit={() => {}} onDelete={() => {}} />)
    expect(screen.getByText('go')).toBeDefined()
    expect(screen.getByText('rust')).toBeDefined()
  })
})

describe('TagDialog', () => {
  it('新增模式显示空表单', () => {
    render(<TagDialog open={true} onSubmit={async () => {}} onClose={() => {}} />)
    expect(screen.getByText('新增标签')).toBeDefined()
    expect(screen.getByText('名称')).toBeDefined()
  })

  it('编辑模式预填名称', () => {
    const tag: Tag = { id: 1, name: 'existing', type: 'category' }
    render(<TagDialog open={true} tag={tag} onSubmit={async () => {}} onClose={() => {}} />)
    expect(screen.getByText('编辑标签')).toBeDefined()
    expect(screen.getByDisplayValue('existing')).toBeDefined()
  })
})
