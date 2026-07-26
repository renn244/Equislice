import {
  CheckCircleFillIcon,
  DownloadIcon,
  HourglassIcon,
  SyncIcon,
  UploadIcon,
  XIcon,
} from '@primer/octicons-react'
import { useEffect, useState, type ChangeEvent, type CSSProperties } from 'react'
import { SiteHeader } from '../components/site-header'
import {
  createPanoramaSlice,
  getPanoramaArchiveUrl,
  getPanoramaDownloads,
  getPanoramaStatus,
} from '../lib/panorama-api'
import './cutter-page.css'
import { PanoramaPreview } from './panorama-preview'

type ImageDimensions = {
  height: number
  width: number
}

type JobState = 'idle' | 'submitting' | 'queued' | 'processing' | 'completed' | 'failed'

type SamplePanorama = {
  dimensions: string
  fileName: string
  path: string
}

const minGridValue = 1
const maxGridValue = 100
const cutterJobStorageKey = 'equislice.cutter.job-id'

const samplePanoramas: SamplePanorama[] = [
  { dimensions: '14506×2809', fileName: 'ridgeline-valley.jpg', path: '/ridgeline-valley.jpg' },
  { dimensions: '14673×3476', fileName: 'alpine-path.jpg', path: '/alpine-path.jpg' },
  { dimensions: '9968×2848', fileName: 'summit-overlook.jpg', path: '/summit-overlook.jpg' },
  { dimensions: '6754×4503', fileName: 'harbor-nightfall.jpg', path: '/harbor-nightfall.jpg' },
]

const workflowStates: { key: JobState; label: string }[] = [
  { key: 'submitting', label: 'Uploading' },
  { key: 'queued', label: 'Queued' },
  { key: 'processing', label: 'Slicing' },
  { key: 'completed', label: 'Complete' },
]

function isValidGridValue(value: string) {
  const parsedValue = Number(value)
  return Number.isInteger(parsedValue) && parsedValue >= minGridValue && parsedValue <= maxGridValue
}

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function getWorkflowIndex(state: JobState) {
  return workflowStates.findIndex(({ key }) => key === state)
}

export function CutterPage() {
  const [columns, setColumns] = useState('4')
  const [dimensions, setDimensions] = useState<ImageDimensions | null>(null)
  const [fileError, setFileError] = useState<string | null>(null)
  const [jobError, setJobError] = useState<string | null>(null)
  const [jobId, setJobId] = useState<string | null>(
    () => window.localStorage.getItem(cutterJobStorageKey)?.trim() || null,
  )
  const [jobState, setJobState] = useState<JobState>('idle')
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [rows, setRows] = useState('3')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)

  const [downloadUrls, setDownloadUrls] = useState<string[]>([])
  const [downloadColumns, setDownloadColumns] = useState<number | null>(null)

  const validGrid = isValidGridValue(columns) && isValidGridValue(rows)
  const parsedColumns = validGrid ? Number(columns) : null
  const parsedRows = validGrid ? Number(rows) : null
  const tileCount = validGrid ? Number(columns) * Number(rows) : null
  const gridError = columns !== '' && rows !== '' && !validGrid
  const tileGridStyle = validGrid
    ? ({ gridTemplateColumns: `repeat(${downloadColumns}, minmax(0, 1fr))` } as CSSProperties)
    : undefined
  const isWorking = jobState === 'submitting' || jobState === 'queued' || jobState === 'processing'
  const workflowIndex = getWorkflowIndex(jobState)
  const tileWidth = dimensions && parsedColumns ? Math.floor(dimensions.width / parsedColumns) : null
  const tileHeight = dimensions && parsedRows ? Math.floor(dimensions.height / parsedRows) : null

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
          setDownloadColumns(parsedColumns)
          setDownloadRows(parsedRows)
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

  function prepareFile(file: File) {
    if (file.type !== 'image/jpeg' && file.type !== 'image/png') {
      setDimensions(null)
      setFileError('Choose a JPG or PNG image.')
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

  function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.currentTarget.files?.[0]
    if (file) prepareFile(file)
  }

  async function handleSample(sample: SamplePanorama) {
    try {
      const response = await fetch(sample.path)
      const blob = await response.blob()
      prepareFile(new File([blob], sample.fileName, { type: 'image/jpeg' }))
    } catch {
      setFileError('We could not load that sample panorama.')
    }
  }

  function resetSelectedFile() {
    setDimensions(null)
    setPreviewUrl(null)
    setSelectedFile(null)
    setFileError(null)
  }

  function stepGrid(target: 'columns' | 'rows', direction: -1 | 1) {
    const current = Number(target === 'columns' ? columns : rows)
    const next = Math.min(maxGridValue, Math.max(minGridValue, current + direction))
    if (target === 'columns') setColumns(String(next))
    else setRows(String(next))
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

  return (
    <div className="cutter-page">
      <SiteHeader compact />

      <main className="cutter-main">
        {jobId && !selectedFile && (
          <div className="restored-session" role="status">
            <CheckCircleFillIcon aria-hidden="true" size={16} />
            <p>
              <strong>Session restored.</strong> Checking the job saved on this device.
            </p>
            <code>{jobId}</code>
          </div>
        )}

        {!selectedFile ? (
          <section className="empty-cutter" aria-labelledby="empty-cutter-title">
            <p className="es-kicker"><span /> New job</p>
            <h1 id="empty-cutter-title">Add a panorama to begin.</h1>
            <p className="empty-cutter-copy">
              Drop in one equirectangular 360° image. We’ll read its dimensions, then you decide
              how it divides into tiles.
            </p>

            <label className="precision-dropzone" htmlFor="panorama-image">
              <input
                accept="image/jpeg,image/png"
                aria-describedby={fileError ? 'panorama-file-error' : undefined}
                aria-label="Panorama image"
                id="panorama-image"
                name="panorama-image"
                onChange={handleFileChange}
                type="file"
              />
              <span className="dropzone-grid-icon" aria-hidden="true">
                {Array.from({ length: 8 }, (_, index) => <i key={index} />)}
              </span>
              <strong>Drag &amp; drop your panorama here</strong>
              <span>or <em>browse your files</em></span>
              <small>JPG · PNG · up to 60 MB</small>
            </label>
            {fileError && (
              <p className="cutter-error" id="panorama-file-error" role="alert">{fileError}</p>
            )}

            <div className="empty-cutter-actions">
              <label className="browse-button" htmlFor="panorama-image">
                <UploadIcon aria-hidden="true" size={15} />
                Browse files
              </label>
              <a href="/">← Back to overview</a>
            </div>

            <div className="sample-section">
              <h2>No panorama handy?</h2>
              <p>Start from one of these sample panoramas to see the whole flow.</p>
              <div className="sample-grid">
                {samplePanoramas.map((sample) => (
                  <button
                    aria-label={`${sample.fileName} ${sample.dimensions}`}
                    key={sample.fileName}
                    onClick={() => void handleSample(sample)}
                    type="button"
                  >
                    <img alt="" src={sample.path} />
                    <span>
                      <strong>{sample.fileName}</strong>
                      <small>{sample.dimensions}</small>
                    </span>
                  </button>
                ))}
              </div>
            </div>
          </section>
        ) : (
          <div className="configured-cutter">
            <section className="instrument-card preview-card" aria-labelledby="preview-card-title">
              <div className="instrument-card-heading">
                <div>
                  <p className="es-kicker"><span /> Preview · live grid</p>
                  <p id="preview-card-title">
                    This is exactly how <strong>{columns}×{rows}</strong> will divide your panorama.
                  </p>
                </div>
                <div className="preview-actions">
                  <label htmlFor="panorama-replacement">Replace image</label>
                  <input
                    accept="image/jpeg,image/png"
                    id="panorama-replacement"
                    onChange={handleFileChange}
                    type="file"
                  />
                  <button onClick={resetSelectedFile} type="button">
                    <XIcon aria-hidden="true" size={13} />
                    Remove
                  </button>
                </div>
              </div>

              <PanoramaPreview
                columns={parsedColumns}
                fileName={selectedFile.name}
                onFileChange={handleFileChange}
                previewUrl={previewUrl}
                processing={jobState === 'processing'}
                rows={parsedRows}
              />
            </section>

            <section className="instrument-card" aria-labelledby="grid-title">
              <p className="es-kicker" id="grid-title"><span /> Grid</p>
              <div className="instrument-grid-controls">
                <label htmlFor="columns">
                  <span>Columns</span>
                  <div className="instrument-stepper">
                    <button aria-label="Decrease Columns" onClick={() => stepGrid('columns', -1)} type="button">−</button>
                    <input
                      aria-label="Columns"
                      id="columns"
                      max={maxGridValue}
                      min={minGridValue}
                      onChange={(event) => setColumns(event.target.value)}
                      type="number"
                      value={columns}
                    />
                    <button aria-label="Increase Columns" onClick={() => stepGrid('columns', 1)} type="button">+</button>
                  </div>
                </label>
                <label htmlFor="rows">
                  <span>Rows</span>
                  <div className="instrument-stepper">
                    <button aria-label="Decrease Rows" onClick={() => stepGrid('rows', -1)} type="button">−</button>
                    <input
                      aria-label="Rows"
                      id="rows"
                      max={maxGridValue}
                      min={minGridValue}
                      onChange={(event) => setRows(event.target.value)}
                      type="number"
                      value={rows}
                    />
                    <button aria-label="Increase Rows" onClick={() => stepGrid('rows', 1)} type="button">+</button>
                  </div>
                </label>
              </div>
              {gridError && (
                <p className="cutter-error" role="alert">
                  Rows and columns must be whole numbers from 1 to 100.
                </p>
              )}
              <div className="instrument-total">
                <span>Total tiles</span>
                <strong>{tileCount ?? '—'}</strong>
              </div>
            </section>

            <section className="instrument-card" aria-labelledby="format-title">
              <p className="es-kicker" id="format-title"><span /> Output format</p>
              <div className="format-grid">
                <button disabled type="button">
                  <strong>PNG</strong>
                  <span>Not available in worker</span>
                </button>
                <button aria-pressed="true" className="is-selected" type="button">
                  <strong>JPG</strong>
                  <span>Worker output · photographic</span>
                  <i>✓</i>
                </button>
                <button disabled type="button">
                  <strong>WebP</strong>
                  <span>Not available in worker</span>
                </button>
              </div>
            </section>

            <section className="instrument-card file-details" aria-labelledby="details-title">
              <p className="es-kicker" id="details-title"><span /> File details</p>
              <dl>
                <div><dt>Name</dt><dd>{selectedFile.name}</dd></div>
                <div><dt>Dimensions</dt><dd>{dimensions ? `${dimensions.width} × ${dimensions.height}` : 'Reading…'}</dd></div>
                <div><dt>Type</dt><dd>{selectedFile.type === 'image/png' ? 'PNG' : 'JPG'}</dd></div>
                <div><dt>Size</dt><dd>{formatFileSize(selectedFile.size)}</dd></div>
                <div><dt>Tile size</dt><dd>{tileWidth && tileHeight ? `${tileWidth} × ${tileHeight}` : '—'}</dd></div>
              </dl>
            </section>

            <button
              className="begin-slicing-button"
              disabled={!validGrid || isWorking}
              onClick={handleSubmit}
              type="button"
            >
              {isWorking ? (
                <>
                  <SyncIcon className="spin-icon" aria-hidden="true" size={15} />
                  {jobState === 'submitting' ? 'Uploading panorama' : 'Slicing in progress'}
                </>
              ) : 'Begin slicing →'}
            </button>
            <p className="retention-note">Uploads and exports are automatically deleted after 24 hours.</p>
          </div>
        )}

        {(jobState !== 'idle' || jobError) && (
          <section className={`job-panel job-panel-${jobState}`} aria-live="polite">
            <div className="job-panel-heading">
              <div>
                <p className="es-kicker"><span /> Live job</p>
                <h2>
                  {jobState === 'completed'
                    ? 'The grid is ready.'
                    : jobState === 'failed'
                      ? 'The job needs attention.'
                      : 'EquiSlice is working.'}
                </h2>
              </div>
              {jobId && <code title={jobId}>{jobId}</code>}
            </div>

            {jobState !== 'failed' && (
              <ol className="job-progress" aria-label="Job progress">
                {workflowStates.map(({ key, label }, index) => {
                  const isComplete = workflowIndex > index || jobState === 'completed'
                  const isActive = key === jobState
                  return (
                    <li
                      className={`${isComplete ? 'is-complete' : ''}${isActive ? ' is-active' : ''}`}
                      key={key}
                    >
                      <span>
                        {isComplete ? (
                          <CheckCircleFillIcon aria-hidden="true" size={16} />
                        ) : isActive ? (
                          <HourglassIcon aria-hidden="true" size={16} />
                        ) : String(index + 1).padStart(2, '0')}
                      </span>
                      <strong>{label}</strong>
                    </li>
                  )
                })}
              </ol>
            )}

            {jobError && <p className="cutter-error job-error" role="alert">{jobError}</p>}

            {downloadUrls.length > 0 && (
              <div className="cutter-tile-results">
                <div className="result-heading">
                  <div>
                    <p className="es-kicker"><span /> Finished grid</p>
                    <h3>{downloadUrls.length} aligned tiles.</h3>
                    <p>Each image below is loaded from the real Azure Storage output.</p>
                  </div>
                  {jobId && (
                    <a className="download-all-button" href={getPanoramaArchiveUrl(jobId)}>
                      <DownloadIcon aria-hidden="true" size={15} />
                      Download all tiles (.zip)
                    </a>
                  )}
                </div>

                <div className="cutter-tile-gallery" style={tileGridStyle}>
                  {downloadUrls.map((url, index) => (
                    <a href={url} key={`${url}-${index}`} rel="noreferrer" target="_blank">
                      <img alt={`Sliced tile ${index + 1}`} loading="lazy" src={url} />
                      <span>
                        {String.fromCharCode(65 + (index % Number(columns)))}
                        {Math.floor(index / Number(columns)) + 1}
                      </span>
                    </a>
                  ))}
                </div>
              </div>
            )}
          </section>
        )}
      </main>
    </div>
  )
}
