'use client'

import { useState } from 'react'
import { X } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { TAG_TYPE_LABELS, type Tag, type TagType } from '@/types'

interface TagTableProps {
  tags: Tag[]
  onEdit: (tag: Tag) => void
  onDelete: (tag: Tag) => void
}

const GROUP_ORDER: TagType[] = ['platform', 'category']

export function TagTable({ tags, onEdit, onDelete }: TagTableProps) {
  const [query, setQuery] = useState('')
  const groups = GROUP_ORDER.map((type) => ({
    type,
    items: tags.filter((t) => t.type === type),
  })).filter((g) => g.items.length > 0)

  return (
    <div className="space-y-6">
      {groups.map((group) => {
        const filtered = group.type === 'category' ? group.items.filter((t) => t.name.toLowerCase().includes(query.toLowerCase())) : group.items
        return (
          <section key={group.type} className="space-y-3">
            <div className="flex items-center justify-between">
              <h2 className="font-display text-sm font-bold text-muted-foreground">
                {TAG_TYPE_LABELS[group.type]}（{group.items.length}）
              </h2>
              {group.type === 'category' && <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="搜索标签..." className="h-8 w-44 text-sm" />}
            </div>
            <div className="flex flex-wrap gap-2">
              {filtered.map((tag) => (
                <div key={tag.id} className="flex items-center gap-1.5 rounded-full border border-border bg-muted/50 px-2.5 py-1 text-sm transition-colors hover:border-primary/40">
                  <button type="button" onClick={() => onEdit(tag)} className="font-medium hover:text-primary">
                    {tag.name}
                  </button>
                  <button type="button" onClick={() => onDelete(tag)} className="text-muted-foreground transition-colors hover:text-destructive" title="删除">
                    <X className="size-3" />
                  </button>
                </div>
              ))}
              {filtered.length === 0 && <p className="text-sm text-muted-foreground">无匹配标签</p>}
            </div>
          </section>
        )
      })}
    </div>
  )
}
