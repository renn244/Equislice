import axios from 'axios'
import { apiClient } from '../../../lib/axios'
import type { PanoramaSliceInput, TileDownload } from '../types/panorama.type'

export async function createPanoramaSlice(input: PanoramaSliceInput) {
  const { data } = await apiClient.post<{ job_id: string }>('/panorama/slice', input)
  return { jobId: data.job_id }
}

export async function getPanoramaStatus(jobId: string) {
  const { data } = await apiClient.get<{ status: string; columns: number; rows: number }>('/panorama/status', {
    params: { 'job-id': jobId },
  })
  return data
}

export async function getPanoramaUploadUrl(fileName: string, contentType: string) {
  const { data } = await apiClient.post<{ urlSAS: string; blobName: string }>('/panorama/upload', {
    file_name: fileName,
    content_type: contentType,
  })
  return data
}

export async function uploadToBlob(presignedUrl: string, file: File) {
  await axios.put(presignedUrl, file, {
    headers: {
      'x-ms-blob-type': 'BlockBlob',
      'Content-Type': file.type || 'application/octet-stream',
    },
  })
}

export async function downloadPanoramaTile(urlSAS: string) {
  const { data } = await axios.get<Blob>(urlSAS, { responseType: 'blob' })
  return data
}

export async function getPanoramaDownloads(jobId: string) {
  const { data } = await apiClient.get<{ tiles: TileDownload[] }>('/panorama/download', {
    params: { 'job-id': jobId },
  })
  return data
}

export function getPanoramaArchiveUrl(jobId: string) {
  return apiClient.getUri({ url: '/panorama/download-all', params: { 'job-id': jobId } })
}
