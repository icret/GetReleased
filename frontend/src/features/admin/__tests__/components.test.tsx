import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { ConfirmDialog } from '@/features/admin/ConfirmDialog'
import { RepositoryTable } from '@/features/admin/RepositoryTable'
import { RepositoryDialog } from '@/features/admin/RepositoryDialog'
import type { Repository, Tag } from '@/types'

afterEach(cleanup)

describe('ConfirmDialog', () => {
  it('渲染标题和描述', () => {
    render(<ConfirmDialog open={true} title="删除仓库" description="确认删除？" onConfirm={() => {}} onCancel={() => {}} />)
    expect(screen.getByText('删除仓库')).toBeDefined()
    expect(screen.getByText('确认删除？')).toBeDefined()
  })

  it('未打开时不渲染内容', () => {
    render(<ConfirmDialog open={false} title="不可见" onConfirm={() => {}} onCancel={() => {}} />)
    expect(screen.queryByText('不可见')).toBeNull()
  })
})

describe('RepositoryTable', () => {
  const repos: Repository[] = [
    {
      id: 1,
      owner: 'octocat',
      name: 'hello',
      full_name: 'octocat/hello',
      description: 'demo repo',
      stars: 0,
      is_archived: false,
      is_private: false,
      latest_version: 'v1.2.3',
      latest_release_url: 'https://github.com/octocat/hello/releases/tag/v1.2.3',
      last_checked_at: '2026-08-29T12:00:00Z',
      remark: '内部备注',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
      tags: [{ id: 1, name: 'go', type: 'category' }],
    },
  ]

  it('渲染仓库全名、描述、版本与备注', () => {
    render(<RepositoryTable repos={repos} onSync={() => {}} onEdit={() => {}} onDelete={() => {}} />)
    expect(screen.getByText('octocat/hello')).toBeDefined()
    expect(screen.getByText('demo repo')).toBeDefined()
    expect(screen.getByText('v1.2.3')).toBeDefined()
    expect(screen.getByText('内部备注')).toBeDefined()
  })

  it('渲染标签', () => {
    render(<RepositoryTable repos={repos} onSync={() => {}} onEdit={() => {}} onDelete={() => {}} />)
    expect(screen.getByText('go')).toBeDefined()
  })

  it('未同步仓库显示占位符', () => {
    const emptyRepos: Repository[] = [
      {
        id: 2,
        owner: 'o',
        name: 'n',
        full_name: 'o/n',
        description: '',
        stars: 0,
        is_archived: false,
        is_private: false,
        created_at: '2024-01-01',
        updated_at: '2024-01-01',
      },
    ]
    render(<RepositoryTable repos={emptyRepos} onSync={() => {}} onEdit={() => {}} onDelete={() => {}} />)
    expect(screen.getByText('未同步')).toBeDefined()
  })
})

describe('RepositoryDialog', () => {
  const tags: Tag[] = [
    { id: 1, name: 'go', type: 'platform' },
    { id: 2, name: 'rust', type: 'category' },
  ]

  it('新增模式显示空表单', () => {
    render(<RepositoryDialog open={true} tags={tags} onSubmit={async () => {}} onClose={() => {}} />)
    expect(screen.getByText('新增仓库')).toBeDefined()
    expect(screen.getByText('分类')).toBeDefined()
    expect(screen.getByText('标签')).toBeDefined()
  })

  it('编辑模式锁定 owner/name 并回填备注', () => {
    const repo: Repository = {
      id: 1,
      owner: 'octocat',
      name: 'hello',
      full_name: 'octocat/hello',
      description: 'desc',
      remark: '内部备注内容',
      stars: 0,
      is_archived: false,
      is_private: false,
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    }
    render(<RepositoryDialog open={true} repo={repo} tags={tags} onSubmit={async () => {}} onClose={() => {}} />)
    expect(screen.getByText('编辑仓库')).toBeDefined()
    const repoUrlInput = screen.getByDisplayValue('octocat/hello') as HTMLInputElement
    expect(repoUrlInput.disabled).toBe(true)
    expect(screen.getByDisplayValue('内部备注内容')).toBeDefined()
  })
})
