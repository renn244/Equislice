type SiteHeaderProps = {
  compact?: boolean
}

export function SiteHeader({ compact = false }: SiteHeaderProps) {
  return (
    <header className={`site-header${compact ? ' site-header-compact' : ''}`}>
      <a className="site-brand" href="/" aria-label="EquiSlice home">
        <span className="site-brand-mark" aria-hidden="true">
          <i />
          <i />
          <i />
          <i />
        </span>
        <span>Equi<span>Slice</span></span>
      </a>

      {compact && <span className="site-workspace-label">Cutter workspace</span>}

      <nav className="site-nav" aria-label="Primary navigation">
        {compact ? (
          <a className="site-nav-back" href="/">Back to overview</a>
        ) : (
          <>
            <a href="#how">How it works</a>
            <a href="#why">Why EquiSlice</a>
            <a className="site-nav-action" href="/cutter">Slice a panorama</a>
          </>
        )}
      </nav>
    </header>
  )
}
