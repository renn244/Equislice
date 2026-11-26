import * as Sentry from '@sentry/react'
import { useEffect, useRef, useState } from 'react'
import {
  createPanoramaSlice,
  getPanoramaUploadUrl,
  getPanoramaDownloads,
  getPanoramaStatus,
  uploadToBlob,
} from '../api/panorama.api'
import { resolveFileNameTemplate, type CutterFormValues } from '../forms/cutterForm.schema'
import type { JobState, TileDownload } from '../types/panorama.type'

const jobStorageKey = 'equislice.cutter.job-id'

function readSavedJob() {
  try {
    return window.localStorage.getItem(jobStorageKey)?.trim() || null
  } catch {
    return null
  }
}

export function usePanoramaJob() {
  const [jobId, setJobId] = useState<string | null>(readSavedJob)
  const [jobState, setJobState] = useState<JobState>(jobId ? 'queued' : 'idle')
  const [jobError, setJobError] = useState<string | null>(null)
  const [downloads, setDownloads] = useState<TileDownload[]>([])
  const [downloadColumns, setDownloadColumns] = useState<number | null>(null)
  const generation = useRef(0)
  const submitting = useRef(false)

  useEffect(() => {
    if (!jobId) return
    const activeJobId = jobId
    const version = generation.current
    let cancelled = false
    let timer: ReturnType<typeof setTimeout> | undefined
    const isStale = () => cancelled || version !== generation.current

    async function checkStatus() {
      try {
        const { status, columns } = await getPanoramaStatus(activeJobId)
        if (isStale()) return
        setJobError(null)
        if (status === 'Completed') {
          const { tiles } = await getPanoramaDownloads(activeJobId)
          if (isStale()) return
          setDownloads(tiles)
          setDownloadColumns(columns)
          setJobState('completed')
          Sentry.metrics.count('equislice.panorama_slice.completed')
          return
        }
        if (status === 'Failed') {
          setJobError('The tile set processing failed. Please try again with a different image.')
          setJobState('failed')
          Sentry.metrics.count('equislice.panorama_slice.failed')
          return
        }
        setJobState(status === 'Processing' ? 'processing' : 'queued')
      } catch {
        if (isStale()) return
        setJobError('We could not check this tile set. Please try again.')
        Sentry.metrics.count('equislice.panorama_status_check.failed')
      }
      if (!isStale()) timer = setTimeout(() => void checkStatus(), 1000)
    }

    void checkStatus()
    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [jobId])

  async function submit(file: File, values: CutterFormValues) {
    if (submitting.current) return
    submitting.current = true
    generation.current++
    setDownloads([])
    setDownloadColumns(null)
    setJobError(null)
    setJobId(null)
    setJobState('submitting')

    Sentry.metrics.count('equislice.panorama_slice.started')
    Sentry.logger.info('Panorama slicing started', {
      attributes: { 'panorama.columns': Number(values.columns), 'panorama.rows': Number(values.rows) },
    })

    try {
      const { urlSAS, blobName } = await getPanoramaUploadUrl(file.name, file.type)
      await uploadToBlob(urlSAS, file)
      const { jobId: createdJobId } = await createPanoramaSlice({
        file: blobName,
        rows: Number(values.rows),
        columns: Number(values.columns),
        file_formats: 'jpg',
        file_name_format: resolveFileNameTemplate(values.fileNameTemplate),
      })
      setJobId(createdJobId)
      setJobState('queued')
      try {
        window.localStorage.setItem(jobStorageKey, createdJobId)
      } catch {
        // The job still works when the browser does not permit local storage.
      }
    } catch {
      setJobError('We could not create this tile set. Please try again.')
      setJobState('failed')
      Sentry.metrics.count('equislice.panorama_slice.submit_failed')
      Sentry.logger.error('Panorama slicing submission failed')
    } finally {
      submitting.current = false
    }
  }

  return { jobId, jobState, jobError, downloads, downloadColumns, submit }
}
