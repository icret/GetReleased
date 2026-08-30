'use client'

import { RefreshCw } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { LanguageDistribution } from '@/features/dashboard/LanguageDistribution'
import { RecentReleasesTable } from '@/features/dashboard/RecentReleasesTable'
import { ReleaseTrendChart } from '@/features/dashboard/ReleaseTrendChart'
import { StatsCards } from '@/features/dashboard/StatsCards'
import { TopRepositories } from '@/features/dashboard/TopRepositories'
import { TrackCard } from '@/features/dashboard/TrackCard'
import { useDashboardStats } from '@/hooks/useDashboardStats'

export default function AdminDashboard() {
  const { stats, loading, error, reload } = useDashboardStats()

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-2xl font-bold tracking-tight">仪表盘</h1>
        <Button variant="ghost" size="sm" onClick={reload} disabled={loading}>
          <RefreshCw className={loading ? 'animate-spin' : ''} />
          刷新
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">加载失败: {error}</p>}

      {!stats ? (
        <div className="space-y-6">
          <Skeleton className="h-28 w-full" />
          <Skeleton className="h-80 w-full" />
          <Skeleton className="h-80 w-full" />
        </div>
      ) : (
        <>
          <StatsCards overview={stats.overview} tagTypes={stats.tag_types} />

          <div className="grid gap-6 lg:grid-cols-3">
            <div className="lg:col-span-2">
              <ReleaseTrendChart trend={stats.release_trend} />
            </div>
            <LanguageDistribution languages={stats.languages} />
          </div>

          <div className="grid gap-6 lg:grid-cols-3">
            <div className="lg:col-span-2">
              <RecentReleasesTable releases={stats.recent_releases} />
            </div>
            <TopRepositories repos={stats.top_repositories} />
          </div>

          <TrackCard onTracked={reload} />
        </>
      )}
    </div>
  )
}
