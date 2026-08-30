'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { parseRepoInput } from '@/lib/repo-url'
import type { Repository, Tag } from '@/types'

interface RepositoryDialogProps {
  open: boolean
  repo?: Repository | null
  tags: Tag[]
  onSubmit: (data: { owner?: string; name?: string; description: string; remark: string; tagIds: number[] }) => Promise<void>
  onClose: () => void
}

export function RepositoryDialog({ open, repo, tags, onSubmit, onClose }: RepositoryDialogProps) {
  const isEdit = !!repo
  const [repoUrl, setRepoUrl] = useState('')
  const [description, setDescription] = useState('')
  const [remark, setRemark] = useState('')
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [loading, setLoading] = useState(false)

  const platformTags = tags.filter((t) => t.type === 'platform')
  const categoryTags = tags.filter((t) => t.type === 'category')

  useEffect(() => {
    if (open) {
      if (repo) {
        setRepoUrl(`${repo.owner}/${repo.name}`)
        setDescription(repo.description || '')
        setRemark(repo.remark || '')
        setSelectedIds(repo.tags?.map((t) => t.id) ?? [])
      } else {
        setRepoUrl('')
        setDescription('')
        setRemark('')
        setSelectedIds([])
      }
    }
  }, [repo, open])

  function toggle(id: number) {
    setSelectedIds((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))
  }

  const parsedRepo = parseRepoInput(repoUrl)
  const showPreview = repoUrl.trim().length > 0 && !isEdit
  const canSubmit = isEdit || !!parsedRepo

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!isEdit) {
      const parsed = parseRepoInput(repoUrl)
      if (!parsed) return
      setLoading(true)
      try {
        await onSubmit({ owner: parsed.owner, name: parsed.name, description, remark, tagIds: selectedIds })
        onClose()
      } finally {
        setLoading(false)
      }
      return
    }
    setLoading(true)
    try {
      await onSubmit({ description, remark, tagIds: selectedIds })
      onClose()
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? '编辑仓库' : '新增仓库'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="repo-url">仓库地址</Label>
            <Input id="repo-url" value={repoUrl} onChange={(e) => setRepoUrl(e.target.value)} disabled={isEdit || loading} placeholder="https://github.com/owner/repo 或 owner/repo" />
            {showPreview && parsedRepo && (
              <p className="text-xs text-muted-foreground">
                识别为 {parsedRepo.owner}/{parsedRepo.name}
              </p>
            )}
            {showPreview && !parsedRepo && <p className="text-xs text-destructive">无法识别，请输入 owner/repo 或 GitHub URL</p>}
          </div>
          <div className="space-y-2">
            <Label htmlFor="repo-description">描述</Label>
            <Textarea id="repo-description" value={description} onChange={(e) => setDescription(e.target.value)} disabled={loading} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="repo-remark">备注</Label>
            <Textarea id="repo-remark" value={remark} onChange={(e) => setRemark(e.target.value)} disabled={loading} placeholder="管理员备注（可选）" />
          </div>
          <div className="space-y-2">
            <Label>分类</Label>
            <div className="flex flex-wrap gap-1.5">
              {platformTags.map((t) => {
                const active = selectedIds.includes(t.id)
                return (
                  <Badge key={t.id} variant={active ? 'default' : 'secondary'} className="cursor-pointer select-none" role="button" tabIndex={0} onClick={() => !loading && toggle(t.id)}>
                    {t.name}
                  </Badge>
                )
              })}
            </div>
          </div>
          <div className="space-y-2">
            <Label>标签</Label>
            <div className="flex flex-wrap gap-1.5">
              {categoryTags.map((t) => {
                const active = selectedIds.includes(t.id)
                return (
                  <Badge key={t.id} variant={active ? 'default' : 'secondary'} className="cursor-pointer select-none" role="button" tabIndex={0} onClick={() => !loading && toggle(t.id)}>
                    {t.name}
                  </Badge>
                )
              })}
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose} disabled={loading}>
              取消
            </Button>
            <Button type="submit" disabled={loading || !canSubmit}>
              {loading ? '处理中...' : '保存'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
