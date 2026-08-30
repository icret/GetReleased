import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useTrackStatus } from './useTrackStatus'

vi.mock('@/lib/api', () => ({
  fetchAPI: vi.fn(),
}))

vi.mock('@/components/ui/toast', () => ({
  toast: { add: vi.fn() },
}))

const { fetchAPI } = await import('@/lib/api')
const { toast } = await import('@/components/ui/toast')

beforeEach(() => {
  vi.clearAllMocks()
  vi.mocked(fetchAPI).mockResolvedValue({ running: false })
})

describe('useTrackStatus', () => {
  it('初始化时拉取状态', async () => {
    const { result } = renderHook(() => useTrackStatus())
    await waitFor(() => expect(result.current.status.running).toBe(false))
    expect(fetchAPI).toHaveBeenCalledWith('/api/admin/track/status')
  })

  it('trigger 发起 track 请求并切换为运行中', async () => {
    const { result } = renderHook(() => useTrackStatus())
    await waitFor(() => expect(fetchAPI).toHaveBeenCalled())

    vi.mocked(fetchAPI).mockResolvedValueOnce({ running: true })
    await act(async () => {
      await result.current.trigger()
    })
    expect(fetchAPI).toHaveBeenCalledWith('/api/admin/track', { method: 'POST' })
    expect(result.current.status.running).toBe(true)
    expect(result.current.triggering).toBe(false)
  })

  it('trigger 失败时弹出错误提示且不进入运行态', async () => {
    const { result } = renderHook(() => useTrackStatus())
    await waitFor(() => expect(fetchAPI).toHaveBeenCalled())

    vi.mocked(fetchAPI).mockRejectedValueOnce(new Error('network'))
    await act(async () => {
      await result.current.trigger()
    })
    expect(toast.add).toHaveBeenCalledWith({ title: '触发失败', description: 'network' })
    expect(result.current.status.running).toBe(false)
    expect(result.current.triggering).toBe(false)
  })

  it('exportOnly 发起 export 请求并提示完成', async () => {
    const { result } = renderHook(() => useTrackStatus())
    await waitFor(() => expect(fetchAPI).toHaveBeenCalled())

    vi.mocked(fetchAPI).mockResolvedValueOnce({ exported: true })
    await act(async () => {
      await result.current.exportOnly()
    })
    expect(fetchAPI).toHaveBeenCalledWith('/api/admin/export', { method: 'POST' })
    expect(toast.add).toHaveBeenCalledWith({ title: '导出完成' })
    expect(result.current.exporting).toBe(false)
  })

  it('exportOnly 失败时提示导出失败', async () => {
    const { result } = renderHook(() => useTrackStatus())
    await waitFor(() => expect(fetchAPI).toHaveBeenCalled())

    vi.mocked(fetchAPI).mockRejectedValueOnce(new Error('disk full'))
    await act(async () => {
      await result.current.exportOnly()
    })
    expect(toast.add).toHaveBeenCalledWith({ title: '导出失败', description: 'disk full' })
    expect(result.current.exporting).toBe(false)
  })

  it('track 完成后自动导出', async () => {
    vi.useFakeTimers()
    const { result } = renderHook(() => useTrackStatus())
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })

    await act(async () => {
      await result.current.trigger()
    })

    vi.mocked(fetchAPI).mockResolvedValueOnce({ running: false })
    vi.mocked(fetchAPI).mockResolvedValueOnce({ exported: true })

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000)
    })
    expect(fetchAPI).toHaveBeenCalledWith('/api/admin/export', { method: 'POST' })
    expect(toast.add).toHaveBeenCalledWith({ title: '追踪并导出完成' })
    vi.useRealTimers()
  })

  it('track 失败时不导出并提示追踪失败', async () => {
    vi.useFakeTimers()
    const { result } = renderHook(() => useTrackStatus())
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })

    await act(async () => {
      await result.current.trigger()
    })

    vi.mocked(fetchAPI).mockResolvedValueOnce({ running: false, error: 'boom' })

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000)
    })
    expect(toast.add).toHaveBeenCalledWith({ title: '追踪失败', description: 'boom' })
    expect(fetchAPI).not.toHaveBeenCalledWith('/api/admin/export', { method: 'POST' })
    vi.useRealTimers()
  })
})
