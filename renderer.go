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
	paintBackground(img, project)

	animated := buildAnimatedFrame(project, stats, frame)
	lines := strings.Split(animated.Text, "\n")

	switch project.Display.Mode {
	case displayModeDotMatrix:
		if err := drawDotMatrixText(img, project, lines); err != nil {
			return nil, err
		}
	default:
		if err := drawSegmentText(img, project, lines); err != nil {
			return nil, err
		}
	}

	drawFooterInfo(img, project, animated)
	return img, nil
}

func paintBackground(img *image.NRGBA, project Project) {
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
		draw.Draw(img, bounds, image.NewUniform(color.NRGBA{R: 18, G: 22, B: 28, A: 255}), image.Point{}, draw.Src)
		drawPlaceholderBands(img, color.NRGBA{R: 56, G: 78, B: 86, A: 255}, color.NRGBA{R: 24, G: 32, B: 36, A: 255})
	case backgroundModeVideo:
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

func drawSegmentText(img *image.NRGBA, project Project, lines []string) error {
	face, lineHeight, err := fittedFace(lines, img.Bounds().Dx(), img.Bounds().Dy())
	if err != nil {
		return err
	}
	defer closeFace(face)

	mainColor := parseHexColor(project.Style.MainColor, color.NRGBA{R: 255, G: 96, B: 64, A: 255})
	glowColor := parseHexColor(project.Style.GlowColor, color.NRGBA{R: 255, G: 140, B: 102, A: 255})

	glowAlpha := uint8(40 + 1.8*project.Style.GlowIntensity)
	if glowAlpha > 180 {
		glowAlpha = 180
	}
	glow := color.NRGBA{R: glowColor.R, G: glowColor.G, B: glowColor.B, A: glowAlpha}

	totalHeight := lineHeight * len(lines)
	startY := (img.Bounds().Dy()-totalHeight)/2 + lineHeight
	for lineIndex, line := range lines {
		width := font.MeasureString(face, line).Ceil()
		x := alignedX(project.Layout.Alignment, img.Bounds().Dx(), width, int(project.Layout.Padding))
		y := startY + lineIndex*lineHeight
		for _, offset := range []image.Point{{-2, 0}, {2, 0}, {0, -2}, {0, 2}, {-1, -1}, {1, 1}} {
			drawTextLine(img, face, line, x+offset.X, y+offset.Y, glow)
		}
		drawTextLine(img, face, line, x, y, mainColor)
	}
	return nil
}

func drawDotMatrixText(img *image.NRGBA, project Project, lines []string) error {
	mask := image.NewAlpha(img.Bounds())
	face, lineHeight, err := fittedFace(lines, img.Bounds().Dx(), img.Bounds().Dy())
	if err != nil {
		return err
	}
	defer closeFace(face)

	totalHeight := lineHeight * len(lines)
	startY := (img.Bounds().Dy()-totalHeight)/2 + lineHeight
	for lineIndex, line := range lines {
		width := font.MeasureString(face, line).Ceil()
		x := alignedX(project.Layout.Alignment, img.Bounds().Dx(), width, int(project.Layout.Padding))
		y := startY + lineIndex*lineHeight
		drawTextLineAlpha(mask, face, line, x, y, color.Alpha{A: 255})
	}

	mainColor := parseHexColor(project.Style.MainColor, color.NRGBA{R: 255, G: 96, B: 64, A: 255})
	glowColor := parseHexColor(project.Style.GlowColor, color.NRGBA{R: 255, G: 140, B: 102, A: 255})
	step := maxInt(int(14*project.Layout.CellScale), 8)
	radius := maxInt(step/4, 2)
	for y := mask.Bounds().Min.Y; y < mask.Bounds().Max.Y; y += step {
		for x := mask.Bounds().Min.X; x < mask.Bounds().Max.X; x += step {
			if mask.AlphaAt(x, y).A < 10 {
				continue
			}
			drawGlowDot(img, x, y, radius+1, color.NRGBA{R: glowColor.R, G: glowColor.G, B: glowColor.B, A: 70})
			fillCircle(img, x, y, radius, color.NRGBA{R: mainColor.R, G: mainColor.G, B: mainColor.B, A: 255})
		}
	}
	return nil
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

func fittedFace(lines []string, width, height int) (font.Face, int, error) {
	lineCount := maxInt(len(lines), 1)
	for size := float64(height)/float64(lineCount+2) + 10; size >= 14; size -= 2 {
		face, lineHeight, err := loadFace(size)
		if err != nil {
			return nil, 0, err
		}
		maxWidth := 0
		for _, line := range lines {
			maxWidth = maxInt(maxWidth, font.MeasureString(face, line).Ceil())
		}
		totalHeight := lineHeight * lineCount
		if maxWidth <= width-40 && totalHeight <= height-80 {
			return face, lineHeight, nil
		}
		closeFace(face)
	}
	return loadFace(14)
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

func drawTextLineAlpha(img *image.Alpha, face font.Face, text string, x, y int, col color.Alpha) {
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
