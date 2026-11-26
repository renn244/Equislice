import * as Sentry from '@sentry/react'
import { useState } from 'react'
import { downloadPanoramaTile } from '../api/panorama.api'
import type { TileDownload } from '../types/panorama.type'

export function useTileDownloads() {
  const [downloadingTileKey, setDownloadingTileKey] = useState<string | null>(null)
  const [tileDownloadError, setTileDownloadError] = useState<string | null>(null)

  async function download(tile: TileDownload) {
    const tileKey = `${tile.row}-${tile.column}`
    setDownloadingTileKey(tileKey)
    setTileDownloadError(null)
    try {
      const blob = await downloadPanoramaTile(tile.urlSAS)
      const objectUrl = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = objectUrl
      link.download = tile.fileName
      document.body.append(link)
      link.click()
      link.remove()
      window.setTimeout(() => URL.revokeObjectURL(objectUrl), 0)
    } catch (error) {
      Sentry.captureException(error, {
        tags: { feature: 'panorama-tile-download' },
        extra: { fileName: tile.fileName, row: tile.row, column: tile.column },
      })
      setTileDownloadError(`Couldn't download ${tile.fileName}. Please try again.`)
    } finally {
      setDownloadingTileKey(null)
    }
  }

  return { downloadingTileKey, tileDownloadError, download }
}
