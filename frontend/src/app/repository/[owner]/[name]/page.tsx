import type { Metadata } from 'next'
import { notFound } from 'next/navigation'

import { loadReleasesByRepository, loadRepositories } from '@/data/load'
import Repository from '@/features/repositories/Repository'

export async function generateStaticParams() {
  const repositories = loadRepositories()
  return repositories.map((repo) => ({ owner: repo.owner, name: repo.name }))
}

export async function generateMetadata({ params }: { params: Promise<{ owner: string; name: string }> }): Promise<Metadata> {
  const { owner, name } = await params
  const repositories = loadRepositories()
  const repository = repositories.find((repo) => repo.owner === owner && repo.name === name)
  if (!repository) {
    return { title: '未找到仓库', robots: { index: false, follow: false } }
  }
  return {
    title: `${owner}/${name} · Releases`,
    description: repository.description || `${owner}/${name} 最新 Release 与版本更新一览`,
    alternates: { canonical: `/repository/${owner}/${name}` },
  }
}

export default async function RepositoryPage({ params }: { params: Promise<{ owner: string; name: string }> }) {
  const { owner, name } = await params
  const repositories = loadRepositories()
  const repository = repositories.find((repo) => repo.owner === owner && repo.name === name)
  if (!repository) {
    notFound()
  }
  const releases = loadReleasesByRepository(repository.id)
  return <Repository repository={repository} releases={releases} />
}
