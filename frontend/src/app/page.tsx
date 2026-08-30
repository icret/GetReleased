import { loadReleases, loadRepositories } from '@/data/load'
import Home from '@/features/home/Home'

export default function HomePage() {
  const repositories = loadRepositories()
  const releases = loadReleases()
  return <Home repositories={repositories} releases={releases} />
}
