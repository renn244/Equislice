import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, expect, test, vi } from 'vitest'
import { PanoramaPreview } from './panorama-preview'

afterEach(cleanup)

test('renders a flat panorama with a CSS-variable grid overlay', () => {
  const { rerender } = render(
    <PanoramaPreview
      columns={6}
      fileName="living-room.jpg"
      onFileChange={vi.fn()}
      previewUrl="blob:living-room"
      rows={3}
    />,
  )

  expect(screen.getByRole('img', { name: 'Preview of living-room.jpg' }))
    .toHaveAttribute('src', 'blob:living-room')
  expect(screen.getByRole('img', { name: 'Preview of living-room.jpg' }))
    .toHaveClass('panorama-preview-image')
  expect(screen.getByTestId('panorama-grid'))
    .toHaveStyle({ '--preview-columns': '6', '--preview-rows': '3' })

  rerender(
    <PanoramaPreview
      columns={4}
      fileName="living-room.jpg"
      onFileChange={vi.fn()}
      previewUrl="blob:living-room"
      rows={2}
    />,
  )

  expect(screen.getByTestId('panorama-grid'))
    .toHaveStyle({ '--preview-columns': '4', '--preview-rows': '2' })
})

test('renders an accessible upload state without a selected image', () => {
  render(
    <PanoramaPreview
      columns={6}
      fileName={null}
      onFileChange={vi.fn()}
      previewUrl={null}
      rows={3}
    />,
  )

  expect(screen.getByLabelText('Panorama image')).toHaveAttribute('type', 'file')
  expect(screen.getByText('Upload a panorama')).toBeInTheDocument()
  expect(screen.queryByRole('img')).not.toBeInTheDocument()
})
