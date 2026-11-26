import { UploadIcon } from '@primer/octicons-react'
import type { ChangeEventHandler } from 'react'

type SamplePanorama = {
  dimensions: string
  fileName: string
  path: string
}

const samplePanoramas: SamplePanorama[] = [
  { dimensions: '14506×2809', fileName: 'ridgeline-valley.jpg', path: '/ridgeline-valley.jpg' },
  { dimensions: '14673×3476', fileName: 'alpine-path.jpg', path: '/alpine-path.jpg' },
  { dimensions: '9968×2848', fileName: 'summit-overlook.jpg', path: '/summit-overlook.jpg' },
  { dimensions: '6754×4503', fileName: 'harbor-nightfall.jpg', path: '/harbor-nightfall.jpg' },
]

type EmptyCutterProps = {
  fileError: string | null
  onFileChange: ChangeEventHandler<HTMLInputElement>
  onSample: (sample: SamplePanorama) => Promise<void>
}

export function EmptyCutter({ fileError, onFileChange: handleFileChange, onSample: handleSample }: EmptyCutterProps) {
  return (
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
  )
}
