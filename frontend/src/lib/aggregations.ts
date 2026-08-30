import type { Release, Repository } from '@/types'

export type SortKey = 'latest' | 'count' | 'name' | 'created'

export function sortReleasesNewestFirst(releases: Release[]): Release[] {
  return releases.slice().sort((a, b) => b.published_at.localeCompare(a.published_at))
}

export function sortRepositories(repositories: Repository[], sortKey: SortKey): Repository[] {
  const sorted = repositories.slice()
  switch (sortKey) {
    case 'latest':
      sorted.sort((a, b) => (b.latest_release_date ?? '').localeCompare(a.latest_release_date ?? ''))
      break
    case 'count':
      sorted.sort((a, b) => (b.release_count ?? 0) - (a.release_count ?? 0))
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

export function paginate<T>(items: T[], page: number, pageSize: number): T[] {
  const start = (page - 1) * pageSize
  return items.slice(start, start + pageSize)
}
