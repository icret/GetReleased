'use client'

import type { LucideIcon } from 'lucide-react'
import { Archive, Flame, GitFork, Lock, Rocket, Tag, Unlink, Users } from 'lucide-react'
import { TAG_TYPE_LABELS, type DashboardOverview, type TagTypeCount } from '@/types'

interface StatsCardsProps {
  overview: DashboardOverview
  tagTypes: TagTypeCount[]
}

const PRIMARY_CARDS: { key: keyof DashboardOverview; label: string; icon: LucideIcon }[] = [
  { key: 'repository_count', label: '追踪仓库', icon: GitFork },
  { key: 'release_count', label: 'Release 总数', icon: Rocket },
  { key: 'user_count', label: '管理员', icon: Users },
]

const SECONDARY_CARDS: { key: keyof DashboardOverview; label: string; icon: LucideIcon }[] = [
  { key: 'prerelease_count', label: '预发布', icon: Flame },
  { key: 'archived_count', label: '已归档', icon: Archive },
  { key: 'private_count', label: '私有仓库', icon: Lock },
  { key: 'untagged_count', label: '无标签', icon: Unlink },
]

export function StatsCards({ overview, tagTypes }: StatsCardsProps) {
  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
      {PRIMARY_CARDS.map((card) => (
        <StatCard key={card.key} label={card.label} value={overview[card.key]} icon={card.icon} primary />
      ))}
      {SECONDARY_CARDS.map((card) => (
        <StatCard key={card.key} label={card.label} value={overview[card.key]} icon={card.icon} />
      ))}
      {tagTypes.map((t) => (
        <StatCard key={`tag-type-${t.type}`} label={TAG_TYPE_LABELS[t.type as keyof typeof TAG_TYPE_LABELS] ?? t.type} value={t.count} icon={Tag} />
      ))}
    </div>
  )
}

function StatCard({ label, value, icon: Icon, primary }: { label: string; value: number; icon: LucideIcon; primary?: boolean }) {
  return (
    <div className="glass rounded-xl p-4">
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">{label}</span>
        <Icon className={`h-4 w-4 ${primary ? 'text-primary' : 'text-muted-foreground'}`} />
      </div>
      <p className="mt-2 font-display text-2xl font-bold tabular-nums">{value}</p>
    </div>
  )
}
