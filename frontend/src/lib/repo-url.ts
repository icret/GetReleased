export interface ParsedRepo {
  owner: string
  name: string
}

export function parseRepoInput(input: string): ParsedRepo | null {
  const trimmed = input.trim()
  if (!trimmed) return null

  let path: string
  try {
    path = new URL(trimmed).pathname
  } catch {
    path = trimmed
  }

  const parts = path.split('/').filter(Boolean)
  if (parts.length < 2) return null

  const owner = parts[0]
  let name = parts[1]
  if (name.endsWith('.git')) {
    name = name.slice(0, -4)
  }
  if (!owner || !name) return null
  return { owner, name }
}
