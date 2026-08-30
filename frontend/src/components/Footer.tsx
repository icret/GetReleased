import { Rocket } from 'lucide-react'

const REPO_URL = 'https://github.com/icret/GetReleased'

export function Footer() {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="relative mt-16 border-t border-border/60">
      <div className="mx-auto flex max-w-[1920px] flex-col items-center justify-between gap-3 px-4 py-6 text-sm text-muted-foreground sm:flex-row sm:px-6">
        <div className="flex items-center gap-2">
          <span className="grid size-6 place-items-center rounded-md accent-gradient text-white">
            <Rocket className="size-3" />
          </span>
          <span className="font-medium text-foreground">GetReleased</span>
          <span className="text-muted-foreground/60">·</span>
          <span>追踪开源软件 Release</span>
        </div>
        <div className="flex items-center gap-4 text-xs">
          <span>
            &copy; {currentYear}{' '}
            <a href={REPO_URL} target="_blank" rel="noopener noreferrer" className="font-medium text-foreground transition hover:text-primary">
              GetReleased
            </a>
          </span>
          <span className="text-muted-foreground/60">·</span>
          <span>数据来自 GitHub API</span>
          <span className="text-muted-foreground/60">·</span>
          <a href="/privacy" className="transition hover:text-primary">
            隐私政策
          </a>
        </div>
      </div>
    </footer>
  )
}
