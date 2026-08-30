'use client'

import { useState, useEffect, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { toast } from '@/components/ui/toast'
import { fetchAPI, fetchAPIList } from '@/lib/api'
import { TagTable } from '@/features/admin/TagTable'
import { TagDialog } from '@/features/admin/TagDialog'
import { ConfirmDialog } from '@/features/admin/ConfirmDialog'
import type { Tag, TagType } from '@/types'

export default function TagsPage() {
  const [tags, setTags] = useState<Tag[]>([])
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingTag, setEditingTag] = useState<Tag | null>(null)
  const [deleteTag, setDeleteTag] = useState<Tag | null>(null)

  const load = useCallback(async () => {
    try {
      const t = await fetchAPIList<Tag>('/api/admin/tags')
      setTags(t)
    } catch (err) {
      toast.add({ title: '加载失败', description: (err as Error).message })
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  async function handleSubmit(name: string, type: TagType) {
    try {
      if (editingTag) {
        await fetchAPI(`/api/admin/tags/${editingTag.id}`, {
          method: 'PUT',
          body: JSON.stringify({ name, type }),
        })
      } else {
        await fetchAPI('/api/admin/tags', {
          method: 'POST',
          body: JSON.stringify({ name, type }),
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
    if (!deleteTag) return
    try {
      await fetchAPI(`/api/admin/tags/${deleteTag.id}`, { method: 'DELETE' })
      setDeleteTag(null)
      await load()
      toast.add({ title: '删除成功' })
    } catch (err) {
      toast.add({ title: '删除失败', description: (err as Error).message })
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">标签管理</h1>
        <Button
          onClick={() => {
            setEditingTag(null)
            setDialogOpen(true)
          }}
        >
          新增标签
        </Button>
      </div>
      <TagTable
        tags={tags}
        onEdit={(tag) => {
          setEditingTag(tag)
          setDialogOpen(true)
        }}
        onDelete={(tag) => setDeleteTag(tag)}
      />
      <TagDialog open={dialogOpen} tag={editingTag} onSubmit={handleSubmit} onClose={() => setDialogOpen(false)} />
      <ConfirmDialog open={!!deleteTag} title="删除标签" description={`确认删除标签「${deleteTag?.name}」？`} onConfirm={handleDelete} onCancel={() => setDeleteTag(null)} />
    </div>
  )
}
