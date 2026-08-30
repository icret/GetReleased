import type { Release, Repository } from '@/types'

export type SortKey = 'latest' | 'count' | 'name' | 'created'

export interface SortRepositoriesOptions {
  latestByRepo: Record<number, Release | undefined>
  countByRepo: Record<number, number>
  sortKey: SortKey
}

export function sortReleasesNewestFirst(releases: Release[]): Release[] {
  return releases.slice().sort((a, b) => b.published_at.localeCompare(a.published_at))
}

export function sortRepositories(repositories: Repository[], options: SortRepositoriesOptions): Repository[] {
  const { latestByRepo, countByRepo, sortKey } = options
  const sorted = repositories.slice()
  switch (sortKey) {
    case 'latest':
      sorted.sort((a, b) => {
        const aTime = latestByRepo[a.id]?.published_at ?? ''
        const bTime = latestByRepo[b.id]?.published_at ?? ''
        return bTime.localeCompare(aTime)
      })
      break
    case 'count':
      sorted.sort((a, b) => (countByRepo[b.id] ?? 0) - (countByRepo[a.id] ?? 0))
      break
    case 'name':
      sorted.sort((a, b) => a.name.localeCompare(b.name))
      break
    case 'created':
      sorted.sort((a, b) => b.created_at.localeCompare(a.created_at))
      break
  }
  return sorted
}

export function latestReleasesByRepository(releases: Release[]): Record<number, Release> {
  const latest: Record<number, Release> = {}
  for (const release of releases) {
    const current = latest[release.repository_id]
    if (!current || release.published_at > current.published_at) {
      latest[release.repository_id] = release
    }
  }
  return latest
}

export function releaseCountByRepository(releases: Release[]): Record<number, number> {
  const counts: Record<number, number> = {}
  for (const release of releases) {
    counts[release.repository_id] = (counts[release.repository_id] ?? 0) + 1
  }
  return counts
}

export function releasesOfRepository(releases: Release[], repositoryId: number): Release[] {
  return sortReleasesNewestFirst(releases.filter((release) => release.repository_id === repositoryId))
}

export function paginate<T>(items: T[], page: number, pageSize: number): T[] {
  const start = (page - 1) * pageSize
  return items.slice(start, start + pageSize)
}
