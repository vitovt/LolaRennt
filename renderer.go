package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	fontOnce sync.Once
	fontTTF  *opentype.Font
	fontErr  error
)

func renderImage(project Project, stats textStats, frame int, width, height int) (image.Image, error) {
	if width <= 0 {
		width = 1280
	}
	if height <= 0 {
		height = 720
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	animated := buildAnimatedFrame(project, stats, frame)
	paintBackground(img, project, animated.Time)
	lines := strings.Split(animated.Text, "\n")

	switch project.Display.Mode {
	case displayModeDotMatrix:
		if err := drawMatrixDisplayText(img, project, lines, displayModeDotMatrix); err != nil {
			return nil, err
		}
	case displayModeSegment:
		if err := drawSegmentDisplayText(img, project, lines); err != nil {
			return nil, err
		}
	default:
		if err := drawMatrixDisplayText(img, project, lines, displayModeBlockMatrix); err != nil {
			return nil, err
		}
	}
	if project.Style.FlickerAmount > 0 {
		applyFlickerEffect(img, project.Style.FlickerAmount, animated.Frame, project.Animation.Seed)
	}
	if project.Style.NoiseAmount > 0 {
		applyNoiseOverlay(img, project.Style.NoiseAmount, animated.Frame, project.Animation.Seed)
	}
	if project.Style.Scanlines {
		applyScanlineOverlay(img)
	}

	drawFooterInfo(img, project, animated)
	return img, nil
}

func paintBackground(img *image.NRGBA, project Project, timeSec float64) {
	bounds := img.Bounds()
	switch project.Background.Mode {
	case backgroundModeTransparent:
		return
	case backgroundModeSolid:
		draw.Draw(img, bounds, image.NewUniform(parseHexColor(project.Background.SolidColor, color.NRGBA{R: 5, G: 6, B: 8, A: 255})), image.Point{}, draw.Src)
	case backgroundModeGradient:
		top := parseHexColor(project.Background.GradientA, color.NRGBA{R: 6, G: 10, B: 13, A: 255})
		bottom := parseHexColor(project.Background.GradientB, color.NRGBA{R: 18, G: 38, B: 43, A: 255})
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			t := float64(y-bounds.Min.Y) / float64(maxInt(bounds.Dy()-1, 1))
			row := color.NRGBA{
				R: uint8(float64(top.R)*(1-t) + float64(bottom.R)*t),
				G: uint8(float64(top.G)*(1-t) + float64(bottom.G)*t),
				B: uint8(float64(top.B)*(1-t) + float64(bottom.B)*t),
				A: 255,
			}
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				img.SetNRGBA(x, y, row)
			}
		}
	case backgroundModeImage:
		if project.Background.ImagePath != "" {
			bg, err := loadBackgroundImage(project.Background.ImagePath)
			if err == nil {
				drawImageBackground(img, bg, project.Background.FitMode, project.Background.ImageOpacity)
				return
			}
		}
		draw.Draw(img, bounds, image.NewUniform(color.NRGBA{R: 18, G: 22, B: 28, A: 255}), image.Point{}, draw.Src)
		drawPlaceholderBands(img, color.NRGBA{R: 56, G: 78, B: 86, A: 255}, color.NRGBA{R: 24, G: 32, B: 36, A: 255})
	case backgroundModeVideo:
		if project.Background.VideoPath != "" {
			tools := resolveFFmpegTools(project)
			bg, err := loadVideoBackgroundFrame(tools.FFmpegPath, project.Background.VideoPath, timeSec)
			if err == nil {
				drawImageBackground(img, bg, project.Background.FitMode, 100)
				return
			}
		}
		draw.Draw(img, bounds, image.NewUniform(color.NRGBA{R: 16, G: 12, B: 20, A: 255}), image.Point{}, draw.Src)
		drawPlaceholderBands(img, color.NRGBA{R: 83, G: 52, B: 89, A: 255}, color.NRGBA{R: 28, G: 20, B: 32, A: 255})
	default:
		draw.Draw(img, bounds, image.NewUniform(color.NRGBA{R: 5, G: 6, B: 8, A: 255}), image.Point{}, draw.Src)
	}
}

func drawPlaceholderBands(img *image.NRGBA, a, b color.NRGBA) {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if (x+y/2)%180 < 90 {
				img.SetNRGBA(x, y, a)
			} else {
				img.SetNRGBA(x, y, b)
			}
		}
	}
}

func drawMatrixDisplayText(img *image.NRGBA, project Project, lines []string, mode string) error {
	layout := computeDisplayLayout(project, lines, img.Bounds().Dx(), img.Bounds().Dy(), mode)
	mainColor := parseHexColor(project.Style.MainColor, color.NRGBA{R: 255, G: 96, B: 64, A: 255})
	glowColor := parseHexColor(project.Style.GlowColor, color.NRGBA{R: 255, G: 140, B: 102, A: 255})
	inactiveColor := parseHexColor(project.Style.InactiveColor, color.NRGBA{R: 49, G: 20, B: 14, A: 255})
	inactiveColor.A = uint8(40 + (project.Style.InactiveVisibility/100.0)*140)
	blockMode := mode == displayModeBlockMatrix

	totalHeight := len(lines)*layout.charHeight + maxInt(len(lines)-1, 0)*layout.lineGap
	startY := maxInt((img.Bounds().Dy()-totalHeight)/2, layout.padding)
	for lineIndex, line := range lines {
		lineRunes := []rune(line)
		lineWidth := len(lineRunes)*layout.charWidth + maxInt(len(lineRunes)-1, 0)*layout.charGap
		x := alignedX(project.Layout.Alignment, img.Bounds().Dx(), lineWidth, layout.padding)
		y := startY + lineIndex*(layout.charHeight+layout.lineGap)

		for _, r := range lineRunes {
			matrix := matrixGlyphForRune(r, layout.cols, layout.rows)
			drawGlyphBox(img, x, y, layout, matrix, inactiveColor, mainColor, glowColor, blockMode)
			x += layout.charWidth + layout.charGap
		}
	}
	return nil
}

func drawSegmentDisplayText(img *image.NRGBA, project Project, lines []string) error {
	layout := computeDisplayLayout(project, lines, img.Bounds().Dx(), img.Bounds().Dy(), displayModeSegment)
	mainColor := parseHexColor(project.Style.MainColor, color.NRGBA{R: 255, G: 96, B: 64, A: 255})
	glowColor := parseHexColor(project.Style.GlowColor, color.NRGBA{R: 255, G: 140, B: 102, A: 255})

	totalHeight := len(lines)*layout.charHeight + maxInt(len(lines)-1, 0)*layout.lineGap
	startY := maxInt((img.Bounds().Dy()-totalHeight)/2, layout.padding)
	for lineIndex, line := range lines {
		lineRunes := []rune(line)
		lineWidth := len(lineRunes)*layout.charWidth + maxInt(len(lineRunes)-1, 0)*layout.charGap
		x := alignedX(project.Layout.Alignment, img.Bounds().Dx(), lineWidth, layout.padding)
		y := startY + lineIndex*(layout.charHeight+layout.lineGap)

		for _, r := range lineRunes {
			matrix := matrixGlyphForRune(r, layout.cols, layout.rows)
			drawSegmentGlyph(img, x, y, layout, matrix, mainColor, glowColor)
			x += layout.charWidth + layout.charGap
		}
	}
	return nil
}

func applyScanlineOverlay(img *image.NRGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		if y%4 == 0 {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				darkenNRGBAPixel(img, x, y, 0.74)
			}
			continue
		}
		if y%4 == 2 {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				darkenNRGBAPixel(img, x, y, 0.9)
			}
		}
	}
}

func applyFlickerEffect(img *image.NRGBA, amount float64, frame int, seed string) {
	seedBase := hashSeed(seed)
	wave := math.Sin(float64(frame)*0.55 + float64(seedBase%360)*math.Pi/180)
	jitter := effectNoiseValue(seedBase, frame, 0, 0)
	intensity := 1 + ((wave*0.65 + jitter*0.35) * (amount / 100.0) * 0.16)
	intensity = minFloat(1.2, math.Max(0.82, intensity))

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			scaleNRGBAPixel(img, x, y, intensity)
		}
	}
}

func applyNoiseOverlay(img *image.NRGBA, amount float64, frame int, seed string) {
	seedBase := hashSeed(seed)
	strength := maxInt(int((amount/100.0)*24), 1)
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 2 {
			delta := int(effectNoiseValue(seedBase, frame, x, y) * float64(strength))
			for yy := y; yy < minInt(y+2, bounds.Max.Y); yy++ {
				for xx := x; xx < minInt(x+2, bounds.Max.X); xx++ {
					adjustNRGBAPixel(img, xx, yy, delta)
				}
			}
		}
	}
}

func effectNoiseValue(seedBase uint64, frame, x, y int) float64 {
	value := uint64(x+1)*0x9E3779B185EBCA87 ^ uint64(y+1)*0xC2B2AE3D27D4EB4F
	value ^= uint64(frame+1) * 0x165667B19E3779F9
	value ^= seedBase
	value ^= value >> 30
	value *= 0xBF58476D1CE4E5B9
	value ^= value >> 27
	value *= 0x94D049BB133111EB
	value ^= value >> 31
	return float64(value&0xff)/255.0*2 - 1
}

func darkenNRGBAPixel(img *image.NRGBA, x, y int, amount float64) {
	scaleNRGBAPixel(img, x, y, amount)
}

func scaleNRGBAPixel(img *image.NRGBA, x, y int, amount float64) {
	offset := img.PixOffset(x, y)
	img.Pix[offset+0] = uint8(float64(img.Pix[offset+0]) * amount)
	img.Pix[offset+1] = uint8(float64(img.Pix[offset+1]) * amount)
	img.Pix[offset+2] = uint8(float64(img.Pix[offset+2]) * amount)
}

func adjustNRGBAPixel(img *image.NRGBA, x, y, delta int) {
	offset := img.PixOffset(x, y)
	img.Pix[offset+0] = clampByte(img.Pix[offset+0], delta)
	img.Pix[offset+1] = clampByte(img.Pix[offset+1], delta)
	img.Pix[offset+2] = clampByte(img.Pix[offset+2], delta)
}

func clampByte(value uint8, delta int) uint8 {
	result := int(value) + delta
	if result < 0 {
		return 0
	}
	if result > 255 {
		return 255
	}
	return uint8(result)
}

type displayLayout struct {
	cols       int
	rows       int
	cellWidth  int
	cellHeight int
	cellGapX   int
	cellGapY   int
	charWidth  int
	charHeight int
	charGap    int
	lineGap    int
	padding    int
}

func computeDisplayLayout(project Project, lines []string, width, height int, mode string) displayLayout {
	cols, rows := 6, 9
	baseCellW, baseCellH := 12.0, 12.0
	baseGapX, baseGapY := 3.0, 3.0
	switch mode {
	case displayModeBlockMatrix:
		cols, rows = 7, 11
		baseCellW, baseCellH = 16.0, 10.0
		baseGapX, baseGapY = 2.0, 2.0
	case displayModeSegment:
		cols, rows = 9, 13
		baseCellW, baseCellH = 10.0, 10.0
		baseGapX, baseGapY = 2.0, 2.0
	}

	scale := project.Layout.CellScale
	if scale <= 0 {
		scale = 1
	}

	cellW := maxInt(int(baseCellW*scale), 4)
	cellH := maxInt(int(baseCellH*scale), 4)
	gapX := maxInt(int(baseGapX*scale), 1)
	gapY := maxInt(int(baseGapY*scale), 1)
	charGap := maxInt(int(project.Layout.CharacterSpacing*scale/3), 4)
	lineGap := maxInt(int(project.Layout.LineSpacing*scale/2), 8)
	padding := maxInt(int(project.Layout.Padding), 18)

	charWidth := cols*cellW + maxInt(cols-1, 0)*gapX
	charHeight := rows*cellH + maxInt(rows-1, 0)*gapY

	maxLineLen := 1
	for _, line := range lines {
		maxLineLen = maxInt(maxLineLen, len([]rune(line)))
	}
	totalWidth := maxLineLen*charWidth + maxInt(maxLineLen-1, 0)*charGap + 2*padding
	totalHeight := len(lines)*charHeight + maxInt(len(lines)-1, 0)*lineGap + 2*padding
	fitFactor := minFloat(float64(width)/float64(maxInt(totalWidth, 1)), float64(height)/float64(maxInt(totalHeight, 1)))
	if fitFactor < 1 {
		cellW = maxInt(int(float64(cellW)*fitFactor), 2)
		cellH = maxInt(int(float64(cellH)*fitFactor), 2)
		gapX = maxInt(int(float64(gapX)*fitFactor), 1)
		gapY = maxInt(int(float64(gapY)*fitFactor), 1)
		charGap = maxInt(int(float64(charGap)*fitFactor), 2)
		lineGap = maxInt(int(float64(lineGap)*fitFactor), 4)
		padding = maxInt(int(float64(padding)*fitFactor), 12)
		charWidth = cols*cellW + maxInt(cols-1, 0)*gapX
		charHeight = rows*cellH + maxInt(rows-1, 0)*gapY
	}

	return displayLayout{
		cols:       cols,
		rows:       rows,
		cellWidth:  cellW,
		cellHeight: cellH,
		cellGapX:   gapX,
		cellGapY:   gapY,
		charWidth:  charWidth,
		charHeight: charHeight,
		charGap:    charGap,
		lineGap:    lineGap,
		padding:    padding,
	}
}

type segmentBar struct {
	startX int
	startY int
	endX   int
	endY   int
}

type segmentDot struct {
	minX int
	minY int
	maxX int
	maxY int
}

func drawSegmentGlyph(img *image.NRGBA, x, y int, layout displayLayout, matrix glyphMatrix, activeColor, glowColor color.NRGBA) {
	bars, dots := extractSegmentPrimitives(matrix)
	thickness := maxInt(minInt(layout.cellWidth, layout.cellHeight)*5/8, 2)
	glowAlpha := uint8(52)

	for _, bar := range bars {
		rect := segmentBarRect(x, y, layout, bar, thickness)
		drawBarGlow(img, rect, glowColor, glowAlpha)
		fillRoundedRect(img, rect, maxInt(thickness/2, 1), activeColor)
	}

	for _, dot := range dots {
		cx, cy, radius := segmentDotGeometry(x, y, layout, dot, thickness)
		drawGlowDot(img, cx, cy, radius+2, color.NRGBA{R: glowColor.R, G: glowColor.G, B: glowColor.B, A: glowAlpha})
		fillCircle(img, cx, cy, radius, activeColor)
	}
}

func extractSegmentPrimitives(matrix glyphMatrix) ([]segmentBar, []segmentDot) {
	covered := make([]bool, len(matrix.Data))
	bars := make([]segmentBar, 0, matrix.Rows+matrix.Cols)

	for y := 0; y < matrix.Rows; y++ {
		for x := 0; x < matrix.Cols; {
			if !matrix.At(x, y) {
				x++
				continue
			}
			start := x
			for x < matrix.Cols && matrix.At(x, y) {
				x++
			}
			end := x - 1
			if shouldUseHorizontalBar(matrix, start, end, y) {
				bars = append(bars, segmentBar{startX: start, startY: y, endX: end, endY: y})
				markCoveredCells(covered, matrix.Cols, start, y, end, y)
			}
		}
	}

	for x := 0; x < matrix.Cols; x++ {
		for y := 0; y < matrix.Rows; {
			if !matrix.At(x, y) {
				y++
				continue
			}
			start := y
			for y < matrix.Rows && matrix.At(x, y) {
				y++
			}
			end := y - 1
			if shouldUseVerticalBar(matrix, x, start, end) {
				bars = append(bars, segmentBar{startX: x, startY: start, endX: x, endY: end})
				markCoveredCells(covered, matrix.Cols, x, start, x, end)
			}
		}
	}

	dots := extractSegmentDots(matrix, covered)
	return bars, dots
}

func shouldUseHorizontalBar(matrix glyphMatrix, startX, endX, y int) bool {
	runLength := endX - startX + 1
	if runLength >= 2 {
		return true
	}
	if runLength <= 1 {
		return false
	}
	for x := startX; x <= endX; x++ {
		if matrix.At(x, y-1) || matrix.At(x, y+1) {
			return true
		}
	}
	return false
}

func shouldUseVerticalBar(matrix glyphMatrix, x, startY, endY int) bool {
	runLength := endY - startY + 1
	if runLength >= 2 {
		return true
	}
	if runLength <= 1 {
		return false
	}
	for y := startY; y <= endY; y++ {
		if matrix.At(x-1, y) || matrix.At(x+1, y) {
			return true
		}
	}
	return false
}

func markCoveredCells(covered []bool, cols, startX, startY, endX, endY int) {
	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			covered[y*cols+x] = true
		}
	}
}

func extractSegmentDots(matrix glyphMatrix, covered []bool) []segmentDot {
	visited := make([]bool, len(matrix.Data))
	dots := make([]segmentDot, 0, matrix.Cols)
	neighbors := [][2]int{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	}

	for y := 0; y < matrix.Rows; y++ {
		for x := 0; x < matrix.Cols; x++ {
			index := y*matrix.Cols + x
			if !matrix.At(x, y) || covered[index] || visited[index] {
				continue
			}

			queue := [][2]int{{x, y}}
			visited[index] = true
			dot := segmentDot{minX: x, minY: y, maxX: x, maxY: y}

			for len(queue) > 0 {
				cell := queue[0]
				queue = queue[1:]
				cx, cy := cell[0], cell[1]
				if cx < dot.minX {
					dot.minX = cx
				}
				if cy < dot.minY {
					dot.minY = cy
				}
				if cx > dot.maxX {
					dot.maxX = cx
				}
				if cy > dot.maxY {
					dot.maxY = cy
				}

				for _, step := range neighbors {
					nx := cx + step[0]
					ny := cy + step[1]
					if !matrix.At(nx, ny) {
						continue
					}
					nIndex := ny*matrix.Cols + nx
					if covered[nIndex] || visited[nIndex] {
						continue
					}
					visited[nIndex] = true
					queue = append(queue, [2]int{nx, ny})
				}
			}

			dots = append(dots, dot)
		}
	}

	return dots
}

func segmentBarRect(originX, originY int, layout displayLayout, bar segmentBar, thickness int) image.Rectangle {
	startCX, startCY := segmentCellCenter(originX, originY, layout, bar.startX, bar.startY)
	endCX, endCY := segmentCellCenter(originX, originY, layout, bar.endX, bar.endY)

	if bar.startY == bar.endY {
		return image.Rect(startCX-thickness/2, startCY-thickness/2, endCX+thickness/2+1, startCY+thickness/2+1)
	}
	return image.Rect(startCX-thickness/2, startCY-thickness/2, startCX+thickness/2+1, endCY+thickness/2+1)
}

func segmentDotGeometry(originX, originY int, layout displayLayout, dot segmentDot, thickness int) (int, int, int) {
	minX := originX + dot.minX*(layout.cellWidth+layout.cellGapX)
	minY := originY + dot.minY*(layout.cellHeight+layout.cellGapY)
	maxX := originX + dot.maxX*(layout.cellWidth+layout.cellGapX) + layout.cellWidth
	maxY := originY + dot.maxY*(layout.cellHeight+layout.cellGapY) + layout.cellHeight
	cx := (minX + maxX) / 2
	cy := (minY + maxY) / 2
	radius := maxInt(minInt(maxX-minX, maxY-minY)/2, maxInt(thickness/3, 1))
	return cx, cy, radius
}

func segmentCellCenter(originX, originY int, layout displayLayout, gx, gy int) (int, int) {
	cellX := originX + gx*(layout.cellWidth+layout.cellGapX)
	cellY := originY + gy*(layout.cellHeight+layout.cellGapY)
	return cellX + layout.cellWidth/2, cellY + layout.cellHeight/2
}

func drawGlyphBox(img *image.NRGBA, x, y int, layout displayLayout, matrix glyphMatrix, inactiveColor, activeColor, glowColor color.NRGBA, blockMode bool) {
	glowAlpha := uint8(32)
	if activeColor.A > 0 {
		glowAlpha = 40
	}

	for gy := 0; gy < layout.rows; gy++ {
		for gx := 0; gx < layout.cols; gx++ {
			cx := x + gx*(layout.cellWidth+layout.cellGapX)
			cy := y + gy*(layout.cellHeight+layout.cellGapY)

			if blockMode {
				drawBlockCell(img, cx, cy, layout.cellWidth, layout.cellHeight, inactiveColor)
				if matrix.At(gx, gy) {
					drawBlockGlow(img, cx, cy, layout.cellWidth, layout.cellHeight, color.NRGBA{R: glowColor.R, G: glowColor.G, B: glowColor.B, A: glowAlpha})
					drawBlockCell(img, cx, cy, layout.cellWidth, layout.cellHeight, activeColor)
				}
			} else {
				radius := maxInt(minInt(layout.cellWidth, layout.cellHeight)/2-1, 1)
				drawGlowDot(img, cx+layout.cellWidth/2, cy+layout.cellHeight/2, radius+2, color.NRGBA{R: glowColor.R, G: glowColor.G, B: glowColor.B, A: 26})
				fillCircle(img, cx+layout.cellWidth/2, cy+layout.cellHeight/2, radius, inactiveColor)
				if matrix.At(gx, gy) {
					drawGlowDot(img, cx+layout.cellWidth/2, cy+layout.cellHeight/2, radius+3, color.NRGBA{R: glowColor.R, G: glowColor.G, B: glowColor.B, A: glowAlpha + 40})
					fillCircle(img, cx+layout.cellWidth/2, cy+layout.cellHeight/2, radius, activeColor)
				}
			}
		}
	}
}

func drawFooterInfo(img *image.NRGBA, project Project, frame animatedFrame) {
	info := project.Background.Mode
	if project.Background.Mode == backgroundModeImage && project.Background.ImagePath != "" {
		info = "Image: " + filepathBase(project.Background.ImagePath)
	}
	if project.Background.Mode == backgroundModeVideo && project.Background.VideoPath != "" {
		info = "Video: " + filepathBase(project.Background.VideoPath)
	}

	lines := []string{
		info,
		"Frame " + strconv.Itoa(frame.Frame) + " • " + project.Animation.Seed,
	}
	face, _, err := loadFace(14)
	if err != nil {
		return
	}
	defer closeFace(face)

	y := img.Bounds().Dy() - 32
	for _, line := range lines {
		drawTextLine(img, face, line, 18, y, color.NRGBA{R: 220, G: 225, B: 232, A: 180})
		y += 16
	}
}

func loadFace(size float64) (font.Face, int, error) {
	fontOnce.Do(func() {
		fontTTF, fontErr = opentype.Parse(gomono.TTF)
	})
	if fontErr != nil {
		return nil, 0, fontErr
	}

	face, err := opentype.NewFace(fontTTF, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, 0, err
	}
	metrics := face.Metrics()
	lineHeight := (metrics.Ascent + metrics.Descent).Ceil() + 6
	return face, lineHeight, nil
}

func closeFace(face font.Face) {
	if c, ok := face.(interface{ Close() error }); ok {
		_ = c.Close()
	}
}

func drawTextLine(img draw.Image, face font.Face, text string, x, y int, col color.Color) {
	d := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}

func alignedX(alignment string, width, lineWidth, padding int) int {
	switch alignment {
	case alignmentLeft:
		return padding
	case alignmentRight:
		return maxInt(width-lineWidth-padding, padding)
	default:
		return maxInt((width-lineWidth)/2, padding)
	}
}

func drawGlowDot(img *image.NRGBA, cx, cy, radius int, col color.NRGBA) {
	fillCircle(img, cx, cy, radius+2, color.NRGBA{R: col.R, G: col.G, B: col.B, A: col.A / 2})
}

func drawBlockGlow(img *image.NRGBA, x, y, width, height int, col color.NRGBA) {
	glow := color.NRGBA{R: col.R, G: col.G, B: col.B, A: 40}
	drawBlockCell(img, x-1, y-1, width+2, height+2, glow)
}

func drawBarGlow(img *image.NRGBA, rect image.Rectangle, col color.NRGBA, alpha uint8) {
	glow := color.NRGBA{R: col.R, G: col.G, B: col.B, A: alpha / 2}
	expanded := image.Rect(rect.Min.X-2, rect.Min.Y-2, rect.Max.X+2, rect.Max.Y+2)
	fillRoundedRect(img, expanded, maxInt(minInt(expanded.Dx(), expanded.Dy())/2, 1), glow)
}

func drawBlockCell(img *image.NRGBA, x, y, width, height int, col color.NRGBA) {
	radius := maxInt(minInt(width, height)/3, 1)
	fillRoundedRect(img, image.Rect(x, y, x+width, y+height), radius, col)
}

func fillCircle(img *image.NRGBA, cx, cy, radius int, col color.NRGBA) {
	if radius <= 0 {
		return
	}
	minX := maxInt(cx-radius, 0)
	maxX := minInt(cx+radius, img.Bounds().Dx()-1)
	minY := maxInt(cy-radius, 0)
	maxY := minInt(cy+radius, img.Bounds().Dy()-1)
	rsq := float64(radius * radius)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			if dx*dx+dy*dy <= rsq {
				blendNRGBA(img, x, y, col)
			}
		}
	}
}

func fillRoundedRect(img *image.NRGBA, rect image.Rectangle, radius int, col color.NRGBA) {
	if rect.Empty() {
		return
	}
	if radius <= 0 {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				blendNRGBA(img, x, y, col)
			}
		}
		return
	}

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if pointInRoundedRect(x, y, rect, radius) {
				blendNRGBA(img, x, y, col)
			}
		}
	}
}

func pointInRoundedRect(x, y int, rect image.Rectangle, radius int) bool {
	if x >= rect.Min.X+radius && x < rect.Max.X-radius {
		return true
	}
	if y >= rect.Min.Y+radius && y < rect.Max.Y-radius {
		return true
	}

	corners := []image.Point{
		{X: rect.Min.X + radius, Y: rect.Min.Y + radius},
		{X: rect.Max.X - radius - 1, Y: rect.Min.Y + radius},
		{X: rect.Min.X + radius, Y: rect.Max.Y - radius - 1},
		{X: rect.Max.X - radius - 1, Y: rect.Max.Y - radius - 1},
	}
	rsq := float64(radius * radius)
	for _, corner := range corners {
		dx := float64(x - corner.X)
		dy := float64(y - corner.Y)
		if dx*dx+dy*dy <= rsq {
			return true
		}
	}
	return false
}

func blendNRGBA(img *image.NRGBA, x, y int, src color.NRGBA) {
	dst := img.NRGBAAt(x, y)
	alpha := float64(src.A) / 255.0
	inv := 1 - alpha
	img.SetNRGBA(x, y, color.NRGBA{
		R: uint8(math.Round(float64(src.R)*alpha + float64(dst.R)*inv)),
		G: uint8(math.Round(float64(src.G)*alpha + float64(dst.G)*inv)),
		B: uint8(math.Round(float64(src.B)*alpha + float64(dst.B)*inv)),
		A: uint8(math.Round(float64(src.A) + float64(dst.A)*inv)),
	})
}

func filepathBase(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return path
	}
	return parts[len(parts)-1]
}
