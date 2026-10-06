package main

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
)

func cropSlice(x0 int, y0 int, x1 int, y1 int, img image.Image) ([]byte, error) {
	croppedDimension := image.Rect(x0, y0, x1, y1)
	tile := image.NewRGBA(image.Rect(0, 0, croppedDimension.Dx(), croppedDimension.Dy()))

	draw.Draw(tile, tile.Bounds(), img, croppedDimension.Min, draw.Src)

	var buf bytes.Buffer

	err := jpeg.Encode(&buf, tile, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, err
	}
	buffers := buf.Bytes()

	return buffers, nil
}
