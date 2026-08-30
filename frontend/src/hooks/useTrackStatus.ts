'use client'

import { useEffect, useRef, useState } from 'react'
import { fetchAPI } from '@/lib/api'
import { toast } from '@/components/ui/toast'

export interface TrackStatus {
  running: boolean
  last_task_id?: string
  started_at?: string
  finished_at?: string
  error?: string
}

const POLL_INTERVAL = 2000

export function useTrackStatus(onDone?: () => void) {
  const [status, setStatus] = useState<TrackStatus>({ running: false })
  const [triggering, setTriggering] = useState(false)
  const [exporting, setExporting] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const onDoneRef = useRef(onDone)
  useEffect(() => {
    onDoneRef.current = onDone
  }, [onDone])

  useEffect(() => {
    fetchAPI<TrackStatus>('/api/admin/track/status')
      .then((s) => setStatus(s ?? { running: false }))
      .catch(() => {})
    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [])

  async function postExport() {
    setExporting(true)
    try {
      await fetchAPI('/api/admin/export', { method: 'POST' })
    } finally {
      setExporting(false)
    }
  }

  function startPolling() {
    if (pollRef.current) clearInterval(pollRef.current)
    pollRef.current = setInterval(async () => {
      try {
        const s = await fetchAPI<TrackStatus>('/api/admin/track/status')
        setStatus(s ?? { running: false })
        if (!s.running) {
          if (pollRef.current) clearInterval(pollRef.current)
          pollRef.current = null
          if (s.error) {
            toast.add({ title: '追踪失败', description: s.error })
          } else {
            try {
              await postExport()
              toast.add({ title: '追踪并导出完成' })
            } catch (err) {
              toast.add({ title: '追踪完成，导出失败', description: (err as Error).message })
            }
          }
          onDoneRef.current?.()
        }
      } catch {
        /* ignore poll errors */
      }
    }, POLL_INTERVAL)
  }

  async function trigger() {
    setTriggering(true)
    try {
      await fetchAPI('/api/admin/track', { method: 'POST' })
      setStatus({ running: true })
      startPolling()
    } catch (err) {
      toast.add({ title: '触发失败', description: (err as Error).message })
    } finally {
      setTriggering(false)
    }
  }

  async function exportOnly() {
    try {
      await postExport()
      toast.add({ title: '导出完成' })
    } catch (err) {
      toast.add({ title: '导出失败', description: (err as Error).message })
    }
  }

  return { status, triggering, exporting, trigger, exportOnly }
}
