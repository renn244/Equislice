import { z } from 'zod'

export const minGridValue = 1
export const maxGridValue = 100
export const defaultFileNameTemplate = 'tile_r{row}_c{col}'

export function isValidGridValue(value: string) {
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed >= minGridValue && parsed <= maxGridValue
}

export function resolveFileNameTemplate(template: string) {
  return template.trim() || defaultFileNameTemplate
}

const gridValue = z.string().refine(isValidGridValue, 'Rows and columns must be whole numbers from 1 to 100.')

export const CutterSchema = z.object({
  columns: gridValue,
  rows: gridValue,
  fileNameTemplate: z.string().refine(value => {
    const template = resolveFileNameTemplate(value)
    return template.includes('{row}') && template.includes('{col}')
  }, 'Include both {row} and {col}.'),
})

export type CutterFormValues = z.infer<typeof CutterSchema>

export const cutterDefaults: CutterFormValues = {
  columns: '4',
  rows: '3',
  fileNameTemplate: '',
}
