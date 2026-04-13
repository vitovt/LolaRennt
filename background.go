package main

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"os"
	"sync"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

var (
	backgroundCacheMu sync.Mutex
	backgroundCache   = map[string]image.Image{}
)

func loadBackgroundImage(path string) (image.Image, error) {
	backgroundCacheMu.Lock()
	if cached, ok := backgroundCache[path]; ok {
		backgroundCacheMu.Unlock()
		return cached, nil
	}
	backgroundCacheMu.Unlock()

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	backgroundCacheMu.Lock()
	backgroundCache[path] = img
	backgroundCacheMu.Unlock()
	return img, nil
}

func drawImageBackground(dst *image.NRGBA, src image.Image, fitMode string) {
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.NRGBA{R: 7, G: 8, B: 10, A: 255}), image.Point{}, draw.Src)

	srcBounds := src.Bounds()
	if srcBounds.Dx() == 0 || srcBounds.Dy() == 0 {
		return
	}

	dstW := dst.Bounds().Dx()
	dstH := dst.Bounds().Dy()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	scaleX := float64(dstW) / float64(srcW)
	scaleY := float64(dstH) / float64(srcH)

	var drawRect image.Rectangle
	switch fitMode {
	case "Stretch":
		drawRect = image.Rect(0, 0, dstW, dstH)
	case "Fill":
		scale := maxFloat(scaleX, scaleY)
		w := int(float64(srcW) * scale)
		h := int(float64(srcH) * scale)
		drawRect = image.Rect((dstW-w)/2, (dstH-h)/2, (dstW-w)/2+w, (dstH-h)/2+h)
	case "Center":
		w := srcW
		h := srcH
		if w > dstW || h > dstH {
			scale := minFloat(scaleX, scaleY)
			w = int(float64(srcW) * scale)
			h = int(float64(srcH) * scale)
		}
		drawRect = image.Rect((dstW-w)/2, (dstH-h)/2, (dstW-w)/2+w, (dstH-h)/2+h)
	default:
		scale := minFloat(scaleX, scaleY)
		w := int(float64(srcW) * scale)
		h := int(float64(srcH) * scale)
		drawRect = image.Rect((dstW-w)/2, (dstH-h)/2, (dstW-w)/2+w, (dstH-h)/2+h)
	}

	scaled := image.NewNRGBA(drawRect)
	xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), src, srcBounds, draw.Over, nil)
	draw.Draw(dst, drawRect, scaled, scaled.Bounds().Min, draw.Over)
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
