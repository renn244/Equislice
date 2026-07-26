import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const cutterStyles = readFileSync(resolve(process.cwd(), 'src/pages/cutter-page.css'), 'utf8')

describe('cutter page styles', () => {
  it('uses the shared Precision Instrument tokens instead of a separate palette', () => {
    expect(cutterStyles).toContain('color: var(--eq-ink)')
    expect(cutterStyles).toContain('var(--eq-teal)')
    expect(cutterStyles).toContain('var(--eq-amber)')
    expect(cutterStyles).toContain('var(--eq-canvas)')
    expect(cutterStyles).not.toContain('--cutter-night')
  })
})