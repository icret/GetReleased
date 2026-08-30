'use client'

const TOKEN_KEY = 'getreleased-admin-token'
const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || ''

export function getToken(): string | null {
  if (typeof window === 'undefined') return null
  return window.localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  if (typeof window === 'undefined') return
  window.localStorage.removeItem(TOKEN_KEY)
}

export function isLoggedIn(): boolean {
  return getToken() !== null
}

export function logout(): void {
  clearToken()
  if (typeof window !== 'undefined') {
    window.location.href = '/login'
  }
}

export async function login(username: string, password: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
  const json = await res.json()
  if (!res.ok) {
    throw new Error(json.error || 'login failed')
  }
  setToken(json.data.token)
}

export function getCurrentUsername(): string | null {
  const token = getToken()
  if (!token) return null
  try {
    const payload = token.split('.')[1]
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const padded = normalized.padEnd(normalized.length + ((4 - (normalized.length % 4)) % 4), '=')
    const decoded = JSON.parse(atob(padded))
    return typeof decoded.sub === 'string' ? decoded.sub : null
  } catch {
    return null
  }
}
