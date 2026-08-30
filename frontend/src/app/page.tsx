import { loadRecentReleases, loadRepositories } from '@/data/load'
import Home from '@/features/home/Home'

export default function HomePage() {
  const repositories = loadRepositories()
  const recentReleases = loadRecentReleases()
  return <Home repositories={repositories} recentReleases={recentReleases} />
}
