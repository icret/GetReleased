'use client'

import { useCallback, useEffect, useState } from 'react'

import { fetchAPI } from '@/lib/api'
import type { DashboardStats } from '@/types'

export function useDashboardStats() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await fetchAPI<DashboardStats>('/api/admin/stats')
      setStats(data)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  return { stats, loading, error, reload: load }
}
