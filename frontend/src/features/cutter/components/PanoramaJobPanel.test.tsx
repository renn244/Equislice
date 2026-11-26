import { act, cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'
import { PanoramaJobPanel } from './PanoramaJobPanel'
import type { JobState } from '../types/panorama.type'

vi.mock('@sentry/react', () => ({ captureException: vi.fn() }))

const startupNote = /Starting up the slicer/

function renderPanel(jobState: JobState, jobId: string | null = 'job-1') {
  const props = { jobId, jobState, jobError: null, downloads: [], downloadColumns: null }
  const view = render(<PanoramaJobPanel {...props} />)
  return {
    rerender: (nextState: JobState, nextJobId: string | null = jobId) =>
      view.rerender(<PanoramaJobPanel {...props} jobState={nextState} jobId={nextJobId} />),
  }
}

beforeEach(() => vi.useFakeTimers())

afterEach(() => {
  cleanup()
  vi.useRealTimers()
})

test('explains the startup wait once a job has been queued for 3 seconds', () => {
  renderPanel('queued')
  act(() => vi.advanceTimersByTime(2999))
  expect(screen.queryByText(startupNote)).not.toBeInTheDocument()

  act(() => vi.advanceTimersByTime(1))
  expect(screen.getByText(startupNote)).toBeInTheDocument()
})

test('hides the startup note once slicing begins', () => {
  const { rerender } = renderPanel('queued')
  act(() => vi.advanceTimersByTime(3000))
  expect(screen.getByText(startupNote)).toBeInTheDocument()

  rerender('processing')
  expect(screen.queryByText(startupNote)).not.toBeInTheDocument()
})

test('restarts the wait for a new job instead of carrying over the note', () => {
  const { rerender } = renderPanel('queued', 'job-1')
  act(() => vi.advanceTimersByTime(3000))
  expect(screen.getByText(startupNote)).toBeInTheDocument()

  rerender('queued', 'job-2')
  expect(screen.queryByText(startupNote)).not.toBeInTheDocument()
  act(() => vi.advanceTimersByTime(3000))
  expect(screen.getByText(startupNote)).toBeInTheDocument()
})

test('does not show the note for a job that moves on quickly', () => {
  const { rerender } = renderPanel('queued')
  act(() => vi.advanceTimersByTime(1500))
  rerender('processing')
  act(() => vi.advanceTimersByTime(5000))
  expect(screen.queryByText(startupNote)).not.toBeInTheDocument()
})
