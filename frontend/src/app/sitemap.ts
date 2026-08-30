import type { MetadataRoute } from 'next'

import { loadRepositories } from '@/data/load'

export const dynamic = 'force-static'

const SITE_URL = process.env.SITE_URL || 'https://getreleased.example.com'

export default function sitemap(): MetadataRoute.Sitemap {
  const repositories = loadRepositories()
  const entries: MetadataRoute.Sitemap = [{ url: SITE_URL, lastModified: new Date(), changeFrequency: 'daily', priority: 1 }]
  for (const repo of repositories) {
    entries.push({
      url: `${SITE_URL}/repository/${repo.owner}/${repo.name}`,
      lastModified: new Date(repo.created_at),
      changeFrequency: 'daily',
      priority: 0.8,
    })
  }
  return entries
}
