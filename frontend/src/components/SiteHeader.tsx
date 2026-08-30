'use client'

import Link from 'next/link'
import { Rocket } from 'lucide-react'

import { useTheme } from '@/hooks/useTheme'
import { ThemeToggle } from '@/components/ui/theme-toggle'

export function SiteHeader() {
  const [theme, setTheme] = useTheme()

  return (
    <header className="glass-strong sticky top-0 z-50">
      <div className="mx-auto flex h-16 max-w-[1920px] items-center justify-between gap-4 px-4 sm:px-6">
        <Link href="/" className="group flex items-center gap-2.5">
          <span className="relative grid size-8 place-items-center overflow-hidden rounded-lg accent-gradient text-white shadow-[0_4px_14px_-2px_oklch(0.55_0.25_264/0.4)]">
            <Rocket className="relative z-10 size-4" />
            <span aria-hidden className="absolute inset-0 bg-white/20 opacity-0 transition-opacity duration-300 group-hover:opacity-100" />
          </span>
          <span className="flex flex-col leading-none">
            <span className="font-display text-base font-bold tracking-tight text-foreground">GetReleased</span>
            <span className="mt-0.5 text-[10px] font-medium uppercase tracking-[0.18em] text-muted-foreground">追踪开源发布</span>
          </span>
        </Link>
        <ThemeToggle theme={theme} onToggle={setTheme} />
      </div>
    </header>
  )
}
