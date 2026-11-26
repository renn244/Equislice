import { CheckCircleFillIcon, DownloadIcon, HourglassIcon, SyncIcon } from '@primer/octicons-react'
import { getPanoramaArchiveUrl } from '../api/panorama.api'
import { useTileDownloads } from '../hooks/useTileDownloads'
import type { JobState, TileDownload } from '../types/panorama.type'

const workflowStates: { key: JobState; label: string }[] = [
  { key: 'submitting', label: 'Uploading' },
  { key: 'queued', label: 'Queued' },
  { key: 'processing', label: 'Slicing' },
  { key: 'completed', label: 'Complete' },
]

type PanoramaJobPanelProps = {
  jobId: string | null
  jobState: JobState
  jobError: string | null
  downloads: TileDownload[]
  downloadColumns: number | null
}

export function PanoramaJobPanel({ jobId, jobState, jobError, downloads, downloadColumns }: PanoramaJobPanelProps) {
  const { downloadingTileKey, tileDownloadError, download: handleTileDownload } = useTileDownloads()
  const workflowIndex = workflowStates.findIndex(({ key }) => key === jobState)
  const tileGridStyle = downloadColumns ? { gridTemplateColumns: `repeat(${downloadColumns}, minmax(0, 1fr))` } : undefined

  return (
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

      {downloads.length > 0 && (
        <div className="cutter-tile-results">
          <div className="result-heading">
            <div>
              <p className="es-kicker"><span /> Finished grid</p>
              <h3>{downloads.length} aligned tiles.</h3>
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
            {downloads.map((tile) => {
              const tileKey = `${tile.row}-${tile.column}`
              const isDownloading = downloadingTileKey === tileKey

              return (
                <div className="cutter-tile" key={tileKey}>
                  <a href={tile.urlSAS} rel="noreferrer" target="_blank" title={`Preview ${tile.fileName}`}>
                    <img alt={`Sliced tile ${tile.row}, ${tile.column}`} loading="lazy" src={tile.urlSAS} />
                    <span>{tile.fileName}</span>
                  </a>
                  <button
                    aria-label={`Download ${tile.fileName}`}
                    className="tile-download-button"
                    disabled={isDownloading}
                    onClick={() => void handleTileDownload(tile)}
                    type="button"
                  >
                    {isDownloading ? <SyncIcon aria-hidden="true" className="spin-icon" size={13} /> : <DownloadIcon aria-hidden="true" size={13} />}
                  </button>
                </div>
              )
            })}
          </div>
          {tileDownloadError && <p className="cutter-error tile-download-error" role="alert">{tileDownloadError}</p>}
        </div>
      )}
    </section>
  )
}
