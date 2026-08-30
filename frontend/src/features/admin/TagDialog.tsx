'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { TAG_TYPE_OPTIONS, type Tag, type TagType } from '@/types'

interface TagDialogProps {
  open: boolean
  tag?: Tag | null
  onSubmit: (name: string, type: TagType) => Promise<void>
  onClose: () => void
}

export function TagDialog({ open, tag, onSubmit, onClose }: TagDialogProps) {
  const isEdit = !!tag
  const [name, setName] = useState('')
  const [type, setType] = useState<TagType>('category')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (open) {
      setName(tag?.name || '')
      setType(tag?.type ?? 'category')
    }
  }, [tag, open])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    try {
      await onSubmit(name, type)
      onClose()
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? '编辑标签' : '新增标签'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="tag-name">名称</Label>
            <Input id="tag-name" value={name} onChange={(e) => setName(e.target.value)} disabled={loading} required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="tag-type">类型</Label>
            <Select value={type} onValueChange={(v) => setType(v as TagType)} options={TAG_TYPE_OPTIONS} disabled={loading} />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose} disabled={loading}>
              取消
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? '保存中...' : '保存'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
