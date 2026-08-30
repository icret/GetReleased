'use client'

import { Star } from 'lucide-react'

import type { TopRepository } from '@/types'

interface TopRepositoriesProps {
  repos: TopRepository[]
}

export function TopRepositories({ repos }: TopRepositoriesProps) {
  return (
    <div className="glass rounded-xl p-6">
      <h2 className="font-display text-lg font-semibold">Star 排行</h2>
      <p className="text-sm text-muted-foreground">Top {repos.length} 仓库</p>
      <div className="mt-6 space-y-3">
        {repos.map((repo, i) => (
          <div key={repo.id} className="flex items-center gap-3">
            <span className="font-display w-5 shrink-0 text-sm font-bold text-muted-foreground">{i + 1}</span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{repo.full_name}</p>
              {repo.latest_version &&
                (repo.latest_release_url ? (
                  <a href={repo.latest_release_url} target="_blank" rel="noopener noreferrer" className="truncate text-xs text-primary hover:underline">
                    {repo.latest_version}
                  </a>
                ) : (
                  <p className="truncate text-xs text-muted-foreground">{repo.latest_version}</p>
                ))}
            </div>
            <div className="flex shrink-0 items-center gap-1 text-sm text-muted-foreground">
              <Star className="h-3.5 w-3.5" />
              <span className="tabular-nums">{repo.stars}</span>
            </div>
          </div>
        ))}
        {repos.length === 0 && <p className="text-sm text-muted-foreground">暂无数据</p>}
      </div>
    </div>
  )
}
