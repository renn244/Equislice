import axios, { AxiosError, type AxiosAdapter } from 'axios'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'
import { apiClient } from '../../../lib/axios'
import { createPanoramaSlice, getPanoramaUploadUrl, getPanoramaDownloads, getPanoramaStatus, getPanoramaArchiveUrl, uploadToBlob, downloadPanoramaTile } from './panorama.api'

const originalApiAdapter = apiClient.defaults.adapter
const originalDefaultAdapter = axios.defaults.adapter
const adapter = vi.fn<AxiosAdapter>()

beforeEach(() => {
  adapter.mockReset()
  apiClient.defaults.adapter = adapter
  axios.defaults.adapter = adapter
})

afterEach(() => {
  apiClient.defaults.adapter = originalApiAdapter
  axios.defaults.adapter = originalDefaultAdapter
})

function respondWith(data: unknown, status = 200) {
  adapter.mockImplementationOnce(async config => ({ data, status, statusText: 'OK', headers: {}, config }))
}

test('requests a presigned upload URL and submits JSON using the configured API URL', async () => {
  respondWith({ urlSAS: 'https://blob.example/upload', blobName: 'blob-1-room.png' })
  respondWith({ job_id: 'job-1' }, 201)
  await expect(getPanoramaUploadUrl('room.png', 'image/png')).resolves.toEqual({
    urlSAS: 'https://blob.example/upload', blobName: 'blob-1-room.png',
  })
  const input = {
    file: 'blob-1-room.png', rows: 3, columns: 6,
    file_formats: 'jpg', file_name_format: 'tile_r{row}_c{col}',
  }
  await expect(createPanoramaSlice(input)).resolves.toEqual({ jobId: 'job-1' })
  const config = adapter.mock.calls[1][0]
  expect(apiClient.getUri(config)).toBe(apiClient.defaults.baseURL + '/panorama/slice')
  expect(config.method).toBe('post')
  expect(config.headers.get('Content-Type')).toBe('application/json')
  expect(JSON.parse(config.data)).toEqual(input)
})

test('uploads the raw image to Azure with its required headers and no API base URL', async () => {
  respondWith(null, 201)
  const file = new File(['image'], 'room.png', { type: 'image/png' })
  const url = 'https://blob.example/upload?sig=signed%2Bvalue'
  await expect(uploadToBlob(url, file)).resolves.toBeUndefined()
  const config = adapter.mock.calls[0][0]
  expect(axios.getUri(config)).toBe(url)
  expect(config.method).toBe('put')
  expect(config.headers.get('x-ms-blob-type')).toBe('BlockBlob')
  expect(config.headers.get('Content-Type')).toBe('image/png')
  expect(config.data).toBe(file)
  expect(config.baseURL).toBeUndefined()
})

test('reads job status and downloads, encoding the job ID as a query parameter', async () => {
  respondWith({ status: 'Completed', columns: 4, rows: 3 })
  const tiles = [{ urlSAS: 'https://tiles.example/0.jpg', fileName: 'tile_r1_c1.jpg', row: 1, column: 1 }]
  respondWith({ tiles })
  await expect(getPanoramaStatus('job&other=1')).resolves.toEqual({ status: 'Completed', columns: 4, rows: 3 })
  await expect(getPanoramaDownloads('job&other=1')).resolves.toEqual({ tiles })
  expect(apiClient.getUri(adapter.mock.calls[0][0])).toBe(apiClient.defaults.baseURL + '/panorama/status?job-id=job%26other%3D1')
  expect(getPanoramaArchiveUrl('job&other=1')).toBe(apiClient.defaults.baseURL + '/panorama/download-all?job-id=job%26other%3D1')
})

test('downloads a tile as a Blob while keeping its Azure SAS URL unchanged', async () => {
  const blob = new Blob(['tile'], { type: 'image/jpeg' })
  respondWith(blob)
  const url = 'https://tiles.example/tile.jpg?sig=signed%2Bvalue'
  await expect(downloadPanoramaTile(url)).resolves.toBe(blob)
  const config = adapter.mock.calls[0][0]
  expect(axios.getUri(config)).toBe(url)
  expect(config.responseType).toBe('blob')
  expect(config.baseURL).toBeUndefined()
})

test('propagates Axios failures for the job hook to handle', async () => {
  const error = new AxiosError('Request failed', 'ERR_NETWORK')
  adapter.mockRejectedValueOnce(error)
  await expect(getPanoramaStatus('job-1')).rejects.toBe(error)
})
