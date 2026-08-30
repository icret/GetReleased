import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { UserTable } from '@/features/admin/UserTable'
import { UserDialog } from '@/features/admin/UserDialog'
import { ResetPasswordDialog } from '@/features/admin/ResetPasswordDialog'
import type { User } from '@/types'

afterEach(cleanup)

describe('UserTable', () => {
  const users: User[] = [
    { id: 1, username: 'admin', role: 'admin', created_at: '2024-01-01 00:00:00' },
    { id: 2, username: 'alice', role: 'admin', created_at: '2024-01-02 00:00:00' },
  ]

  it('渲染用户名和角色', () => {
    render(<UserTable users={users} currentUsername={null} onResetPassword={() => {}} onDelete={() => {}} />)
    expect(screen.getAllByText('admin').length).toBeGreaterThan(0)
    expect(screen.getByText('alice')).toBeDefined()
  })

  it('当前用户显示标记', () => {
    render(<UserTable users={users} currentUsername="admin" onResetPassword={() => {}} onDelete={() => {}} />)
    expect(screen.getByText('当前')).toBeDefined()
  })
})

describe('UserDialog', () => {
  it('新增模式显示空表单', () => {
    render(<UserDialog open={true} onSubmit={async () => {}} onClose={() => {}} />)
    expect(screen.getByText('新增账号')).toBeDefined()
    expect(screen.getByText('用户名')).toBeDefined()
    expect(screen.getByText('密码')).toBeDefined()
  })
})

describe('ResetPasswordDialog', () => {
  it('显示目标账号名', () => {
    const user: User = { id: 1, username: 'alice', role: 'admin', created_at: '2024-01-01' }
    render(<ResetPasswordDialog open={true} user={user} onSubmit={async () => {}} onClose={() => {}} />)
    expect(screen.getByText('重置密码')).toBeDefined()
    expect(screen.getByText(/alice/)).toBeDefined()
  })
})
