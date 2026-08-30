const gradients = ['from-blue-500 to-indigo-600', 'from-indigo-500 to-violet-600', 'from-sky-500 to-blue-600', 'from-violet-500 to-purple-600', 'from-blue-400 to-cyan-600', 'from-cyan-500 to-blue-600']

export function gradientFor(name: string): string {
  let hash = 0
  for (const ch of name) {
    hash = (hash * 31 + ch.charCodeAt(0)) >>> 0
  }
  return gradients[hash % gradients.length]
}
