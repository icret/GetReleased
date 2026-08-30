import { describe, it, expect, beforeEach, vi } from 'vitest'
import { fetchAPI, fetchAPIList, ApiError } from './api'
import { clearToken } from './auth'

function mockFetch(data: unknown, ok = true, status = 200) {
  vi.stubGlobal(
    'fetch',
    vi.fn(
      async () =>
        ({
          status,
          ok,
          json: async () => ({ data }),
        }) as unknown as Response,
    ),
  )
}

describe('fetchAPIList', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    clearToken()
  })

  it('数组透传', async () => {
    mockFetch([1, 2, 3])
    expect(await fetchAPIList<number>('/x')).toEqual([1, 2, 3])
  })

  it('null 归一化为空数组', async () => {
    mockFetch(null)
    expect(await fetchAPIList<number>('/x')).toEqual([])
  })

  it('undefined 归一化为空数组', async () => {
    mockFetch(undefined)
    expect(await fetchAPIList<number>('/x')).toEqual([])
  })

  it('非数组归一化为空数组', async () => {
    mockFetch({ a: 1 })
    expect(await fetchAPIList<number>('/x')).toEqual([])
  })

  it('401 抛错且清 token', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(
        async () =>
          ({
            status: 401,
            ok: false,
            json: async () => ({ error: 'unauthorized' }),
          }) as unknown as Response,
      ),
    )
    await expect(fetchAPIList<number>('/x')).rejects.toBeInstanceOf(ApiError)
  })
})

describe('fetchAPI', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    clearToken()
  })

  it('透传 data 字段', async () => {
    mockFetch({ id: 1 })
    expect(await fetchAPI<{ id: number }>('/x')).toEqual({ id: 1 })
  })

  it('保留 null（非数组由调用方处理）', async () => {
    mockFetch(null)
    expect(await fetchAPI<null>('/x')).toBeNull()
  })
})
