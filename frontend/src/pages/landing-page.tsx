import { SiteHeader } from '../components/site-header'
import type { CSSProperties } from 'react'
import './landing-page.css'

const workflow = [
  ['Upload', 'Drop in one equirectangular 360° panorama. We read its true dimensions instantly.'],
  ['Configure', 'Choose rows and columns. A live grid shows exactly how your image will divide.'],
  ['Slice', 'Watch the panorama split, tile by tile, into a perfectly aligned set.'],
  ['Download', 'Grab a single tile or the whole grid — order and alignment preserved.'],
]

const reasons = [
  ['Alignment you can trust', 'Every tile is cut on an exact fraction of the source. Reassemble them and the seam disappears.'],
  ['Built for large sources', 'Panoramas run wide — tens of thousands of pixels. EquiSlice keeps the full resolution in every slice.'],
  ['Nothing to relearn', 'One panorama in, a labeled grid out. No layers, no timelines, no dashboard to navigate.'],
]

const heroTiles = Array.from({ length: 8 }, (_, index) => ({
  column: index % 4,
  label: `${String.fromCharCode(65 + (index % 4))}${Math.floor(index / 4) + 1}`,
  row: Math.floor(index / 4),
}))

export function LandingPage() {
  return (
    <div className="landing-page">
      <SiteHeader />

      <main>
        <section className="precision-hero" aria-labelledby="landing-title">
          <p className="es-kicker"><span /> Equirectangular · 360° · Tile engine</p>
          <h1 id="landing-title">
            <span>One panorama,</span>
            <span>perfectly sliced.</span>
          </h1>
          <p className="precision-hero-copy">
            EquiSlice turns a single 360° panorama into an evenly cut tile grid — labeled,
            aligned, and ready to download in the format you need.
          </p>
          <div className="precision-hero-actions">
            <a className="es-primary-button" href="/cutter">Slice a panorama →</a>
            <a className="es-text-link" href="#how">See how it works</a>
          </div>

          <div className="hero-grid-frame">
            <div className="hero-grid-meta">
              <span>source.jpg · 14506×2809</span>
              <span>4 × 2 grid</span>
            </div>
            <div
              className="hero-slicing-grid"
              role="img"
              aria-label="A wide mountain valley panorama repeatedly slicing itself into an aligned tile grid and reassembling"
            >
              {heroTiles.map(({ column, label, row }) => (
                <div
                  className="hero-slicing-tile"
                  key={label}
                  style={{
                    '--tile-column': column,
                    '--tile-row': row,
                  } as CSSProperties}
                >
                  <span>{label}</span>
                </div>
              ))}
              <i className="hero-scan-line" aria-hidden="true" />
            </div>
          </div>
        </section>

        <section className="precision-section workflow-section" id="how" aria-labelledby="workflow-title">
          <div className="precision-shell">
            <p className="es-kicker"><span /> The workflow</p>
            <h2 id="workflow-title">Four steps from full panorama to finished grid.</h2>
            <div className="precision-workflow">
              {workflow.map(([title, description], index) => (
                <article key={title}>
                  <span>{String(index + 1).padStart(2, '0')} / 04</span>
                  <h3>{title}</h3>
                  <p>{description}</p>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section className="precision-section why-section" id="why" aria-labelledby="why-title">
          <div className="precision-shell">
            <div className="why-heading">
              <div>
                <p className="es-kicker"><span /> Why EquiSlice</p>
                <h2 id="why-title">A focused instrument, not a dashboard.</h2>
              </div>
              <p>Everything here serves one job: dividing a panorama cleanly and giving the pieces back to you in order.</p>
            </div>
            <div className="reason-grid">
              {reasons.map(([title, description], index) => (
                <article key={title}>
                  <span>{String(index + 1).padStart(2, '0')}</span>
                  <h3>{title}</h3>
                  <p>{description}</p>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section className="precision-cta" aria-labelledby="cta-title">
          <div className="precision-shell">
            <p className="es-kicker"><span /> Ready when you are</p>
            <h2 id="cta-title">Turn one panorama into a perfectly organized tile set.</h2>
            <a className="es-primary-button" href="/cutter">Open the cutter →</a>
          </div>
        </section>
      </main>

      <footer className="precision-footer">
        <div className="precision-shell">
          <a className="site-brand" href="/" aria-label="EquiSlice home">
            <span className="site-brand-mark" aria-hidden="true"><i /><i /><i /><i /></span>
            <span>Equi<span>Slice</span></span>
          </a>
          <span>Panorama → tile grid</span>
          <a href="/cutter">Start slicing →</a>
        </div>
      </footer>
    </div>
  )
}
