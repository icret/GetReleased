'use client'

import { clearToken, getToken } from './auth'

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || ''

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function rawFetch(path: string, options: RequestInit = {}): Promise<unknown> {
  const headers = new Headers(options.headers)
  headers.set('Content-Type', 'application/json')
  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  const res = await fetch(`${BASE_URL}${path}`, { ...options, headers })
  if (res.status === 401) {
    clearToken()
    if (typeof window !== 'undefined') {
      window.location.href = '/login'
    }
    throw new ApiError(401, 'unauthorized')
  }
  const json = await res.json()
  if (!res.ok) {
    throw new ApiError(res.status, json.error || 'request failed')
  }
  return json.data
}

export async function fetchAPI<T>(path: string, options: RequestInit = {}): Promise<T> {
  return (await rawFetch(path, options)) as T
}

export async function fetchAPIList<T>(path: string, options: RequestInit = {}): Promise<T[]> {
  const data = await rawFetch(path, options)
  return Array.isArray(data) ? (data as T[]) : []
}
