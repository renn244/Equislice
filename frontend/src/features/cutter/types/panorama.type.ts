export type ImageDimensions = {
  height: number
  width: number
}

export type JobState = 'idle' | 'submitting' | 'queued' | 'processing' | 'completed' | 'failed'

export type PanoramaSliceInput = {
  file: string
  rows: number
  columns: number
  file_formats: string
  file_name_format: string
}

export type TileDownload = {
  urlSAS: string
  fileName: string
  row: number
  column: number
}
