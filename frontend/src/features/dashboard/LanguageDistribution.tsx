'use client'

import type { LanguageCount } from '@/types'

interface LanguageDistributionProps {
  languages: LanguageCount[]
}

export function LanguageDistribution({ languages }: LanguageDistributionProps) {
  const total = languages.reduce((sum, l) => sum + l.count, 0)

  return (
    <div className="glass rounded-xl p-6">
      <h2 className="font-display text-lg font-semibold">语言分布</h2>
      <p className="text-sm text-muted-foreground">共 {total} 个仓库</p>
      <div className="mt-6 space-y-3">
        {languages.slice(0, 8).map((l) => {
          const ratio = total > 0 ? (l.count / total) * 100 : 0
          const percent = Math.round(ratio)
          return (
            <div key={l.language} className="space-y-1">
              <div className="flex items-center justify-between text-sm">
                <span className="font-medium">{l.language}</span>
                <span className="tabular-nums text-muted-foreground">
                  {l.count} · {percent}%
                </span>
              </div>
              <div className="h-2 rounded-full bg-muted">
                <div className="h-2 rounded-full bg-primary/80" style={{ width: `${ratio}%` }} />
              </div>
            </div>
          )
        })}
        {languages.length === 0 && <p className="text-sm text-muted-foreground">暂无数据</p>}
      </div>
    </div>
  )
}
