'use client'

import { useState, useEffect, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { toast } from '@/components/ui/toast'
import { fetchAPI, fetchAPIList } from '@/lib/api'
import { getCurrentUsername } from '@/lib/auth'
import { UserTable } from '@/features/admin/UserTable'
import { UserDialog } from '@/features/admin/UserDialog'
import { ResetPasswordDialog } from '@/features/admin/ResetPasswordDialog'
import { ConfirmDialog } from '@/features/admin/ConfirmDialog'
import type { User } from '@/types'

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([])
  const [dialogOpen, setDialogOpen] = useState(false)
  const [resetUser, setResetUser] = useState<User | null>(null)
  const [deleteUser, setDeleteUser] = useState<User | null>(null)

  const load = useCallback(async () => {
    try {
      const u = await fetchAPIList<User>('/api/admin/users')
      setUsers(u)
    } catch (err) {
      toast.add({ title: '加载失败', description: (err as Error).message })
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  async function handleCreate(username: string, password: string) {
    try {
      await fetchAPI('/api/admin/users', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      })
      await load()
      toast.add({ title: '创建成功' })
    } catch (err) {
      toast.add({ title: '创建失败', description: (err as Error).message })
      throw err
    }
  }

  async function handleResetPassword(password: string) {
    if (!resetUser) return
    try {
      await fetchAPI(`/api/admin/users/${resetUser.id}/password`, {
        method: 'PUT',
        body: JSON.stringify({ password }),
      })
      toast.add({ title: '重置成功' })
    } catch (err) {
      toast.add({ title: '重置失败', description: (err as Error).message })
      throw err
    }
  }

  async function handleDelete() {
    if (!deleteUser) return
    try {
      await fetchAPI(`/api/admin/users/${deleteUser.id}`, { method: 'DELETE' })
      setDeleteUser(null)
      await load()
      toast.add({ title: '删除成功' })
    } catch (err) {
      toast.add({ title: '删除失败', description: (err as Error).message })
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">账号管理</h1>
        <Button
          onClick={() => {
            setDialogOpen(true)
          }}
        >
          新增账号
        </Button>
      </div>
      <UserTable users={users} currentUsername={getCurrentUsername()} onResetPassword={(user) => setResetUser(user)} onDelete={(user) => setDeleteUser(user)} />
      <UserDialog open={dialogOpen} onSubmit={handleCreate} onClose={() => setDialogOpen(false)} />
      <ResetPasswordDialog open={!!resetUser} user={resetUser} onSubmit={handleResetPassword} onClose={() => setResetUser(null)} />
      <ConfirmDialog open={!!deleteUser} title="删除账号" description={`确认删除账号「${deleteUser?.username}」？该账号将立即无法登录。`} onConfirm={handleDelete} onCancel={() => setDeleteUser(null)} />
    </div>
  )
}
