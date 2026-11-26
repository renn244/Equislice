import type { CSSProperties, ChangeEventHandler } from 'react'

type PanoramaPreviewProps = {
  columns: number | null
  errorId?: string
  fileName: string | null
  onFileChange: ChangeEventHandler<HTMLInputElement>
  previewUrl: string | null
  processing?: boolean
  rows: number | null
}

type GridStyle = CSSProperties & {
  '--preview-columns'?: number
  '--preview-rows'?: number
}

export function PanoramaPreview({
  columns,
  errorId,
  fileName,
  onFileChange,
  previewUrl,
  processing = false,
  rows,
}: PanoramaPreviewProps) {
  const hasValidGrid = columns !== null && rows !== null
  const gridStyle: GridStyle = hasValidGrid
    ? { '--preview-columns': columns, '--preview-rows': rows }
    : {}

  return (
    <div className={`panorama-preview${previewUrl ? ' panorama-preview-selected' : ''}${processing ? ' is-processing' : ''}`}>
      <label className="panorama-input-label" htmlFor="panorama-image">
        Panorama image
      </label>
      <input
        accept="image/jpeg,image/png"
        aria-describedby={errorId}
        className="panorama-input"
        id="panorama-image"
        name="panorama-image"
        onChange={onFileChange}
        type="file"
      />

      {previewUrl && fileName ? (
        <>
          <img
            alt={`Preview of ${fileName}`}
            className="panorama-preview-image"
            src={previewUrl}
          />
          {hasValidGrid && (
            <>
              <div
                aria-hidden="true"
                className="panorama-preview-grid"
                data-testid="panorama-grid"
                style={gridStyle}
              />
              <div
                aria-hidden="true"
                className="panorama-column-labels"
                style={{ gridTemplateColumns: `repeat(${columns}, 1fr)` }}
              >
                {Array.from({ length: columns }, (_, index) => (
                  <span key={index}>{String.fromCharCode(65 + (index % 26))}</span>
                ))}
              </div>
              <div
                aria-hidden="true"
                className="panorama-row-labels"
                style={{ gridTemplateRows: `repeat(${rows}, 1fr)` }}
              >
                {Array.from({ length: rows }, (_, index) => <span key={index}>{index + 1}</span>)}
              </div>
              {processing && <div aria-hidden="true" className="panorama-processing-line" />}
            </>
          )}
        </>
      ) : (
        <label className="panorama-empty" htmlFor="panorama-image">
          <strong>Upload a panorama</strong>
          <span>Choose a JPG or PNG.</span>
        </label>
      )}
    </div>
  )
}
