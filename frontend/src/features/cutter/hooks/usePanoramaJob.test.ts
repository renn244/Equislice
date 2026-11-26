import { act, cleanup, renderHook } from '@testing-library/react'
import { afterEach, expect, test, vi } from 'vitest'
import { getPanoramaDownloads, getPanoramaStatus } from '../api/panorama.api'
import { usePanoramaJob } from './usePanoramaJob'

vi.mock('../api/panorama.api', () => ({
  getPanoramaStatus: vi.fn(),
  getPanoramaDownloads: vi.fn().mockResolvedValue({ tiles: [] }),
}))

vi.mock('@sentry/react', () => ({ metrics: { count: vi.fn() } }))

afterEach(() => {
  cleanup()
  window.localStorage.clear()
  vi.useRealTimers()
  vi.clearAllMocks()
})

test('waits for a slow status request before scheduling the next poll and stops on completion', async () => {
  vi.useFakeTimers()
  window.localStorage.setItem('equislice.cutter.job-id', 'saved-job')
  let finishStatus!: (status: { status: string; columns: number; rows: number }) => void
  vi.mocked(getPanoramaStatus)
    .mockImplementationOnce(() => new Promise(resolve => { finishStatus = resolve }))
    .mockResolvedValue({ status: 'Completed', columns: 4, rows: 3 })
  const { result } = renderHook(usePanoramaJob)

  await act(() => vi.advanceTimersByTimeAsync(3000))
  expect(getPanoramaStatus).toHaveBeenCalledTimes(1)

  await act(async () => { finishStatus({ status: 'Processing', columns: 4, rows: 3 }) })
  expect(result.current.jobState).toBe('processing')
  await act(() => vi.advanceTimersByTimeAsync(1000))
  expect(result.current.jobState).toBe('completed')
  expect(getPanoramaDownloads).toHaveBeenCalledWith('saved-job')
  await act(() => vi.advanceTimersByTimeAsync(5000))
  expect(getPanoramaStatus).toHaveBeenCalledTimes(2)
})

test('discards the result of a status request after the hook unmounts', async () => {
  window.localStorage.setItem('equislice.cutter.job-id', 'saved-job')
  let finishStatus!: (status: { status: string; columns: number; rows: number }) => void
  vi.mocked(getPanoramaStatus).mockImplementationOnce(() => new Promise(resolve => { finishStatus = resolve }))
  const { unmount } = renderHook(usePanoramaJob)
  unmount()
  await act(async () => { finishStatus({ status: 'Completed', columns: 4, rows: 3 }) })
  expect(getPanoramaDownloads).not.toHaveBeenCalled()
})
