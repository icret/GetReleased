import Link from 'next/link'
import { ArrowLeft, SearchX } from 'lucide-react'

export default function NotFound() {
  return (
    <div className="glass rounded-xl px-6 py-16 text-center">
      <SearchX className="mx-auto size-12 text-muted-foreground/40" />
      <p className="mt-4 text-sm text-muted-foreground">未找到该页面。</p>
      <Link href="/" className="mt-5 inline-flex items-center gap-1.5 rounded-lg bg-primary/15 px-3 py-2 text-sm text-primary transition hover:bg-primary/25">
        <ArrowLeft className="size-4" />
        返回仓库列表
      </Link>
    </div>
  )
}
