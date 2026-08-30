'use client'

import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { ChevronDown, Download, ExternalLink, FileArchive, Package } from 'lucide-react'

import type { Release } from '@/types'
import { extractExcerpt } from '@/lib/excerpt'
import { formatBytes } from '@/lib/format'
import { RelativeTime } from '@/components/RelativeTime'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

const REMARK_PLUGINS = [remarkGfm]
const DOWNLOAD_LINK_CLASS = 'inline-flex items-center gap-1.5 rounded-lg border border-border/60 px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:border-primary/30 hover:text-primary'

interface ReleaseCardProps {
  release: Release
  repositoryName?: string
  expanded?: boolean
  onExpandedChange?: (expanded: boolean) => void
}

export function ReleaseCard({ release, repositoryName, expanded: controlledExpanded = false, onExpandedChange }: ReleaseCardProps) {
  const [internalExpanded, setInternalExpanded] = useState(false)
  const expanded = controlledExpanded ?? internalExpanded
  const excerpt = extractExcerpt(release.body)
  const hasDownloads = !!release.tarball_url || !!release.zipball_url || !!release.assets?.length

  function handleToggle() {
    const next = !expanded
    setInternalExpanded(next)
    onExpandedChange?.(next)
  }

  return (
    <article className={cn('group glass overflow-hidden rounded-xl transition-all duration-300 hover:border-primary/25', expanded && 'border-primary/30 glow-primary')}>
      <button type="button" onClick={handleToggle} className="flex w-full flex-col gap-2 px-4 py-3.5 text-left sm:px-5" aria-expanded={expanded}>
        <div className="flex items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-2.5">
            {release.is_prerelease && <Badge variant="warning">预发布</Badge>}
            <h3 className="truncate font-display text-sm font-bold text-foreground">{release.name || release.tag_name}</h3>
          </div>
          <span className="flex shrink-0 items-center gap-2 text-xs text-muted-foreground">
            <RelativeTime iso={release.published_at} />
            <ChevronDown className={cn('size-4 transition-transform duration-300', expanded && 'rotate-180 text-primary')} />
          </span>
        </div>

        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <code className="rounded-md border border-primary/20 bg-primary/8 px-1.5 py-0.5 font-mono font-medium text-primary">{release.tag_name}</code>
          {repositoryName && (
            <>
              <span className="text-border">/</span>
              <span className="truncate">{repositoryName}</span>
            </>
          )}
        </div>

        <p className="line-clamp-2 text-sm leading-6 text-muted-foreground">{excerpt}</p>
      </button>

      {expanded && (
        <div className="flex flex-col gap-4 border-t border-border/40 px-4 py-4 sm:px-5">
          {release.body ? (
            <div className="release-markdown">
              <ReactMarkdown remarkPlugins={REMARK_PLUGINS}>{release.body}</ReactMarkdown>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">暂无 Release Notes</p>
          )}
          {hasDownloads && (
            <div className="space-y-2">
              <p className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider text-muted-foreground">
                <Download className="size-3.5" />
                下载
              </p>
              <div className="flex flex-wrap gap-2">
                {release.tarball_url && (
                  <a href={release.tarball_url} target="_blank" rel="noopener noreferrer" className={DOWNLOAD_LINK_CLASS}>
                    <FileArchive className="size-3.5" />
                    Source code (tar.gz)
                  </a>
                )}
                {release.zipball_url && (
                  <a href={release.zipball_url} target="_blank" rel="noopener noreferrer" className={DOWNLOAD_LINK_CLASS}>
                    <FileArchive className="size-3.5" />
                    Source code (zip)
                  </a>
                )}
                {release.assets?.map((asset) => (
                  <a key={asset.id} href={asset.download_url} target="_blank" rel="noopener noreferrer" className={DOWNLOAD_LINK_CLASS}>
                    <Package className="size-3.5" />
                    {asset.name}
                    <span className="text-muted-foreground/70">{formatBytes(asset.size)}</span>
                  </a>
                ))}
              </div>
            </div>
          )}
          <a href={release.html_url} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-1.5 self-start rounded-lg border border-border/60 px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:border-primary/30 hover:text-primary">
            <ExternalLink className="size-3.5" />在 GitHub 查看该版本
          </a>
        </div>
      )}
    </article>
  )
}
