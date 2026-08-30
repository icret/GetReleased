'use client'

import type { ReleaseTrendPoint } from '@/types'

interface ReleaseTrendChartProps {
  trend: ReleaseTrendPoint[]
}

export function ReleaseTrendChart({ trend }: ReleaseTrendChartProps) {
  const max = Math.max(1, ...trend.map((p) => p.count))
  const total = trend.reduce((sum, p) => sum + p.count, 0)

  return (
    <div className="glass h-full rounded-xl p-6">
      <div className="flex items-baseline justify-between">
        <div>
          <h2 className="font-display text-lg font-semibold">发布趋势</h2>
          <p className="text-sm text-muted-foreground">最近 12 个月</p>
        </div>
        <span className="text-sm text-muted-foreground">共 {total} 次</span>
      </div>
      <div className="mt-6 flex gap-2" style={{ height: 160 }}>
        {trend.map((p) => {
          const h = Math.round((p.count / max) * 100)
          return (
            <div key={p.month} className="group flex flex-1 flex-col items-center gap-2">
              <div className="relative flex w-full flex-1 items-end">
                <div className="w-full rounded-t-md bg-primary/70 transition-all group-hover:bg-primary" style={{ height: `${Math.max(h, p.count > 0 ? 4 : 0)}%` }} title={`${p.month}: ${p.count} 次`} />
              </div>
              <span className="text-[10px] text-muted-foreground">{p.month.slice(5)}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
