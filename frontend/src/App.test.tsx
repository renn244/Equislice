import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, expect, test, vi } from 'vitest'
import App from './App'
import { createPanoramaSlice } from './lib/panorama-api'
import { router } from './router'

vi.mock('./lib/panorama-api', () => ({
  createPanoramaSlice: vi.fn().mockResolvedValue({ jobId: 'job-1' }),
  getPanoramaStatus: vi.fn().mockResolvedValue({ status: 'Completed' }),
  getPanoramaDownloads: vi.fn().mockResolvedValue({ urlsSAS: ['https://tiles.example/0.jpg'] }),
}))

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

test('renders the Open Canvas landing page and links to the cutter route', async () => {
  await router.navigate({ to: '/' })
  render(<App />)

  expect(
    await screen.findByRole('heading', {
      level: 1,
      name: 'Turn one panorama into a clean tile set.',
    }),
  ).toBeInTheDocument()
  expect(
    screen.getByText(
      'Upload an equirectangular image, choose your grid, and download evenly organized tiles for your virtual-tour workflow.',
    ),
  ).toBeInTheDocument()
  expect(screen.getByRole('img', { name: 'A panorama arranged as an evenly sliced tile set' }))
    .toHaveAttribute('src', '/panorama-tiles-transparent.png')
  expect(screen.getAllByText(/automatically deleted after 24 hours/i)).toHaveLength(2)
  expect(screen.getByRole('heading', { name: 'From panorama to tiles in three steps' }))
    .toBeInTheDocument()
  expect(screen.getByRole('heading', { name: 'Upload your panorama' })).toBeInTheDocument()
  expect(screen.getByRole('heading', { name: 'Choose your grid' })).toBeInTheDocument()
  expect(screen.getByRole('heading', { name: 'Export your tiles' })).toBeInTheDocument()
  expect(screen.getByRole('heading', { name: 'Choose the exact grid' })).toBeInTheDocument()
  expect(screen.getByRole('heading', { name: 'See the tile count' })).toBeInTheDocument()
  expect(screen.getByRole('heading', { name: 'Keep names predictable' })).toBeInTheDocument()
  expect(screen.getByText('tile_r01_c01.jpg')).toBeInTheDocument()
  expect(screen.queryByText('Follow every job')).not.toBeInTheDocument()
  expect(screen.getByRole('link', { name: 'How it works' })).toHaveAttribute('href', '#how-it-works')
  expect(screen.getAllByRole('link', { name: 'Open cutter' })).toEqual(
    expect.arrayContaining([
      expect.objectContaining({ href: expect.stringMatching(/\/cutter$/) }),
    ]),
  )
})

test('keeps the closing call to action and footer free of section separators', async () => {
  render(<App />)

  const closingHeading = await screen.findByRole('heading', {
    name: 'One panorama in. A clean tile set out.',
  })
  const closingSection = closingHeading.closest('section')
  const footer = screen.getByRole('contentinfo')

  expect(closingSection).not.toBeNull()
  expect(getComputedStyle(closingSection!).borderTopStyle).toBe('none')
  expect(getComputedStyle(footer).borderTopStyle).toBe('none')
})

test('renders the cutter with disabled submission before an image is selected', async () => {
  await router.navigate({ to: '/cutter' })
  render(<App />)

  expect(await screen.findByRole('heading', { name: 'Create your tile set' })).toBeInTheDocument()
  expect(screen.getByLabelText('Panorama image')).toHaveAttribute('type', 'file')
  expect(screen.getByLabelText('Columns')).toHaveValue(6)
  expect(screen.getByLabelText('Rows')).toHaveValue(3)
  expect(screen.getByRole('button', { name: 'Create tiles' })).toBeDisabled()
  expect(screen.getByText(/automatically deleted after 24 hours/i)).toBeInTheDocument()
})

test('shows the image summary and enables a valid tile setup', async () => {
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
  await router.navigate({ to: '/cutter' })
  render(<App />)

  fireEvent.change(screen.getByLabelText('Panorama image'), {
    target: { files: [new File(['image'], 'living-room.jpg', { type: 'image/jpeg' })] },
  })

  expect(await screen.findByText('living-room.jpg')).toBeInTheDocument()
  expect(screen.getByText(/4096 × 2048 px/i)).toBeInTheDocument()
  expect(screen.getByText('18 tiles')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Create tiles' })).toBeEnabled()
})

test('keeps submission disabled for unsupported files and invalid grids', async () => {
  await router.navigate({ to: '/cutter' })
  render(<App />)

  fireEvent.change(screen.getByLabelText('Panorama image'), {
    target: { files: [new File(['text'], 'notes.txt', { type: 'text/plain' })] },
  })
  fireEvent.change(screen.getByLabelText('Columns'), { target: { value: '0' } })

  expect(await screen.findByText('Choose a supported image file.')).toBeInTheDocument()
  expect(screen.getByText('Rows and columns must be whole numbers from 1 to 100.')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Create tiles' })).toBeDisabled()
})

test('submits a valid cutter job and exposes completed tile downloads', async () => {
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
  await router.navigate({ to: '/cutter' })
  render(<App />)

  fireEvent.change(screen.getByLabelText('Panorama image'), {
    target: { files: [new File(['image'], 'living-room.jpg', { type: 'image/jpeg' })] },
  })
  fireEvent.click(screen.getByRole('button', { name: 'Create tiles' }))

  expect(await screen.findByText('Job completed')).toBeInTheDocument()
  expect(screen.getByRole('link', { name: 'Download tile 1' })).toHaveAttribute(
    'href',
    'https://tiles.example/0.jpg',
  )
})

test('shows an error when a valid cutter job cannot be created', async () => {
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
  vi.mocked(createPanoramaSlice).mockRejectedValueOnce(new Error('network down'))
  await router.navigate({ to: '/cutter' })
  render(<App />)

  fireEvent.change(screen.getByLabelText('Panorama image'), {
    target: { files: [new File(['image'], 'living-room.jpg', { type: 'image/jpeg' })] },
  })
  fireEvent.click(screen.getByRole('button', { name: 'Create tiles' }))

  expect(await screen.findByRole('alert')).toHaveTextContent('We could not create this tile set. Please try again.')
})
