import { readFileSync } from 'node:fs'
import path from 'node:path'

import type { Release, Repository } from '@/types'

function readJSON<T>(name: string): T {
  const filePath = path.join(process.cwd(), 'public/data', name)
  return JSON.parse(readFileSync(filePath, 'utf-8')) as T
}

export function loadRepositories(): Repository[] {
  return readJSON<Repository[]>('repositories.json')
}

export function loadReleases(): Release[] {
  return readJSON<Release[]>('releases.json')
}

export function loadReleasesByRepository(repositoryId: number): Release[] {
  return loadReleases().filter((release) => release.repository_id === repositoryId)
}
