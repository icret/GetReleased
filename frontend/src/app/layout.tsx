import type { Metadata, Viewport } from 'next'
import { Plus_Jakarta_Sans, Inter, JetBrains_Mono } from 'next/font/google'

import './globals.css'
import { SiteHeader } from '@/components/SiteHeader'
import { Footer } from '@/components/Footer'

const displayFont = Plus_Jakarta_Sans({
  subsets: ['latin'],
  weight: ['600', '700', '800'],
  variable: '--font-display-src',
  display: 'swap',
})

const bodyFont = Inter({
  subsets: ['latin'],
  weight: ['400', '500', '600', '700'],
  variable: '--font-body',
  display: 'swap',
})

const monoFont = JetBrains_Mono({
  subsets: ['latin'],
  weight: ['400', '500', '600'],
  variable: '--font-mono-src',
  display: 'swap',
})

const THEME_INIT_SCRIPT = `(function(){var s=localStorage.getItem('getreleased-theme');var m=s==='light'||s==='dark'?s:'dark';document.documentElement.setAttribute('data-theme',m);})()`

export const metadata: Metadata = {
  title: {
    default: 'GetReleased · 追踪开源软件 Release',
    template: '%s — GetReleased',
  },
  description: '聚合追踪仓库的 GitHub Release，版本号、发布时间与 Release Notes 一目了然。',
  icons: { icon: '/favicon.svg' },
  openGraph: {
    type: 'website',
    locale: 'zh_CN',
    siteName: 'GetReleased',
    title: 'GetReleased · 追踪开源软件 Release',
    description: '聚合追踪仓库的 GitHub Release，版本号、发布时间与 Release Notes 一目了然。',
  },
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN" data-theme="dark" suppressHydrationWarning className={`${displayFont.variable} ${bodyFont.variable} ${monoFont.variable}`}>
      <body>
        <script dangerouslySetInnerHTML={{ __html: THEME_INIT_SCRIPT }} />
        <div aria-hidden className="dot-pattern pointer-events-none fixed inset-0 z-0 text-foreground opacity-[0.025]" />
        <div className="relative z-10 flex min-h-screen flex-col bg-background text-foreground">
          <SiteHeader />
          <main className="mx-auto w-full max-w-[1920px] flex-1 px-4 py-10 sm:px-6 sm:pt-14">{children}</main>
          <Footer />
        </div>
      </body>
    </html>
  )
}
