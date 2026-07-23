import { UploadIcon } from '@primer/octicons-react'
import type { CSSProperties, ChangeEventHandler } from 'react'

type PanoramaPreviewProps = {
  columns: number | null
  errorId?: string
  fileName: string | null
  onFileChange: ChangeEventHandler<HTMLInputElement>
  previewUrl: string | null
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
  rows,
}: PanoramaPreviewProps) {
  const hasValidGrid = columns !== null && rows !== null
  const gridStyle: GridStyle = hasValidGrid
    ? { '--preview-columns': columns, '--preview-rows': rows }
    : {}

  return (
    <div className={`panorama-preview${previewUrl ? ' panorama-preview-selected' : ''}`}>
      <label className="panorama-input-label" htmlFor="panorama-image">
        Panorama image
      </label>
      <input
        accept="image/*"
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
            <div
              aria-hidden="true"
              className="panorama-preview-grid"
              data-testid="panorama-grid"
              style={gridStyle}
            />
          )}
        </>
      ) : (
        <label className="panorama-empty" htmlFor="panorama-image">
          <UploadIcon aria-hidden="true" size={30} />
          <strong>Upload a panorama</strong>
          <span>Choose a browser-readable JPG, PNG, WebP, or other image.</span>
        </label>
      )}
    </div>
  )
}
