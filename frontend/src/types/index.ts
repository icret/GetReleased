export type TagType = 'category' | 'platform'

export interface Tag {
  id: number
  name: string
  type: TagType
}

export const TAG_TYPE_LABELS: Record<TagType, string> = {
  category: '标签',
  platform: '分类',
}

export const TAG_TYPE_OPTIONS: { value: TagType; label: string }[] = [
  { value: 'category', label: '标签' },
  { value: 'platform', label: '分类' },
]

export interface Repository {
  id: number
  owner: string
  name: string
  full_name: string
  description?: string
  logo_path?: string
  stars: number
  language?: string
  is_archived: boolean
  is_private: boolean
  latest_version?: string
  latest_release_url?: string
  latest_release_date?: string
  latest_is_prerelease?: boolean
  release_count?: number
  last_checked_at?: string
  remark?: string
  created_at: string
  updated_at: string
  tags?: Tag[]
}

export interface ReleaseAsset {
  id: number
  release_id: number
  name: string
  size: number
  download_url: string
  content_type?: string
  created_at: string
}

export interface Release {
  id: number
  repository_id: number
  tag_name: string
  name: string
  body: string
  html_url: string
  tarball_url?: string
  zipball_url?: string
  published_at: string
  is_prerelease: boolean
  created_at: string
  assets?: ReleaseAsset[]
}

export interface User {
  id: number
  username: string
  role: string
  created_at: string
}

export interface DashboardOverview {
  repository_count: number
  release_count: number
  tag_count: number
  user_count: number
  prerelease_count: number
  archived_count: number
  private_count: number
  untagged_count: number
}

export interface LanguageCount {
  language: string
  count: number
}

export interface TopRepository {
  id: number
  full_name: string
  stars: number
  latest_version: string
  latest_release_url: string
  language: string
}

export interface RecentRelease {
  id: number
  repository_id: number
  full_name: string
  tag_name: string
  name: string
  html_url: string
  published_at: string
  is_prerelease: boolean
}

export interface ReleaseTrendPoint {
  month: string
  count: number
}

export interface TagTypeCount {
  type: string
  count: number
}

export interface DashboardStats {
  overview: DashboardOverview
  languages: LanguageCount[]
  top_repositories: TopRepository[]
  recent_releases: RecentRelease[]
  release_trend: ReleaseTrendPoint[]
  tag_types: TagTypeCount[]
}
