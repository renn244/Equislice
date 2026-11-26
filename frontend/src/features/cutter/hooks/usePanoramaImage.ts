import axios from 'axios'
import { useEffect, useRef, useState, type ChangeEvent } from 'react'
import type { ImageDimensions } from '../types/panorama.type'

type PanoramaImage = {
  selectedFile: File | null
  dimensions: ImageDimensions | null
  previewUrl: string | null
  fileError: string | null
}

const emptyImage: PanoramaImage = {
  selectedFile: null,
  dimensions: null,
  previewUrl: null,
  fileError: null,
}

export function usePanoramaImage() {
  const [state, setState] = useState(emptyImage)
  const selection = useRef(0)
  const pendingUrl = useRef<string | null>(null)

  useEffect(() => () => {
    if (state.previewUrl) URL.revokeObjectURL(state.previewUrl)
  }, [state.previewUrl])

  useEffect(() => () => {
    selection.current++
    if (pendingUrl.current) URL.revokeObjectURL(pendingUrl.current)
  }, [])

  function reset() {
    selection.current++
    if (pendingUrl.current) URL.revokeObjectURL(pendingUrl.current)
    pendingUrl.current = null
    setState(emptyImage)
  }

  function prepareFile(file: File) {
    const version = ++selection.current
    if (pendingUrl.current) URL.revokeObjectURL(pendingUrl.current)
    pendingUrl.current = null

    if (file.type !== 'image/jpeg' && file.type !== 'image/png') {
      setState({ ...emptyImage, fileError: 'Choose a JPG or PNG image.' })
      return
    }
    if (file.size > 60 * 1024 * 1024) {
      setState({ ...emptyImage, fileError: 'Choose an image that is 60 MB or smaller.' })
      return
    }

    const objectUrl = URL.createObjectURL(file)
    pendingUrl.current = objectUrl
    const image = new Image()
    image.onload = () => {
      if (selection.current !== version) return
      pendingUrl.current = null
      setState({
        dimensions: { height: image.naturalHeight, width: image.naturalWidth },
        fileError: null,
        previewUrl: objectUrl,
        selectedFile: file,
      })
    }
    image.onerror = () => {
      if (selection.current !== version) return
      pendingUrl.current = null
      URL.revokeObjectURL(objectUrl)
      setState({ ...emptyImage, fileError: 'We could not read that image file.' })
    }
    image.src = objectUrl
  }

  function onFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.currentTarget.files?.[0]
    if (file) prepareFile(file)
    event.currentTarget.value = ''
  }

  async function loadSample(sample: { path: string; fileName: string }) {
    const version = ++selection.current
    try {
      const { data: blob } = await axios.get<Blob>(sample.path, { responseType: 'blob' })
      if (selection.current !== version) return
      prepareFile(new File([blob], sample.fileName, { type: 'image/jpeg' }))
    } catch {
      if (selection.current !== version) return
      setState(current => ({ ...current, fileError: 'We could not load that sample panorama.' }))
    }
  }

  return { ...state, onFileChange, loadSample, reset }
}
