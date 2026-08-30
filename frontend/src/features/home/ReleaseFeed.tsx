import Link from 'next/link'
import { Zap } from 'lucide-react'

import type { Release, Repository } from '@/types'
import { sortReleasesNewestFirst } from '@/lib/aggregations'
import { RelativeTime } from '@/components/RelativeTime'

interface ReleaseFeedProps {
  releases: Release[]
  repositories: Repository[]
  limit?: number
}

export function ReleaseFeed({ releases, repositories, limit = 12 }: ReleaseFeedProps) {
  const repoMap = new Map(repositories.map((r) => [r.id, r]))
  const recent = sortReleasesNewestFirst(releases).slice(0, limit)
  const items = recent
    .map((release) => {
      const repo = repoMap.get(release.repository_id)
      return repo ? { release, repo } : null
    })
    .filter((item): item is { release: Release; repo: Repository } => item !== null)

  if (items.length === 0) {
    return null
  }

  const loop = [...items, ...items]

  return (
    <div className="marquee-pause glass relative flex items-center gap-3 overflow-hidden rounded-xl py-2.5 text-left">
      <span className="flex shrink-0 items-center gap-1.5 border-r border-border/50 pl-4 pr-3 text-xs font-bold uppercase tracking-wider text-primary">
        <Zap className="size-3.5" />
        最近更新
      </span>
      <div className="relative flex-1 overflow-hidden">
        <div className="animate-marquee flex w-max items-center gap-6 whitespace-nowrap pr-6">
          {loop.map((item, index) => (
            <Link key={`${item.release.id}-${index}`} href={`/repository/${item.repo.owner}/${item.repo.name}`} className="group flex items-center gap-2 text-sm transition-colors">
              <code className="font-mono font-semibold text-foreground transition-colors group-hover:text-primary">{item.release.tag_name}</code>
              <span className="text-muted-foreground">
                {item.repo.owner}/{item.repo.name}
              </span>
              <span className="text-muted-foreground/40">·</span>
              <RelativeTime iso={item.release.published_at} className="text-xs text-muted-foreground" />
            </Link>
          ))}
        </div>
      </div>
    </div>
  )
}
