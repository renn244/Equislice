const apiBaseUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:3000/api'

type PanoramaSliceInput = {
  file: string
  rows: number
  columns: number
  file_formats: string
}

async function requestJson<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, init)

  if (!response.ok) {
    throw new Error('The panorama request failed. Please try again.')
  }

  return response.json() as Promise<T>
}

async function requestBlobUpload(url: string, init?: RequestInit): Promise<void> {
  const response = await fetch(url, init)

  if (!response.ok) {
    throw new Error('The panorama request failed. Please try again.')
  }
}

export function createPanoramaSlice(input: PanoramaSliceInput) {
  return requestJson<{ job_id: string }>('/panorama/slice', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(input),
  }).then(({ job_id }) => ({ jobId: job_id }))
}

export function getPanoramaStatus(jobId: string) {
  return requestJson<{ status: string, columns: number, rows: number }>(`/panorama/status?job-id=${encodeURIComponent(jobId)}`)
}

export function getPanoaramaUploadUrl(fileName: string, contentType: string) {
  return requestJson<{ urlSAS: string, blobName: string }>("/panorama/upload", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ file_name: fileName, content_type: contentType })
  })
}

export function uploadToBlob(presignedUrl: string, file: File) {
  return requestBlobUpload(presignedUrl, {
    method: "PUT",
    headers: {
      'x-ms-blob-type': 'BlockBlob',
      'Content-Type': file.type || 'application/octet-stream'
    },
    body: file
  })
}

export function getPanoramaDownloads(jobId: string) {
  return requestJson<{ urlsSAS: string[] }>(`/panorama/download?job-id=${encodeURIComponent(jobId)}`)
}

export function getPanoramaArchiveUrl(jobId: string) {
  return `${apiBaseUrl}/panorama/download-all?job-id=${encodeURIComponent(jobId)}`
}
