'use client'

import { useEffect, useState } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import Link from 'next/link'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Toaster } from '@/components/ui/toast'
import { getToken, logout } from '@/lib/auth'
import { LayoutDashboard, GitFork, Tag, Users, LogOut } from 'lucide-react'

const NAV_ITEMS = [
  { href: '/admin', label: '仪表盘', icon: LayoutDashboard },
  { href: '/admin/repositories', label: '仓库', icon: GitFork },
  { href: '/admin/tags', label: '标签', icon: Tag },
  { href: '/admin/users', label: '账号', icon: Users },
]

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const [authed, setAuthed] = useState(false)
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    if (!getToken()) {
      router.replace('/login')
      return
    }
    setAuthed(true)
  }, [router])

  if (!authed) {
    return <Skeleton className="h-screen w-full" />
  }

  return (
    <Toaster>
      <div className="flex gap-6">
        <aside className="glass sticky top-20 w-48 shrink-0 space-y-1 self-start rounded-xl p-3">
          {NAV_ITEMS.map((item) => {
            const active = pathname === item.href
            return (
              <Link key={item.href} href={item.href}>
                <Button variant={active ? 'default' : 'ghost'} className="w-full justify-start">
                  <item.icon />
                  {item.label}
                </Button>
              </Link>
            )
          })}
          <Button variant="ghost" className="w-full justify-start" onClick={logout}>
            <LogOut />
            登出
          </Button>
        </aside>
        <div className="flex-1">{children}</div>
      </div>
    </Toaster>
  )
}
