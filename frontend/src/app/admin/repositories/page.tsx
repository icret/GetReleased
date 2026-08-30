'use client'

import { useState, useEffect, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { toast } from '@/components/ui/toast'
import { fetchAPI, fetchAPIList } from '@/lib/api'
import { RepositoryTable } from '@/features/admin/RepositoryTable'
import { RepositoryDialog } from '@/features/admin/RepositoryDialog'
import { ConfirmDialog } from '@/features/admin/ConfirmDialog'
import type { Repository, Tag } from '@/types'

export default function RepositoriesPage() {
  const [repos, setRepos] = useState<Repository[]>([])
  const [tags, setTags] = useState<Tag[]>([])
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingRepo, setEditingRepo] = useState<Repository | null>(null)
  const [deleteRepo, setDeleteRepo] = useState<Repository | null>(null)
  const [syncingId, setSyncingId] = useState<number | null>(null)

  const load = useCallback(async () => {
    try {
      const [r, t] = await Promise.all([fetchAPIList<Repository>('/api/admin/repositories'), fetchAPIList<Tag>('/api/admin/tags')])
      setRepos(r)
      setTags(t)
    } catch (err) {
      toast.add({ title: '加载失败', description: (err as Error).message })
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  async function handleSync(repo: Repository) {
    if (syncingId !== null) return
    setSyncingId(repo.id)
    toast.add({ title: `正在同步 ${repo.full_name}...` })
    try {
      await fetchAPI(`/api/admin/repositories/${repo.id}/sync`, { method: 'POST' })
      await load()
      toast.add({ title: '同步成功' })
    } catch (err) {
      toast.add({ title: '同步失败', description: (err as Error).message })
    } finally {
      setSyncingId(null)
    }
  }

  async function handleSubmit(data: { owner?: string; name?: string; description: string; remark: string; tagIds: number[] }) {
    try {
      if (editingRepo) {
        await fetchAPI(`/api/admin/repositories/${editingRepo.id}`, {
          method: 'PUT',
          body: JSON.stringify({ description: data.description, remark: data.remark }),
        })
        await fetchAPI(`/api/admin/repositories/${editingRepo.id}/tags`, {
          method: 'PUT',
          body: JSON.stringify({ tag_ids: data.tagIds }),
        })
      } else {
        const created = await fetchAPI<Repository>('/api/admin/repositories', {
          method: 'POST',
          body: JSON.stringify({ owner: data.owner, name: data.name }),
        })
        await fetchAPI(`/api/admin/repositories/${created.id}/tags`, {
          method: 'PUT',
          body: JSON.stringify({ tag_ids: data.tagIds }),
        })
      }
      await load()
      toast.add({ title: '保存成功' })
    } catch (err) {
      toast.add({ title: '保存失败', description: (err as Error).message })
      throw err
    }
  }

  async function handleDelete() {
    if (!deleteRepo) return
    try {
      await fetchAPI(`/api/admin/repositories/${deleteRepo.id}`, { method: 'DELETE' })
      setDeleteRepo(null)
      await load()
      toast.add({ title: '删除成功' })
    } catch (err) {
      toast.add({ title: '删除失败', description: (err as Error).message })
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">仓库管理</h1>
        <Button
          onClick={() => {
            setEditingRepo(null)
            setDialogOpen(true)
          }}
        >
          新增仓库
        </Button>
      </div>
      <RepositoryTable
        repos={repos}
        syncingId={syncingId}
        onSync={handleSync}
        onEdit={(repo) => {
          setEditingRepo(repo)
          setDialogOpen(true)
        }}
        onDelete={(repo) => setDeleteRepo(repo)}
      />
      <RepositoryDialog open={dialogOpen} repo={editingRepo} tags={tags} onSubmit={handleSubmit} onClose={() => setDialogOpen(false)} />
      <ConfirmDialog open={!!deleteRepo} title="删除仓库" description={`确认删除 ${deleteRepo?.full_name}？关联的 Release 数据将一并删除。`} onConfirm={handleDelete} onCancel={() => setDeleteRepo(null)} />
    </div>
  )
}
