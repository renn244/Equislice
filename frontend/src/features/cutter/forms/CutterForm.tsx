import { SyncIcon, XIcon } from '@primer/octicons-react'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import type { ChangeEventHandler } from 'react'
import { PanoramaPreview } from '../components/PanoramaPreview'
import type { ImageDimensions, JobState } from '../types/panorama.type'
import {
  CutterSchema,
  cutterDefaults,
  defaultFileNameTemplate,
  minGridValue,
  maxGridValue,
  isValidGridValue,
  resolveFileNameTemplate,
  type CutterFormValues,
} from './cutterForm.schema'

function previewFileName(template: string) {
  return `${template.replaceAll('{row}', '1').replaceAll('{col}', '1')}.jpg`
}

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

type CutterFormProps = {
  selectedFile: File
  dimensions: ImageDimensions | null
  previewUrl: string | null
  fileError: string | null
  jobState: JobState
  onFileChange: ChangeEventHandler<HTMLInputElement>
  onRemoveFile: () => void
  onSubmit: (values: CutterFormValues) => Promise<void>
}

export function CutterForm({
  selectedFile,
  dimensions,
  previewUrl,
  fileError,
  jobState,
  onFileChange: handleFileChange,
  onRemoveFile: resetSelectedFile,
  onSubmit,
}: CutterFormProps) {
  const { register, watch, setValue, handleSubmit, formState: { isSubmitting } } = useForm<CutterFormValues>({
    resolver: zodResolver(CutterSchema),
    defaultValues: cutterDefaults,
    mode: 'onChange',
  })
  const { columns, rows, fileNameTemplate } = watch()
  const validGrid = isValidGridValue(columns) && isValidGridValue(rows)
  const resolvedFileNameTemplate = resolveFileNameTemplate(fileNameTemplate)
  const invalidFileNameTemplate = !CutterSchema.shape.fileNameTemplate.safeParse(fileNameTemplate).success
  const parsedColumns = validGrid ? Number(columns) : null
  const parsedRows = validGrid ? Number(rows) : null
  const tileCount = validGrid ? Number(columns) * Number(rows) : null
  const gridError = columns !== '' && rows !== '' && !validGrid
  const isWorking = jobState === 'submitting' || jobState === 'queued' || jobState === 'processing'
  const tileWidth = dimensions && parsedColumns ? Math.floor(dimensions.width / parsedColumns) : null
  const tileHeight = dimensions && parsedRows ? Math.floor(dimensions.height / parsedRows) : null

  function stepGrid(target: 'columns' | 'rows', direction: -1 | 1) {
    const current = Number(target === 'columns' ? columns : rows)
    const next = Math.min(maxGridValue, Math.max(minGridValue, current + direction))
    setValue(target, String(next), { shouldValidate: true, shouldDirty: true })
  }

  return (
    <form className="configured-cutter" noValidate onSubmit={handleSubmit(onSubmit)}>
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

        {fileError && <p className="cutter-error" role="alert">{fileError}</p>}
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
                aria-invalid={!isValidGridValue(columns)}
                id="columns"
                max={maxGridValue}
                min={minGridValue}
                {...register('columns')}
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
                aria-invalid={!isValidGridValue(rows)}
                id="rows"
                max={maxGridValue}
                min={minGridValue}
                {...register('rows')}
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
        <div className="file-name-template">
          <label htmlFor="file-name-template">
            <span>Tile naming</span>
            <input
              aria-describedby="file-name-template-help"
              aria-invalid={invalidFileNameTemplate}
              id="file-name-template"
              {...register('fileNameTemplate')}
              placeholder={defaultFileNameTemplate}
              spellCheck="false"
              type="text"
              value={fileNameTemplate}
            />
          </label>
          <p id="file-name-template-help">
            Optional. Leave blank for <code>{defaultFileNameTemplate}</code>, or use both <code>{'{row}'}</code> and <code>{'{col}'}</code>. The JPG extension is added automatically.
          </p>
          {!invalidFileNameTemplate ? (
            <p className="file-name-preview">First tile: <code>{previewFileName(resolvedFileNameTemplate)}</code></p>
          ) : (
            <p className="cutter-error" role="alert">Include both <code>{'{row}'}</code> and <code>{'{col}'}</code>.</p>
          )}
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
        disabled={!validGrid || invalidFileNameTemplate || isWorking || isSubmitting}
        type="submit"
      >
        {isWorking ? (
          <>
            <SyncIcon className="spin-icon" aria-hidden="true" size={15} />
            {jobState === 'submitting' ? 'Uploading panorama' : 'Slicing in progress'}
          </>
        ) : 'Begin slicing →'}
      </button>
      <p className="retention-note">Uploads and exports are automatically deleted after 24 hours.</p>
    </form>
  )
}
