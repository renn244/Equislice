import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const cutterStyles = readFileSync(resolve(process.cwd(), 'src/pages/cutter-page.css'), 'utf8')

describe('cutter page styles', () => {
  it('uses the landing page design tokens instead of a separate dark palette', () => {
    expect(cutterStyles).toContain('background: var(--eq-canvas)')
    expect(cutterStyles).toContain('color: var(--eq-ink)')
    expect(cutterStyles).toContain('background: var(--eq-surface)')
    expect(cutterStyles).not.toContain('--cutter-night')
  })
})
