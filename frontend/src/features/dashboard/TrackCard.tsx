'use client'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { RelativeTime } from '@/components/RelativeTime'
import { useTrackStatus } from '@/hooks/useTrackStatus'

interface TrackCardProps {
  onTracked?: () => void
}

export function TrackCard({ onTracked }: TrackCardProps) {
  const { status, triggering, exporting, trigger, exportOnly } = useTrackStatus(onTracked)
  const busy = triggering || exporting || status.running

  return (
    <div className="glass flex flex-col gap-4 rounded-xl p-6 lg:flex-row lg:items-center lg:justify-between">
      <div className="space-y-1">
        <div className="flex items-center gap-2">
          <h2 className="font-display text-lg font-semibold">追踪操作</h2>
          <Badge variant={status.running ? 'default' : 'secondary'}>{status.running ? '追踪中' : '空闲'}</Badge>
        </div>
        <p className="text-sm text-muted-foreground">手动追踪 GitHub Release 并导出 JSON，或仅导出当前数据</p>
        {status.error && <p className="text-sm text-destructive">错误: {status.error}</p>}
        {status.finished_at && !status.running && (
          <p className="text-sm text-muted-foreground">
            上次完成: <RelativeTime iso={status.finished_at} />
          </p>
        )}
      </div>
      <div className="flex gap-2">
        <Button variant="outline" onClick={exportOnly} disabled={busy}>
          {exporting ? '导出中...' : '仅导出 JSON'}
        </Button>
        <Button onClick={trigger} disabled={busy}>
          {status.running ? '追踪中...' : exporting ? '导出中...' : '立即追踪并导出'}
        </Button>
      </div>
    </div>
  )
}
