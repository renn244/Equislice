import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, expect, test, vi } from 'vitest'
import App from './App'
import { createPanoramaSlice } from './lib/panorama-api'
import { router } from './router'

vi.mock('./lib/panorama-api', () => ({
  createPanoramaSlice: vi.fn().mockResolvedValue({ jobId: 'job-1' }),
  getPanoramaStatus: vi.fn().mockResolvedValue({ status: 'Completed' }),
  getPanoramaDownloads: vi.fn().mockResolvedValue({ urlsSAS: ['https://tiles.example/0.jpg'] }),
  getPanoramaArchiveUrl: vi.fn(
    (jobId: string) => `https://api.example/panorama/download-all?job-id=${jobId}`,
  ),
}))

function stubReadableImage() {
  vi.stubGlobal('URL', {
    createObjectURL: vi.fn(() => 'blob:living-room'),
    revokeObjectURL: vi.fn(),
  })
  vi.stubGlobal('Image', class {
    naturalHeight = 2048
    naturalWidth = 4096
    onerror: (() => void) | null = null
    onload: (() => void) | null = null

    set src(_value: string) {
      this.onload?.()
    }
  })
}

async function renderCutter() {
  await router.navigate({ to: '/cutter' })
  render(<App />)
  await screen.findByRole('heading', { name: 'Add a panorama to begin.' })
}

async function selectPanorama() {
  stubReadableImage()
  fireEvent.change(screen.getByLabelText('Panorama image'), {
    target: { files: [new File(['image'], 'living-room.jpg', { type: 'image/jpeg' })] },
  })
  await screen.findByText('living-room.jpg')
}

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
  vi.mocked(createPanoramaSlice).mockReset()
  vi.mocked(createPanoramaSlice).mockResolvedValue({ jobId: 'job-1' })
})

test('renders the empty cutter state before a panorama is selected', async () => {
  await renderCutter()

  expect(screen.getByLabelText('Panorama image')).toHaveAttribute('type', 'file')
  expect(screen.getByRole('button', { name: /ridgeline-valley.jpg/i })).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: /Begin slicing/ })).not.toBeInTheDocument()
})

test('shows the live grid configuration after a valid panorama is selected', async () => {
  await renderCutter()
  await selectPanorama()

  expect(screen.getByRole('img', { name: 'Preview of living-room.jpg' })).toHaveAttribute(
    'src',
    'blob:living-room',
  )
  expect(screen.getByRole('spinbutton', { name: 'Columns' })).toHaveValue(4)
  expect(screen.getByRole('spinbutton', { name: 'Rows' })).toHaveValue(3)
  expect(screen.getByText(/4096.*2048/)).toBeInTheDocument()
  expect(screen.getByText('12')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: /Begin slicing/ })).toBeEnabled()
})

test('rejects unsupported files and disables submission for an invalid grid', async () => {
  await renderCutter()
  fireEvent.change(screen.getByLabelText('Panorama image'), {
    target: { files: [new File(['text'], 'notes.txt', { type: 'text/plain' })] },
  })
  expect(await screen.findByText('Choose a JPG or PNG image.')).toBeInTheDocument()

  await selectPanorama()
  fireEvent.change(screen.getByRole('spinbutton', { name: 'Columns' }), { target: { value: '0' } })

  expect(screen.getByText('Rows and columns must be whole numbers from 1 to 100.')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: /Begin slicing/ })).toBeDisabled()
})

test('submits a valid cutter job and exposes Azure tile downloads', async () => {
  await renderCutter()
  await selectPanorama()
  fireEvent.click(screen.getByRole('button', { name: /Begin slicing/ }))

  expect(await screen.findByRole('heading', { name: 'The grid is ready.' })).toBeInTheDocument()
  expect(screen.getByRole('img', { name: 'Sliced tile 1' })).toHaveAttribute(
    'src',
    'https://tiles.example/0.jpg',
  )
  expect(screen.getByRole('link', { name: 'Download all tiles (.zip)' })).toHaveAttribute(
    'href',
    'https://api.example/panorama/download-all?job-id=job-1',
  )
})

test('shows a submit error when a valid cutter job cannot be created', async () => {
  vi.mocked(createPanoramaSlice).mockRejectedValueOnce(new Error('network down'))
  await renderCutter()
  await selectPanorama()
  fireEvent.click(screen.getByRole('button', { name: /Begin slicing/ }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'We could not create this tile set. Please try again.',
  )
})
