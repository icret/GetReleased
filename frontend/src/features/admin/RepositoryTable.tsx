'use client'

import { Pencil, Trash2, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatRelativeTime } from '@/lib/format'
import type { Repository } from '@/types'

interface RepositoryTableProps {
  repos: Repository[]
  syncingId?: number | null
  onSync: (repo: Repository) => void
  onEdit: (repo: Repository) => void
  onDelete: (repo: Repository) => void
}

export function RepositoryTable({ repos, syncingId, onSync, onEdit, onDelete }: RepositoryTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>仓库信息</TableHead>
          <TableHead>标签</TableHead>
          <TableHead>版本</TableHead>
          <TableHead>更新状态</TableHead>
          <TableHead>备注</TableHead>
          <TableHead className="w-[120px]">管理操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {repos.map((repo) => {
          const syncing = syncingId === repo.id
          return (
            <TableRow key={repo.id}>
              <TableCell>
                <div className="flex flex-col">
                  <span className="font-medium">{repo.full_name}</span>
                  {repo.description ? <span className="max-w-xs truncate text-xs text-muted-foreground">{repo.description}</span> : null}
                </div>
              </TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  {repo.tags?.map((tag) => (
                    <Badge key={tag.id} variant="secondary">
                      {tag.name}
                    </Badge>
                  ))}
                </div>
              </TableCell>
              <TableCell>
                {repo.latest_version ? (
                  repo.latest_release_url ? (
                    <a href={repo.latest_release_url} target="_blank" rel="noopener noreferrer" className="text-sm font-medium text-primary underline-offset-2 hover:underline">
                      {repo.latest_version}
                    </a>
                  ) : (
                    <span className="text-sm font-medium">{repo.latest_version}</span>
                  )
                ) : (
                  <span className="text-muted-foreground">-</span>
                )}
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">{repo.last_checked_at ? formatRelativeTime(repo.last_checked_at) : '未同步'}</TableCell>
              <TableCell className="max-w-xs truncate text-sm">{repo.remark || <span className="text-muted-foreground">-</span>}</TableCell>
              <TableCell>
                <div className="flex items-center gap-1">
                  <Button variant="ghost" size="icon" title="同步" disabled={syncing} onClick={() => onSync(repo)}>
                    <RefreshCw className={syncing ? 'animate-spin' : ''} />
                  </Button>
                  <Button variant="ghost" size="icon" title="修改" disabled={syncing} onClick={() => onEdit(repo)}>
                    <Pencil />
                  </Button>
                  <Button variant="ghost" size="icon" title="删除" disabled={syncing} onClick={() => onDelete(repo)}>
                    <Trash2 />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
