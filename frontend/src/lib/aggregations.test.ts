import { describe, it, expect } from 'vitest'
import type { Release, Repository } from '@/types'
import { latestReleasesByRepository, releaseCountByRepository, releasesOfRepository, sortReleasesNewestFirst, sortRepositories, paginate } from './aggregations'

function release(overrides: Partial<Release>): Release {
  return {
    id: 1,
    repository_id: 1,
    tag_name: 'v1.0.0',
    name: 'v1.0.0',
    body: '',
    html_url: 'https://github.com/owner/repo/releases/tag/v1.0.0',
    published_at: '2026-01-01T00:00:00Z',
    is_prerelease: false,
    created_at: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

function repository(overrides: Partial<Repository>): Repository {
  return {
    id: 1,
    owner: 'owner',
    name: 'repo',
    full_name: 'owner/repo',
    stars: 0,
    is_archived: false,
    is_private: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

const releases: Release[] = [
  release({ id: 1, repository_id: 1, tag_name: 'v1.0.0', published_at: '2026-01-01T00:00:00Z' }),
  release({ id: 2, repository_id: 1, tag_name: 'v1.1.0', published_at: '2026-03-01T00:00:00Z' }),
  release({ id: 3, repository_id: 2, tag_name: 'v2.0.0', published_at: '2026-06-01T00:00:00Z' }),
]

describe('sortReleasesNewestFirst', () => {
  it('按发布时间倒序排列', () => {
    const sorted = sortReleasesNewestFirst(releases)
    expect(sorted.map((r) => r.id)).toEqual([3, 2, 1])
  })

  it('不修改原数组', () => {
    const before = releases.map((r) => r.id)
    sortReleasesNewestFirst(releases)
    expect(releases.map((r) => r.id)).toEqual(before)
  })

  it('空数组返回空数组', () => {
    expect(sortReleasesNewestFirst([])).toEqual([])
  })
})

describe('latestReleasesByRepository', () => {
  it('按仓库取发布时间最新的版本', () => {
    const latest = latestReleasesByRepository(releases)
    expect(latest[1]?.tag_name).toBe('v1.1.0')
    expect(latest[2]?.tag_name).toBe('v2.0.0')
  })

  it('空数组返回空对象', () => {
    expect(latestReleasesByRepository([])).toEqual({})
  })
})

describe('releaseCountByRepository', () => {
  it('统计每个仓库的版本数量', () => {
    expect(releaseCountByRepository(releases)).toEqual({ 1: 2, 2: 1 })
  })
})

describe('releasesOfRepository', () => {
  it('按仓库过滤并倒序排列', () => {
    const filtered = releasesOfRepository(releases, 1)
    expect(filtered.map((r) => r.tag_name)).toEqual(['v1.1.0', 'v1.0.0'])
  })

  it('无匹配版本返回空数组', () => {
    expect(releasesOfRepository(releases, 99)).toEqual([])
  })
})

const repos: Repository[] = [repository({ id: 1, name: 'beta', created_at: '2026-01-01T00:00:00Z' }), repository({ id: 2, name: 'alpha', created_at: '2026-06-01T00:00:00Z' }), repository({ id: 3, name: 'gamma', created_at: '2026-03-01T00:00:00Z' })]

describe('sortRepositories', () => {
  it('latest 按最新发布时间降序', () => {
    const latestByRepo = { 1: releases[1], 2: releases[2] }
    const countByRepo = { 1: 2, 2: 1 }
    const sorted = sortRepositories(repos, { latestByRepo, countByRepo, sortKey: 'latest' })
    expect(sorted.map((r) => r.id)).toEqual([2, 1, 3])
  })

  it('count 按 Release 数量降序', () => {
    const latestByRepo = { 1: releases[1], 2: releases[2] }
    const countByRepo = { 1: 2, 2: 1 }
    const sorted = sortRepositories(repos, { latestByRepo, countByRepo, sortKey: 'count' })
    expect(sorted.map((r) => r.id)).toEqual([1, 2, 3])
  })

  it('name 按名称升序', () => {
    const sorted = sortRepositories(repos, { latestByRepo: {}, countByRepo: {}, sortKey: 'name' })
    expect(sorted.map((r) => r.name)).toEqual(['alpha', 'beta', 'gamma'])
  })

  it('created 按追踪时间降序', () => {
    const sorted = sortRepositories(repos, { latestByRepo: {}, countByRepo: {}, sortKey: 'created' })
    expect(sorted.map((r) => r.id)).toEqual([2, 3, 1])
  })

  it('不修改原数组', () => {
    const before = repos.map((r) => r.id)
    sortRepositories(repos, { latestByRepo: {}, countByRepo: {}, sortKey: 'name' })
    expect(repos.map((r) => r.id)).toEqual(before)
  })
})

describe('paginate', () => {
  const items = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

  it('返回当前页的元素', () => {
    expect(paginate(items, 1, 3)).toEqual([1, 2, 3])
    expect(paginate(items, 2, 3)).toEqual([4, 5, 6])
    expect(paginate(items, 4, 3)).toEqual([10])
  })

  it('超出范围返回空数组', () => {
    expect(paginate(items, 5, 3)).toEqual([])
  })

  it('空数组返回空数组', () => {
    expect(paginate([], 1, 3)).toEqual([])
  })
})
