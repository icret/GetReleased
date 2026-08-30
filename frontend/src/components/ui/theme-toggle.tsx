'use client'

import { Moon, Sun } from 'lucide-react'

import type { ThemeMode } from '@/hooks/useTheme'
import { cn } from '@/lib/utils'

interface ThemeToggleProps {
  theme: ThemeMode
  onToggle: (mode: ThemeMode) => void
}

export function ThemeToggle({ theme, onToggle }: ThemeToggleProps) {
  const isDark = theme === 'dark'
  const nextMode: ThemeMode = isDark ? 'light' : 'dark'
  const Icon = isDark ? Sun : Moon
  const label = isDark ? '切换至亮色' : '切换至暗色'

  return (
    <button type="button" onClick={() => onToggle(nextMode)} aria-label={label} title={label} className={cn('glass inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs font-medium transition-all', 'hover:text-primary text-foreground')}>
      <Icon className="size-3.5" />
      <span className="hidden sm:inline">{isDark ? '亮色' : '暗色'}</span>
    </button>
  )
}
