'use client'

import Link from 'next/link'
import { ExternalLink, Package, Star, Tag } from 'lucide-react'
import { useState } from 'react'

import type { Repository } from '@/types'
import { gradientFor } from '@/lib/gradient'
import { RelativeTime } from '@/components/RelativeTime'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface RepositoryCardProps {
  repository: Repository
  activeTag?: string
  onTagClick?: (tagName: string) => void
}

export function RepositoryCard({ repository, activeTag, onTagClick }: RepositoryCardProps) {
  const githubUrl = `https://github.com/${repository.full_name}`
  const fallback = gradientFor(repository.name)
  const initial = repository.name.charAt(0).toUpperCase()
  const [imgError, setImgError] = useState(false)
  const avatarSrc = repository.logo_path ? `/${repository.logo_path}` : `https://avatars.githubusercontent.com/${repository.owner}`

  return (
    <div className="group glass relative flex h-full rounded-xl transition-all duration-300 hover:-translate-y-1 hover:border-primary/30 hover:shadow-[0_8px_24px_-8px_oklch(0.55_0.25_264/0.25)]">
      <Link
        href={`/repository/${repository.owner}/${repository.name}`}
        className="flex w-full flex-col gap-3.5 p-4"
        onClick={(e) => {
          if ((e.target as HTMLElement).closest('[data-tag-click]')) {
            e.preventDefault()
          }
        }}
      >
        <div className="flex items-center gap-3">
          <span className={cn('relative grid size-10 shrink-0 place-items-center overflow-hidden rounded-lg text-sm font-semibold text-white', imgError ? `bg-gradient-to-br ${fallback}` : 'bg-gradient-to-br')}>
            {!imgError && <img src={avatarSrc} alt={`${repository.owner} 的头像`} className="relative z-10 size-full rounded-lg object-cover" onError={() => setImgError(true)} />}
            {imgError && <span className="relative z-10">{initial}</span>}
            <span aria-hidden className={cn('absolute inset-0 rounded-lg bg-gradient-to-br opacity-0 blur-md transition-opacity duration-300 group-hover:opacity-60', fallback)} />
          </span>
          <div className="min-w-0 flex-1">
            <h3 className="truncate font-display text-base font-bold text-foreground transition-colors group-hover:text-primary">{repository.name}</h3>
            <p className="truncate text-xs text-muted-foreground">{repository.owner}</p>
          </div>
        </div>

        {repository.description ? (
          <p className="line-clamp-2 text-xs leading-5 text-muted-foreground" title={repository.description}>
            {repository.description}
          </p>
        ) : (
          <div className="min-h-[2.5rem]" />
        )}

        <div className="mt-auto space-y-2.5">
          <div className="flex flex-wrap items-center gap-1.5 border-t border-border/40 pt-2.5">
            {repository.latest_version ? (
              <Badge variant={repository.latest_is_prerelease ? 'warning' : 'success'}>
                <Tag />
                {repository.latest_version}
              </Badge>
            ) : (
              <span className="text-xs text-muted-foreground">暂无 Release</span>
            )}
            {repository.tags?.map((tag) => {
              const isActive = activeTag === tag.name
              return (
                <Badge
                  key={tag.id}
                  variant={isActive ? 'default' : 'secondary'}
                  className={onTagClick ? 'cursor-pointer transition-opacity hover:opacity-80' : undefined}
                  role={onTagClick ? 'button' : undefined}
                  tabIndex={onTagClick ? 0 : undefined}
                  data-tag-click={onTagClick ? '' : undefined}
                  onClick={onTagClick ? () => onTagClick(tag.name) : undefined}
                >
                  {tag.name}
                </Badge>
              )
            })}
          </div>

          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span className="inline-flex items-center gap-3">
              <span className="inline-flex items-center gap-1.5">
                <Package className="size-3.5" />
                {repository.release_count ?? 0} 个 Release
              </span>
              {repository.stars > 0 && (
                <span className="inline-flex items-center gap-1">
                  <Star className="size-3.5" />
                  {repository.stars}
                </span>
              )}
              {repository.language && <span className="truncate">{repository.language}</span>}
            </span>
            {repository.latest_release_date ? <RelativeTime iso={repository.latest_release_date} /> : <span>—</span>}
          </div>
        </div>
      </Link>

      <a href={githubUrl} target="_blank" rel="noopener noreferrer" className="absolute right-3 top-3 z-10 rounded-lg p-1.5 text-muted-foreground opacity-50 transition hover:bg-muted hover:text-primary hover:opacity-100" aria-label={`在 GitHub 打开 ${repository.full_name}`}>
        <ExternalLink className="size-3.5" />
      </a>
    </div>
  )
}
