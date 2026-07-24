import { FileIcon, UploadIcon } from '@primer/octicons-react'
import { Button, Heading, Stack, Text } from '@primer/react'
import { useEffect, useState, type ChangeEvent, type CSSProperties } from 'react'
import { createPanoramaSlice, getPanoramaArchiveUrl, getPanoramaDownloads, getPanoramaStatus } from '../lib/panorama-api'
import './cutter-page.css'

type ImageDimensions = {
  height: number
  width: number
}

type JobState = 'idle' | 'submitting' | 'queued' | 'processing' | 'completed' | 'failed'

const minGridValue = 1
const maxGridValue = 100
const cutterJobStorageKey = 'equislice.cutter.job-id'

function isValidGridValue(value: string) {
  const parsedValue = Number(value)
  return Number.isInteger(parsedValue) && parsedValue >= minGridValue && parsedValue <= maxGridValue
}

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function CutterPage() {
  const [columns, setColumns] = useState('6')
  const [dimensions, setDimensions] = useState<ImageDimensions | null>(null)
  const [fileError, setFileError] = useState<string | null>(null)
  const [jobError, setJobError] = useState<string | null>(null)
  const [jobId, setJobId] = useState<string | null>(() => window.localStorage.getItem(cutterJobStorageKey)?.trim() || null)
  const [jobState, setJobState] = useState<JobState>('idle')
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [rows, setRows] = useState('3')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [downloadUrls, setDownloadUrls] = useState<string[]>([])

  const validGrid = isValidGridValue(columns) && isValidGridValue(rows)
  const tileCount = validGrid ? Number(columns) * Number(rows) : null
  const gridError = columns !== '' && rows !== '' && !validGrid
  const tileGridStyle = validGrid
    ? { gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` } as CSSProperties
    : undefined

  useEffect(() => {
    return () => {
      if (previewUrl) URL.revokeObjectURL(previewUrl)
    }
  }, [previewUrl])

  useEffect(() => {
    if (!jobId || jobState === 'completed' || jobState === 'failed') return

    const activeJobId = jobId
    let cancelled = false

    async function checkStatus() {
      try {
        const { status } = await getPanoramaStatus(activeJobId)
        if (cancelled) return

        if (status === 'Completed') {
          const { urlsSAS } = await getPanoramaDownloads(activeJobId)
          if (cancelled) return
          setDownloadUrls(urlsSAS)
          setJobState('completed')
          return
        }

        setJobState(status === 'Processing' ? 'processing' : 'queued')
      } catch {
        if (cancelled) return
        setJobError('We could not check this tile set. Please try again.')
        setJobState('failed')
      }
    }

    void checkStatus()
    const intervalId = window.setInterval(() => void checkStatus(), 1000)

    return () => {
      cancelled = true
      window.clearInterval(intervalId)
    }
  }, [jobId, jobState])

  function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.currentTarget.files?.[0]
    if (!file) return

    if (!file.type.startsWith('image/')) {
      setDimensions(null)
      setFileError('Choose a supported image file.')
      setPreviewUrl(null)
      setSelectedFile(null)
      return
    }

    const objectUrl = URL.createObjectURL(file)
    const image = new Image()

    image.onload = () => {
      setDimensions({ height: image.naturalHeight, width: image.naturalWidth })
      setFileError(null)
      setPreviewUrl(objectUrl)
      setSelectedFile(file)
    }

    image.onerror = () => {
      setDimensions(null)
      setFileError('We could not read that image file.')
      setPreviewUrl(null)
      setSelectedFile(null)
      URL.revokeObjectURL(objectUrl)
    }

    image.src = objectUrl
  }

  async function handleSubmit() {
    if (!selectedFile || !validGrid) return

    setDownloadUrls([])
    setJobError(null)
    setJobId(null)
    setJobState('submitting')

    try {
      const { jobId: createdJobId } = await createPanoramaSlice({
        file: selectedFile,
        rows: Number(rows),
        columns: Number(columns),
        fileFormat: 'jpg',
      })

      setJobId(createdJobId)
      window.localStorage.setItem(cutterJobStorageKey, createdJobId)
      setJobState('queued')
    } catch {
      setJobError('We could not create this tile set. Please try again.')
      setJobState('failed')
    }
  }

  const isSubmitting = jobState === 'submitting' || jobState === 'queued' || jobState === 'processing'

  return (
    <div className="cutter-page">
      <Stack as="header" className="cutter-header" direction="horizontal" align="center" justify="space-between">
        <a className="cutter-brand" href="/">EquiSlice</a>
        <a className="cutter-back" href="/">Back to overview</a>
      </Stack>

      <main className="cutter-main">
        <Stack className="cutter-intro" gap="condensed">
          <Heading as="h1">Create your tile set</Heading>
          <Text as="p" size="large">
            Upload one equirectangular panorama, choose the grid, and prepare it for slicing.
          </Text>
        </Stack>

        <div className="cutter-workspace">
          <section className="cutter-panel" aria-labelledby="upload-title">
            <Stack gap="normal">
              <Stack gap="tight">
                <Heading as="h2" id="upload-title">1. Upload panorama</Heading>
                <Text as="p">JPG, PNG, WebP, or another browser-readable image format.</Text>
              </Stack>

              <label className="upload-control" htmlFor="panorama-image">
                <UploadIcon aria-hidden="true" size={28} />
                <span>Panorama image</span>
                <input
                  accept="image/*"
                  id="panorama-image"
                  name="panorama-image"
                  onChange={handleFileChange}
                  type="file"
                />
              </label>
              {fileError && <Text as="p" className="cutter-error" role="alert">{fileError}</Text>}

              {selectedFile && dimensions && (
                <>
                  <div className="file-summary" aria-live="polite">
                    <FileIcon aria-hidden="true" size={20} />
                    <div>
                      <strong>{selectedFile.name}</strong>
                      <Text as="span">{dimensions.width} × {dimensions.height} px · {formatFileSize(selectedFile.size)}</Text>
                    </div>
                    <Button onClick={() => { setDimensions(null); setPreviewUrl(null); setSelectedFile(null); setDownloadUrls([]); setJobError(null); setJobId(null); setJobState('idle') }} variant="invisible">
                      Replace
                    </Button>
                  </div>
                  {previewUrl && (
                      <div className="cutter-panorama-preview">
                        <img alt={`Preview of ${selectedFile.name}`} src={previewUrl} />
                        {tileGridStyle && (
                            <div aria-hidden="true" className="cutter-slice-overlay" style={tileGridStyle}>
                              {Array.from({ length: tileCount ?? 0 }, (_, index) => <span key={index} />)}
                            </div>
                        )}
                      </div>
                  )}
                </>
              )}
            </Stack>
          </section>

          <section className="cutter-panel" aria-labelledby="settings-title">
            <Stack gap="normal">
              <Stack gap="tight">
                <Heading as="h2" id="settings-title">2. Set the grid</Heading>
                <Text as="p">Tiles use the fixed naming pattern <code>tile_r01_c01.jpg</code>.</Text>
              </Stack>

              <div className="grid-fields">
                <label htmlFor="columns">Columns
                  <input id="columns" max={maxGridValue} min={minGridValue} onChange={(event) => setColumns(event.target.value)} type="number" value={columns} />
                </label>
                <label htmlFor="rows">Rows
                  <input id="rows" max={maxGridValue} min={minGridValue} onChange={(event) => setRows(event.target.value)} type="number" value={rows} />
                </label>
              </div>
              {gridError && (
                <Text as="p" className="cutter-error" role="alert">
                  Rows and columns must be whole numbers from 1 to 100.
                </Text>
              )}

              <div className="grid-summary" aria-live="polite">
                <span>Tile set</span>
                <strong>{tileCount === null ? 'Choose a valid grid' : `${tileCount} tiles`}</strong>
              </div>
            </Stack>
          </section>
        </div>

        <section className="cutter-submit" aria-label="Create tile set">
          <Stack gap="tight">
            <Heading as="h2">Ready to slice?</Heading>
            <Text as="p">Uploads and exports are automatically deleted after 24 hours.</Text>
          </Stack>
          <Button disabled={!selectedFile || !validGrid || isSubmitting} onClick={handleSubmit} size="large" type="button" variant="primary">
            {jobState === 'submitting' ? 'Creating tiles…' : 'Create tiles'}
          </Button>
        </section>

        {(jobState !== 'idle' || jobError) && (
          <section className="cutter-job" aria-live="polite">
            {jobState === 'submitting' && <Text as="p" role="status">Creating job…</Text>}
            {jobState === 'queued' && <Text as="p" role="status">Job queued</Text>}
            {jobState === 'processing' && <Text as="p" role="status">Job processing</Text>}
            {jobState === 'completed' && <Text as="p" role="status">Job completed</Text>}
            {jobError && <Text as="p" className="cutter-error" role="alert">{jobError}</Text>}
            {downloadUrls.length > 0 && (
              <div className="cutter-tile-results">
                <Text as="p">Your panorama has been sliced into {downloadUrls.length} real tiles.</Text>
                <div className="cutter-tile-gallery" style={tileGridStyle}>
                  {downloadUrls.map((url, index) => (
                    <a href={url} key={url} rel="noreferrer" target="_blank">
                      <img alt={`Sliced tile ${index + 1}`} loading="lazy" src={url} />
                    </a>
                  ))}
                </div>
                {jobId && (
                  <a className="cutter-download-all" href={getPanoramaArchiveUrl(jobId)}>
                    Download all tiles (.zip)
                  </a>
                )}
              </div>
            )}
          </section>
        )}
      </main>
    </div>
  )
}
