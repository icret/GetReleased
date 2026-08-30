'use client'

import { Badge } from '@/components/ui/badge'
import { RelativeTime } from '@/components/RelativeTime'
import type { RecentRelease } from '@/types'

interface RecentReleasesTableProps {
  releases: RecentRelease[]
}

export function RecentReleasesTable({ releases }: RecentReleasesTableProps) {
  return (
    <div className="glass rounded-xl p-6">
      <h2 className="font-display text-lg font-semibold">最近更新</h2>
      <p className="text-sm text-muted-foreground">最新发布的 Release</p>
      <div className="mt-6 space-y-3">
        {releases.map((rel) => (
          <div key={rel.id} className="flex items-center gap-3">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <span className="truncate text-sm font-medium">{rel.full_name}</span>
                {rel.is_prerelease && (
                  <Badge variant="secondary" className="shrink-0">
                    预发布
                  </Badge>
                )}
              </div>
              {rel.html_url ? (
                <a href={rel.html_url} target="_blank" rel="noopener noreferrer" className="text-xs text-primary hover:underline">
                  {rel.tag_name}
                </a>
              ) : (
                <span className="text-xs text-muted-foreground">{rel.tag_name}</span>
              )}
            </div>
            <RelativeTime iso={rel.published_at} className="shrink-0 text-xs text-muted-foreground" />
          </div>
        ))}
        {releases.length === 0 && <p className="text-sm text-muted-foreground">暂无数据</p>}
      </div>
    </div>
  )
}
