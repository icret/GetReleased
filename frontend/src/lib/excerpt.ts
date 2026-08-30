export function extractExcerpt(body: string): string {
  const line = body
    .split('\n')
    .map((item) => item.trim())
    .find((item) => item && !item.startsWith('#') && !item.startsWith('>') && !item.startsWith('<') && !item.startsWith('!['))
  if (line) {
    return line
  }
  const text = body.trim()
  return text ? text.slice(0, 120) : '暂无 Release Notes'
}
