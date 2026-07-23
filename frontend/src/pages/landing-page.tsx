import {
  ArrowRightIcon,
  DownloadIcon,
  FileIcon,
  FileZipIcon,
  RowsIcon,
  UploadIcon,
} from '@primer/octicons-react'
import { Heading, LinkButton, Text } from '@primer/react'
import './landing-page.css'

export function LandingPage() {
  return (
    <div className="landing-page">
      <main>
        <section className="landing-main landing-hero" aria-labelledby="landing-title">
          <div className="landing-copy">
            <Heading
              as="h1"
              id="landing-title"
              className="landing-title"
              aria-label="Turn one panorama into a clean tile set."
            >
              <span>Turn one panorama </span>
              <span>into a clean tile set.</span>
            </Heading>

            <Text as="p" size="large" className="landing-description">
              Upload an equirectangular image, choose your grid, and download evenly organized tiles
              for your virtual-tour workflow.
            </Text>

            <LinkButton
              className="landing-cta"
              href="/cutter"
              size="large"
              variant="primary"
              trailingAction={ArrowRightIcon}
            >
              Open cutter
            </LinkButton>

            <Text as="p" size="small" className="landing-privacy">
              Uploads and exports are temporary and automatically deleted after 24 hours.
            </Text>
          </div>

          <div className="landing-visual">
            <img
              src="/panorama-tiles-transparent.png"
              alt="A panorama arranged as an evenly sliced tile set"
              width="1693"
              height="961"
            />
          </div>
        </section>

        <section className="workflow-section" id="how-it-works" aria-labelledby="workflow-title">
          <div className="section-shell">
            <Heading as="h2" id="workflow-title" className="section-title">
              From panorama to tiles in three steps
            </Heading>
            <div className="workflow-steps">
              <article>
                <span className="step-number">1</span>
                <UploadIcon className="step-icon" size={32} aria-hidden="true" />
                <Heading as="h3">Upload your panorama</Heading>
                <Text as="p">Send an equirectangular image when you are ready to slice it.</Text>
              </article>
              <article>
                <span className="step-number">2</span>
                <RowsIcon className="step-icon" size={32} aria-hidden="true" />
                <Heading as="h3">Choose your grid</Heading>
                <Text as="p">Set the rows, columns, and naming pattern for the final tile set.</Text>
              </article>
              <article>
                <span className="step-number">3</span>
                <DownloadIcon className="step-icon" size={32} aria-hidden="true" />
                <Heading as="h3">Export your tiles</Heading>
                <Text as="p">Follow the job, then download one organized ZIP when processing finishes.</Text>
              </article>
            </div>
            <div className="workflow-strip" aria-hidden="true">
              <div />
              <div />
              <div />
              <div />
              <div />
              <div />
              <div />
              <div />
              <div />
              <div />
              <div />
              <div />
            </div>
          </div>
        </section>

        <section className="features-section" id="features" aria-labelledby="features-title">
          <div className="section-shell features-heading">
            <Heading as="h2" id="features-title" className="section-title">
              Control the tile set before you export.
            </Heading>
            <Text as="p" size="large">
              Keep grid size, tile count, and file naming visible in one focused workflow.
            </Text>
          </div>
          <div className="section-shell feature-rows">
            <article className="feature-row">
              <div className="feature-copy">
                <span>01</span>
                <Heading as="h3">Choose the exact grid</Heading>
                <Text as="p">Match rows and columns to the viewer or delivery format you need.</Text>
              </div>
              <div className="feature-preview grid-preview" aria-label="Example tile grid settings">
                <div className="preview-heading">Tile grid</div>
                <div className="grid-controls">
                  <label>Columns <strong>6</strong></label>
                  <label>Rows <strong>3</strong></label>
                </div>
                <div className="tile-grid" aria-hidden="true">
                  {Array.from({ length: 18 }, (_, index) => <i key={index} />)}
                </div>
              </div>
            </article>
            <article className="feature-row feature-row-reverse">
              <div className="feature-copy">
                <span>02</span>
                <Heading as="h3">Keep names predictable</Heading>
                <Text as="p">Use a consistent row-and-column pattern so every tile is easy to place.</Text>
              </div>
              <div className="feature-preview manifest-preview" aria-label="Example export file names">
                <div className="preview-heading">Export files</div>
                {['tile_r01_c01.jpg', 'tile_r01_c02.jpg', 'tile_r01_c03.jpg'].map((name) => (
                  <div className="manifest-row" key={name}>
                    <FileIcon size={16} aria-hidden="true" />
                    <code>{name}</code>
                  </div>
                ))}
              </div>
            </article>
            <article className="feature-row">
              <div className="feature-copy">
                <span>03</span>
                <Heading as="h3">See the tile count</Heading>
                <Text as="p">Know the exact number of output tiles as soon as the grid is valid.</Text>
              </div>
              <div className="feature-preview count-preview" aria-label="Example tile count">
                <RowsIcon size={28} aria-hidden="true" />
                <div>
                  <strong>18 tiles</strong>
                  <span>6 columns x 3 rows</span>
                </div>
              </div>
            </article>
          </div>
        </section>

        <section className="closing-section" id="privacy" aria-labelledby="closing-title">
          <div className="section-shell closing-shell">
            <div>
              <Heading as="h2" id="closing-title" className="section-title">
                One panorama in. A clean tile set out.
              </Heading>
              <Text as="p" size="large">
                Choose the grid once and leave the repetitive slicing work to EquiSlice.
              </Text>
              <Text as="p" className="closing-note">
                Uploads and exports are automatically deleted after 24 hours.
              </Text>
              <LinkButton href="/cutter" size="large" variant="primary" trailingAction={ArrowRightIcon}>
                Open cutter
              </LinkButton>
            </div>
            <div className="closing-flow" aria-hidden="true">
              <FileIcon size={50} />
              <ArrowRightIcon size={24} />
              <RowsIcon size={54} />
              <ArrowRightIcon size={24} />
              <FileZipIcon size={50} />
            </div>
          </div>
        </section>
      </main>

      <footer className="landing-footer">
        <div className="landing-footer-inner">
          <Text as="span" weight="semibold" className="landing-brand">
            EquiSlice
          </Text>
          <nav aria-label="Footer navigation">
            <a href="#how-it-works">How it works</a>
            <a href="#features">Features</a>
            <a href="#privacy">Privacy</a>
            <a href="/cutter">Open cutter</a>
          </nav>
        </div>
      </footer>
    </div>
  )
}
