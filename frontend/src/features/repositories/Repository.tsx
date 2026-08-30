'use client'

import { useMemo, useState } from 'react'
import { flushSync } from 'react-dom'
import Link from 'next/link'
import { Archive, ArrowLeft, ExternalLink, GitCommitVertical, Lock, Package, Star } from 'lucide-react'

import type { Release, Repository } from '@/types'
import { sortReleasesNewestFirst, paginate } from '@/lib/aggregations'
import { ReleaseCard } from '@/features/releases/ReleaseCard'
import { Badge } from '@/components/ui/badge'
import { Pagination } from '@/components/Pagination'
import { cn } from '@/lib/utils'

const PAGE_SIZE = 20
const META_PILL = 'inline-flex items-center gap-1.5 rounded-full border border-border/60 bg-muted/40 px-3 py-1 text-xs font-medium text-muted-foreground'

interface RepositoryPageProps {
  repository: Repository
  releases: Release[]
}

export default function Repository({ repository, releases }: RepositoryPageProps) {
  const [currentPage, setCurrentPage] = useState(1)

  const allRepositoryReleases = useMemo(() => sortReleasesNewestFirst(releases), [releases])
  const pagedReleases = useMemo(() => paginate(allRepositoryReleases, currentPage, PAGE_SIZE), [allRepositoryReleases, currentPage])

  const latest = allRepositoryReleases[0]
  const githubUrl = `https://github.com/${repository.full_name}`
  const [expandedId, setExpandedId] = useState<number | null>(latest?.id ?? null)

  function handleJump(releaseId: number) {
    flushSync(() => setExpandedId(releaseId))
    document.getElementById(`release-${releaseId}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  return (
    <div className="space-y-6">
      <Link href="/" className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition hover:text-primary">
        <ArrowLeft className="size-4" />
        返回仓库列表
      </Link>

      {allRepositoryReleases.length === 0 ? (
        <div className="glass rounded-2xl px-6 py-20 text-center">
          <Package className="mx-auto size-12 text-muted-foreground/40" />
          <p className="mt-4 text-sm text-muted-foreground">该仓库暂无 Release 数据。</p>
        </div>
      ) : (
        <div className="flex gap-6">
          <aside className="hidden w-56 shrink-0 lg:block">
            <nav className="glass sticky top-20 max-h-[calc(100vh-7rem)] overflow-y-auto rounded-xl p-3">
              <p className="mb-3 flex items-center gap-1.5 px-1 text-xs font-bold uppercase tracking-wider text-muted-foreground">
                <GitCommitVertical className="size-3.5" />
                版本时间轴
              </p>
              <ul className="space-y-0.5">
                {pagedReleases.map((release) => (
                  <li key={release.id}>
                    <button type="button" onClick={() => handleJump(release.id)} className={cn('w-full rounded-lg px-2.5 py-1.5 text-left transition-colors', expandedId === release.id ? 'bg-primary/10 text-primary' : 'hover:bg-muted/50')}>
                      <span className="block font-mono text-xs font-semibold text-foreground">{release.tag_name}</span>
                      <span className="block text-[11px] text-muted-foreground">{release.published_at.slice(0, 10)}</span>
                    </button>
                  </li>
                ))}
              </ul>
            </nav>
          </aside>

          <div className="min-w-0 flex-1 space-y-4">
            <header className="glass relative overflow-hidden rounded-2xl p-6 sm:p-8">
              <div aria-hidden className="pointer-events-none absolute -right-20 -top-20 h-48 w-48 rounded-full bg-primary/12 blur-[80px]" />
              <div className="relative flex flex-wrap items-end justify-between gap-6">
                <div className="min-w-0">
                  <p className="font-mono text-sm font-medium text-primary">{repository.owner}</p>
                  <h1 className="mt-1.5 font-display text-3xl font-extrabold tracking-tight text-foreground sm:text-4xl">{repository.name}</h1>
                  {repository.description && <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-foreground">{repository.description}</p>}
                  <div className="mt-4 flex flex-wrap items-center gap-2">
                    <span className={META_PILL}>
                      <Package className="size-3.5" />
                      {allRepositoryReleases.length} 个 Release
                    </span>
                    {repository.stars > 0 && (
                      <span className={META_PILL}>
                        <Star className="size-3.5" />
                        {repository.stars}
                      </span>
                    )}
                    {repository.language && <span className={META_PILL}>{repository.language}</span>}
                    {repository.is_archived && (
                      <span className={META_PILL}>
                        <Archive className="size-3.5" />
                        已归档
                      </span>
                    )}
                    {repository.is_private && (
                      <span className={META_PILL}>
                        <Lock className="size-3.5" />
                        私有
                      </span>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2.5">
                  {latest && <Badge variant={latest.is_prerelease ? 'warning' : 'success'}>最新 {latest.tag_name}</Badge>}
                  <a href={githubUrl} target="_blank" rel="noopener noreferrer" className="glass inline-flex items-center gap-1.5 rounded-lg px-3.5 py-2 text-sm font-medium text-muted-foreground transition hover:border-primary/30 hover:text-primary">
                    <ExternalLink className="size-3.5" />
                    GitHub
                  </a>
                </div>
              </div>
            </header>

            <div className="space-y-3">
              {pagedReleases.map((release) => (
                <div key={release.id} id={`release-${release.id}`} className="scroll-mt-20">
                  <ReleaseCard release={release} expanded={expandedId === release.id} onExpandedChange={(isOpen) => setExpandedId(isOpen ? release.id : null)} />
                </div>
              ))}
            </div>

            <Pagination totalItems={allRepositoryReleases.length} itemsPerPage={PAGE_SIZE} currentPage={currentPage} onPageChange={setCurrentPage} />
          </div>
        </div>
      )}
    </div>
  )
}
