import { act, cleanup, renderHook } from '@testing-library/react'
import type { ChangeEvent } from 'react'
import { afterEach, expect, test, vi } from 'vitest'
import { usePanoramaImage } from './usePanoramaImage'

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

function fileEvent(file: File) {
  return { currentTarget: { files: [file], value: file.name } } as unknown as ChangeEvent<HTMLInputElement>
}

test('ignores a previous image finishing after a replacement and releases both preview URLs', () => {
  const images: DeferredImage[] = []
  class DeferredImage {
    naturalHeight = 2048
    naturalWidth = 4096
    onload: (() => void) | null = null
    onerror: (() => void) | null = null
    src = ''
    constructor() { images.push(this) }
  }
  const revoke = vi.fn()
  vi.stubGlobal('Image', DeferredImage)
  vi.stubGlobal('URL', {
    createObjectURL: vi.fn().mockReturnValueOnce('blob:first').mockReturnValueOnce('blob:second'),
    revokeObjectURL: revoke,
  })
  const { result, unmount } = renderHook(usePanoramaImage)
  act(() => result.current.onFileChange(fileEvent(new File(['one'], 'first.jpg', { type: 'image/jpeg' }))))
  act(() => result.current.onFileChange(fileEvent(new File(['two'], 'second.jpg', { type: 'image/jpeg' }))))
  act(() => images[0].onload?.())
  expect(result.current.selectedFile).toBeNull()
  expect(revoke).toHaveBeenCalledWith('blob:first')

  act(() => images[1].onload?.())
  expect(result.current.selectedFile?.name).toBe('second.jpg')
  expect(result.current.previewUrl).toBe('blob:second')
  unmount()
  expect(revoke).toHaveBeenCalledWith('blob:second')
})

test('rejects an image larger than the advertised 60 MB limit before reading it', () => {
  const createObjectURL = vi.fn()
  vi.stubGlobal('URL', { createObjectURL, revokeObjectURL: vi.fn() })
  const file = new File(['image'], 'large.jpg', { type: 'image/jpeg' })
  Object.defineProperty(file, 'size', { value: 60 * 1024 * 1024 + 1 })
  const { result } = renderHook(usePanoramaImage)
  act(() => result.current.onFileChange(fileEvent(file)))
  expect(result.current.fileError).toBe('Choose an image that is 60 MB or smaller.')
  expect(result.current.selectedFile).toBeNull()
  expect(createObjectURL).not.toHaveBeenCalled()
})
