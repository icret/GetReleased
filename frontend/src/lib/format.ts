const MINUTE = 60 * 1000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

const dateFormatter = new Intl.DateTimeFormat('zh-CN', {
  timeZone: 'UTC',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
})

export function formatDate(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) {
    return 'Invalid Date'
  }
  return dateFormatter.format(date)
}

export function formatRelativeTime(iso: string, now: Date = new Date()): string {
  const then = new Date(iso)
  const diff = now.getTime() - then.getTime()

  if (Number.isNaN(diff)) {
    return formatDate(iso)
  }
  const absolute = Math.abs(diff)
  if (absolute < MINUTE) {
    return '刚刚'
  }
  if (absolute < HOUR) {
    return `${Math.floor(absolute / MINUTE)} 分钟前`
  }
  if (absolute < DAY) {
    return `${Math.floor(absolute / HOUR)} 小时前`
  }
  if (absolute < 30 * DAY) {
    return `${Math.floor(absolute / DAY)} 天前`
  }
  return formatDate(iso)
}

const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB']

export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) {
    return '0 B'
  }
  if (bytes < 1024) {
    return `${bytes} B`
  }
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < BYTE_UNITS.length - 1) {
    value /= 1024
    unit++
  }
  const display = value >= 10 ? Math.round(value).toString() : value.toFixed(1).replace(/\.0$/, '')
  return `${display} ${BYTE_UNITS[unit]}`
}
