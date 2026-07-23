const apiBaseUrl = 'http://localhost:3000/api'

type PanoramaSliceInput = {
  file: File
  rows: number
  columns: number
  fileFormat: string
}

async function requestJson<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, init)

  if (!response.ok) {
    throw new Error('The panorama request failed. Please try again.')
  }

  return response.json() as Promise<T>
}

export function createPanoramaSlice(input: PanoramaSliceInput) {
  const body = new FormData()
  body.set('file', input.file)
  body.set('rows', String(input.rows))
  body.set('columns', String(input.columns))
  body.set('fileFormat', input.fileFormat)

  return requestJson<{ job_id: string }>('/panorama/slice', {
    method: 'POST',
    body,
  }).then(({ job_id }) => ({ jobId: job_id }))
}

export function getPanoramaStatus(jobId: string) {
  return requestJson<{ status: string }>(`/panorama/status?job-id=${encodeURIComponent(jobId)}`)
}

export function getPanoramaDownloads(jobId: string) {
  return requestJson<{ urlsSAS: string[] }>(`/panorama/download?job-id=${encodeURIComponent(jobId)}`)
}
