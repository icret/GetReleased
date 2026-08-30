'use client'

import { useMemo, useState, type ComponentType, type ReactNode } from 'react'
import { Boxes, Flame, GitFork, Package, Search, X } from 'lucide-react'

import type { Release, Repository } from '@/types'
import { latestReleasesByRepository, releaseCountByRepository, sortRepositories, type SortKey } from '@/lib/aggregations'
import { RepositoryCard } from '@/features/repositories/RepositoryCard'
import { ReleaseFeed } from '@/features/home/ReleaseFeed'
import { RelativeTime } from '@/components/RelativeTime'

import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'

function StatChip({ icon: Icon, label, value }: { icon: ComponentType<{ className?: string }>; label: string; value: ReactNode }) {
  return (
    <div className="glass shimmer inline-flex items-center gap-2.5 rounded-lg px-3.5 py-2">
      <span className="relative grid size-6 place-items-center rounded-md bg-primary/12">
        <Icon className="size-3.5 text-primary" />
        <span className="absolute inset-0 rounded-md bg-primary/15 blur-md" />
      </span>
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="text-sm font-semibold tabular-nums text-foreground">{value}</span>
    </div>
  )
}

const sortOptions = [
  { value: 'latest' as const, label: '最近发布' },
  { value: 'count' as const, label: 'Release 数量' },
  { value: 'name' as const, label: '名称 A→Z' },
  { value: 'created' as const, label: '最新追踪' },
]

interface HomeProps {
  repositories: Repository[]
  releases: Release[]
}

export default function Home({ repositories, releases }: HomeProps) {
  const [query, setQuery] = useState('')
  const [sortKey, setSortKey] = useState<SortKey>('latest')
  const [activeTag, setActiveTag] = useState<string | undefined>(undefined)

  const latestByRepo = useMemo(() => latestReleasesByRepository(releases), [releases])
  const countByRepo = useMemo(() => releaseCountByRepository(releases), [releases])
  const newest = useMemo(() => {
    if (releases.length === 0) {
      return null
    }
    return releases.reduce((current, next) => (current.published_at > next.published_at ? current : next))
  }, [releases])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    return repositories.filter((repo) => {
      const matchQuery = !q || repo.owner.toLowerCase().includes(q) || repo.name.toLowerCase().includes(q) || repo.full_name.toLowerCase().includes(q)
      const matchTag = !activeTag || repo.tags?.some((t) => t.name === activeTag)
      return matchQuery && matchTag
    })
  }, [repositories, query, activeTag])

  const sorted = useMemo(() => sortRepositories(filtered, { latestByRepo, countByRepo, sortKey }), [filtered, latestByRepo, countByRepo, sortKey])

  return (
    <div className="space-y-10">
      <section className="relative space-y-6 pt-4 text-center">
        <div aria-hidden className="pointer-events-none absolute inset-x-0 -top-16 mx-auto h-48 max-w-3xl bg-primary/10 blur-[100px]" />
        <h1 className="relative font-display text-4xl font-extrabold leading-tight tracking-tight sm:text-5xl">
          <span className="bg-gradient-to-r from-primary to-[oklch(0.7_0.2_264)] bg-clip-text text-transparent">不错过任何一次</span>
          <span className="text-foreground">开源软件的发布</span>
        </h1>
        <p className="relative mx-auto max-w-4xl text-base leading-7 text-muted-foreground">GetReleased 聚合追踪仓库的 GitHub Release，版本号、发布时间与 Release Notes 一目了然。</p>

        <div className="relative flex flex-wrap justify-center gap-2.5 pt-2">
          <StatChip icon={GitFork} label="追踪仓库" value={String(repositories.length)} />
          <StatChip icon={Package} label="Release 总数" value={String(releases.length)} />
          <StatChip icon={Flame} label="最近发布" value={newest ? <RelativeTime iso={newest.published_at} /> : '—'} />
        </div>

        <ReleaseFeed releases={releases} repositories={repositories} />
      </section>

      {repositories.length > 0 && (
        <section className="space-y-6">
          <div className="glass flex flex-col gap-3 rounded-xl p-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span className="font-display text-base font-bold text-foreground">{sorted.length}</span>
              <span>个仓库</span>
              {activeTag && (
                <button type="button" onClick={() => setActiveTag(undefined)} className="inline-flex items-center gap-1 rounded-full border border-primary/25 bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary transition-opacity hover:opacity-80">
                  标签：{activeTag}
                  <X className="size-3" />
                </button>
              )}
              {query && repositories.length > filtered.length && <span className="text-muted-foreground/60">（已筛选，共 {repositories.length}）</span>}
            </div>
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-2">
              <div className="relative">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="搜索仓库名称或所有者..." className="h-9 w-full border-border/60 bg-muted/40 pl-10 sm:w-64" />
                {query && (
                  <button type="button" onClick={() => setQuery('')} className="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-md p-0.5 text-muted-foreground transition hover:bg-muted hover:text-foreground" aria-label="清除搜索">
                    <X className="size-3.5" />
                  </button>
                )}
              </div>
              <Select value={sortKey} onValueChange={(v) => setSortKey(v as SortKey)} options={sortOptions} placeholder="排序方式" className="h-9 w-full border-border/60 bg-muted/40 sm:w-40" />
            </div>
          </div>

          <div className="flex items-center gap-3">
            <h2 className="font-display text-sm font-bold uppercase tracking-[0.16em] text-muted-foreground">追踪仓库</h2>
            <span className="h-px flex-1 bg-gradient-to-r from-border/60 to-transparent" />
          </div>

          {sorted.length === 0 ? (
            <div className="glass rounded-xl px-6 py-16 text-center">
              <Boxes className="mx-auto size-12 text-muted-foreground/40" />
              <p className="mt-4 text-sm text-muted-foreground">没有匹配的仓库，试试其他关键词</p>
            </div>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {sorted.map((repository) => (
                <RepositoryCard key={repository.id} repository={repository} latest={latestByRepo[repository.id]} releaseCount={countByRepo[repository.id] ?? 0} activeTag={activeTag} onTagClick={(name) => setActiveTag((prev) => (prev === name ? undefined : name))} />
              ))}
            </div>
          )}
        </section>
      )}
    </div>
  )
}
