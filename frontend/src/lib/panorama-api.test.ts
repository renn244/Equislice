import { afterEach, expect, test, vi } from 'vitest'
import { createPanoramaSlice, getPanoramaDownloads, getPanoramaStatus } from './panorama-api'

const apiBaseUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:3000/api'

afterEach(() => {
  vi.unstubAllGlobals()
})

test('submits the selected panorama as the backend multipart contract', async () => {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(JSON.stringify({ job_id: 'job-1' }), { status: 201 }),
  )
  vi.stubGlobal('fetch', fetchMock)

  await expect(
    createPanoramaSlice({
      file: new File(['image'], 'room.png', { type: 'image/png' }),
      rows: 3,
      columns: 6,
      fileFormat: 'jpg',
    }),
  ).resolves.toEqual({ jobId: 'job-1' })

  const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
  expect(url).toBe(`${apiBaseUrl}/panorama/slice`)
  expect(init.method).toBe('POST')
  expect(init.body).toBeInstanceOf(FormData)
  expect((init.body as FormData).get('rows')).toBe('3')
  expect((init.body as FormData).get('columns')).toBe('6')
  expect((init.body as FormData).get('fileFormat')).toBe('jpg')
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
