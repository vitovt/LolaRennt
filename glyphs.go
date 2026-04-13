package main

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type glyphMatrix struct {
	Cols int
	Rows int
	Data []bool
}

func (g glyphMatrix) At(x, y int) bool {
	if x < 0 || y < 0 || x >= g.Cols || y >= g.Rows {
		return false
	}
	return g.Data[y*g.Cols+x]
}

var (
	glyphCacheMu sync.Mutex
	glyphCache   = map[string]glyphMatrix{}
)

func matrixGlyphForRune(r rune, cols, rows int) glyphMatrix {
	if r == ' ' || r == '\t' || r == '\n' {
		return glyphMatrix{Cols: cols, Rows: rows, Data: make([]bool, cols*rows)}
	}

	key := fmt.Sprintf("%d:%d:%U", cols, rows, r)
	glyphCacheMu.Lock()
	if cached, ok := glyphCache[key]; ok {
		glyphCacheMu.Unlock()
		return cached
	}
	glyphCacheMu.Unlock()

	face, _, err := loadFace(90)
	if err != nil {
		return glyphMatrix{Cols: cols, Rows: rows, Data: make([]bool, cols*rows)}
	}
	defer closeFace(face)

	canvasW := 128
	canvasH := 128
	mask := image.NewAlpha(image.Rect(0, 0, canvasW, canvasH))

	advance := font.MeasureString(face, string(r)).Ceil()
	metrics := face.Metrics()
	ascent := metrics.Ascent.Ceil()
	descent := metrics.Descent.Ceil()
	x := maxInt((canvasW-advance)/2, 4)
	y := maxInt((canvasH+(ascent-descent))/2, ascent+4)

	d := font.Drawer{
		Dst:  mask,
		Src:  image.NewUniform(color.Alpha{A: 255}),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(string(r))

	bounds, ok := alphaBounds(mask)
	if !ok {
		empty := glyphMatrix{Cols: cols, Rows: rows, Data: make([]bool, cols*rows)}
		glyphCacheMu.Lock()
		glyphCache[key] = empty
		glyphCacheMu.Unlock()
		return empty
	}

	matrix := glyphMatrix{
		Cols: cols,
		Rows: rows,
		Data: make([]bool, cols*rows),
	}

	for gy := 0; gy < rows; gy++ {
		for gx := 0; gx < cols; gx++ {
			cellMinX := bounds.Min.X + (gx*(bounds.Dx()))/cols
			cellMaxX := bounds.Min.X + ((gx+1)*(bounds.Dx()))/cols
			cellMinY := bounds.Min.Y + (gy*(bounds.Dy()))/rows
			cellMaxY := bounds.Min.Y + ((gy+1)*(bounds.Dy()))/rows

			if cellMaxX <= cellMinX {
				cellMaxX = cellMinX + 1
			}
			if cellMaxY <= cellMinY {
				cellMaxY = cellMinY + 1
			}

			var covered int
			var total int
			for py := cellMinY; py < cellMaxY; py++ {
				for px := cellMinX; px < cellMaxX; px++ {
					total++
					if mask.AlphaAt(px, py).A > 40 {
						covered++
					}
				}
			}
			matrix.Data[gy*cols+gx] = total > 0 && float64(covered)/float64(total) > 0.18
		}
	}

	glyphCacheMu.Lock()
	glyphCache[key] = matrix
	glyphCacheMu.Unlock()
	return matrix
}

func alphaBounds(img *image.Alpha) (image.Rectangle, bool) {
	minX := img.Bounds().Max.X
	minY := img.Bounds().Max.Y
	maxX := img.Bounds().Min.X
	maxY := img.Bounds().Min.Y
	found := false

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			if img.AlphaAt(x, y).A == 0 {
				continue
			}
			found = true
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	if !found {
		return image.Rectangle{}, false
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), true
}
