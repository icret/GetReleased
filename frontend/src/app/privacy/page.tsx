import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: '隐私政策',
  description: 'GetReleased 隐私政策：说明数据来源及访问信息的处理方式。',
}

const REPO_URL = 'https://github.com/icret/GetReleased'
const GITHUB_API = 'https://docs.github.com/en/rest'
const GITHUB_API_TERMS = 'https://docs.github.com/en/site-policy/github-terms/github-terms-of-service#h-api-terms'
const GITHUB_PRIVACY = 'https://docs.github.com/en/site-policy/privacy-policies/github-privacy-statement'
const GITHUB_TERMS = 'https://docs.github.com/en/site-policy/github-terms/github-terms-of-service'
const EFFECTIVE_DATE = '2026 年 8 月 31 日'

function ExternalLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <a href={href} target="_blank" rel="noopener noreferrer" className="font-medium text-foreground underline decoration-border underline-offset-4 transition hover:text-primary hover:decoration-primary">
      {children}
    </a>
  )
}

export default function PrivacyPage() {
  return (
    <div className="mx-auto max-w-3xl">
      <header className="space-y-3">
        <h1 className="text-3xl font-bold tracking-tight">隐私政策</h1>
        <p className="text-sm text-muted-foreground">生效日期：{EFFECTIVE_DATE}</p>
      </header>

      <div className="mt-10 space-y-10 text-sm leading-7 text-muted-foreground">
        <p>本政策说明 GetReleased（以下简称「本站」）如何处理您的访问信息。继续使用本站即视为您接受本政策。</p>

        <section className="space-y-3">
          <h2 className="text-lg font-semibold tracking-tight text-foreground">数据来源</h2>
          <p>
            本站通过 <ExternalLink href={GITHUB_API}>GitHub REST API</ExternalLink> 获取所追踪仓库的 Release 信息，包括版本号、发布时间、Release Notes 及资产元数据。这些数据均为仓库所有者在 GitHub 上公开发布的内容，版权归原作者所有。本站仅做聚合展示，不修改、不重新授权。
          </p>
          <p>
            本站对 GitHub API 的使用受 <ExternalLink href={GITHUB_API_TERMS}>GitHub API 使用条款</ExternalLink> 约束，并遵循其速率限制与使用规范。
          </p>
        </section>

        <section className="space-y-3">
          <h2 className="text-lg font-semibold tracking-tight text-foreground">信息收集</h2>
          <p>公开页面不收集可识别个人身份的信息，不使用 Cookie，不嵌入第三方统计或广告脚本。</p>
        </section>

        <section className="space-y-3">
          <h2 className="text-lg font-semibold tracking-tight text-foreground">第三方服务</h2>
          <p>
            本站数据全部来自 GitHub。当您从本站跳转至 GitHub 时，相关行为将受 <ExternalLink href={GITHUB_PRIVACY}>GitHub 隐私政策</ExternalLink> 与 <ExternalLink href={GITHUB_TERMS}>GitHub 服务条款</ExternalLink> 约束。GitHub
            作为数据控制者，对其平台上您的行为独立负责，本站不对其隐私处理方式承担责任。
          </p>
        </section>

        <section className="space-y-3">
          <h2 className="text-lg font-semibold tracking-tight text-foreground">数据存储</h2>
          <p>本站仅存储所追踪仓库的公开 Release 元数据，不存储访问者的个人数据，不进行跨站追踪，不构建用户画像。</p>
        </section>

        <section className="space-y-3">
          <h2 className="text-lg font-semibold tracking-tight text-foreground">政策变更</h2>
          <p>本政策可能不时更新。更新后将在本页公布并修订生效日期，重大变更将在仓库 README 中公告。继续使用本站即视为您接受最新版本。</p>
        </section>

        <section className="space-y-3">
          <h2 className="text-lg font-semibold tracking-tight text-foreground">联系我们</h2>
          <p>
            如对本政策有疑问，请通过 <ExternalLink href={REPO_URL}>GitHub 仓库</ExternalLink> 提交 issue。
          </p>
        </section>
      </div>
    </div>
  )
}
