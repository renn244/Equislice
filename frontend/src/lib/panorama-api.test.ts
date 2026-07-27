import { afterEach, expect, test, vi } from 'vitest'
import {
  createPanoramaSlice,
  getPanoaramaUploadUrl,
  getPanoramaDownloads,
  getPanoramaStatus,
  uploadToBlob,
} from './panorama-api'

const apiBaseUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:3000/api'

afterEach(() => {
  vi.unstubAllGlobals()
})

test('requests a presigned upload url and submits the slice as JSON', async () => {
  const fetchMock = vi.fn()
  vi.stubGlobal('fetch', fetchMock)

  fetchMock.mockResolvedValueOnce(
    new Response(JSON.stringify({ urlSAS: 'https://blob.example/upload', blobName: 'blob-1-room.png' }), {
      status: 200,
    }),
  )
  fetchMock.mockResolvedValueOnce(
    new Response(JSON.stringify({ job_id: 'job-1' }), { status: 201 }),
  )

  await expect(getPanoaramaUploadUrl('room.png', 'image/png')).resolves.toEqual({
    urlSAS: 'https://blob.example/upload',
    blobName: 'blob-1-room.png',
  })

  await expect(
    createPanoramaSlice({
      file: 'blob-1-room.png',
      rows: 3,
      columns: 6,
      file_formats: 'jpg',
    }),
  ).resolves.toEqual({ jobId: 'job-1' })

  const [url, init] = fetchMock.mock.calls[1] as [string, RequestInit]
  expect(url).toBe(`${apiBaseUrl}/panorama/slice`)
  expect(init.method).toBe('POST')
  expect(init.headers).toEqual({ 'Content-Type': 'application/json' })
  expect(init.body).toBe(JSON.stringify({
    file: 'blob-1-room.png',
    rows: 3,
    columns: 6,
    file_formats: 'jpg',
  }))
})

test('uploads the panorama directly to the presigned blob url', async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 201 }))
  vi.stubGlobal('fetch', fetchMock)

  const file = new File(['image'], 'room.png', { type: 'image/png' })

  await expect(uploadToBlob('https://blob.example/upload', file)).resolves.toBeUndefined()

  const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
  expect(url).toBe('https://blob.example/upload')
  expect(init.method).toBe('PUT')
  expect(init.headers).toEqual({
    'x-ms-blob-type': 'BlockBlob',
    'Content-Type': 'image/png',
  })
  expect(init.body).toBe(file)
})

test('reads status and completed download URLs for a job', async () => {
  vi.stubGlobal(
    'fetch',
    vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ status: 'Completed' }), { status: 200 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ urlsSAS: ['https://tiles.example/0.jpg'] }), { status: 200 }),
      ),
  )

  await expect(getPanoramaStatus('job-1')).resolves.toEqual({ status: 'Completed' })
  await expect(getPanoramaDownloads('job-1')).resolves.toEqual({
    urlsSAS: ['https://tiles.example/0.jpg'],
  })
})
