'use client'

import { useEffect, useState } from 'react'

import { formatDate, formatRelativeTime } from '@/lib/format'

interface RelativeTimeProps {
  iso: string
  className?: string
}

export function RelativeTime({ iso, className }: RelativeTimeProps) {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  return (
    <time dateTime={iso} title={formatDate(iso)} className={className}>
      {mounted ? formatRelativeTime(iso) : formatDate(iso)}
    </time>
  )
}
