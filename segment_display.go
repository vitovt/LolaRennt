package main

import (
	"image"
	"image/color"
	"math"
	"unicode"
)

type segmentID int

const (
	segmentTopLeft segmentID = iota
	segmentTopRight
	segmentUpperRight
	segmentLowerRight
	segmentBottomRight
	segmentBottomLeft
	segmentLowerLeft
	segmentUpperLeft
	segmentMiddleLeft
	segmentMiddleRight
	segmentUpperLeftDiagonal
	segmentUpperRightDiagonal
	segmentCenterUpper
	segmentCenterLower
	segmentLowerLeftDiagonal
	segmentLowerRightDiagonal
)

type segmentStrokeUnit struct {
	x1         float64
	y1         float64
	x2         float64
	y2         float64
	widthScale float64
}

type segmentGlyph struct {
	segments  []segmentID
	strokes   []segmentStrokeUnit
	backplane bool
}

type segmentPoint struct {
	x float64
	y float64
}

var allSegmentIDs = []segmentID{
	segmentTopLeft,
	segmentTopRight,
	segmentUpperRight,
	segmentLowerRight,
	segmentBottomRight,
	segmentBottomLeft,
	segmentLowerLeft,
	segmentUpperLeft,
	segmentMiddleLeft,
	segmentMiddleRight,
	segmentUpperLeftDiagonal,
	segmentUpperRightDiagonal,
	segmentCenterUpper,
	segmentCenterLower,
	segmentLowerLeftDiagonal,
	segmentLowerRightDiagonal,
}

var segmentStrokeByID = map[segmentID]segmentStrokeUnit{
	segmentTopLeft:            {x1: 0.20, y1: 0.08, x2: 0.53, y2: 0.08, widthScale: 1},
	segmentTopRight:           {x1: 0.47, y1: 0.08, x2: 0.80, y2: 0.08, widthScale: 1},
	segmentUpperRight:         {x1: 0.84, y1: 0.14, x2: 0.84, y2: 0.44, widthScale: 1},
	segmentLowerRight:         {x1: 0.84, y1: 0.56, x2: 0.84, y2: 0.86, widthScale: 1},
	segmentBottomRight:        {x1: 0.47, y1: 0.92, x2: 0.80, y2: 0.92, widthScale: 1},
	segmentBottomLeft:         {x1: 0.20, y1: 0.92, x2: 0.53, y2: 0.92, widthScale: 1},
	segmentLowerLeft:          {x1: 0.16, y1: 0.56, x2: 0.16, y2: 0.86, widthScale: 1},
	segmentUpperLeft:          {x1: 0.16, y1: 0.14, x2: 0.16, y2: 0.44, widthScale: 1},
	segmentMiddleLeft:         {x1: 0.20, y1: 0.50, x2: 0.53, y2: 0.50, widthScale: 1},
	segmentMiddleRight:        {x1: 0.47, y1: 0.50, x2: 0.80, y2: 0.50, widthScale: 1},
	segmentUpperLeftDiagonal:  {x1: 0.24, y1: 0.14, x2: 0.47, y2: 0.44, widthScale: 0.95},
	segmentUpperRightDiagonal: {x1: 0.76, y1: 0.14, x2: 0.53, y2: 0.44, widthScale: 0.95},
	segmentCenterUpper:        {x1: 0.50, y1: 0.15, x2: 0.50, y2: 0.52, widthScale: 1},
	segmentCenterLower:        {x1: 0.50, y1: 0.48, x2: 0.50, y2: 0.85, widthScale: 1},
	segmentLowerLeftDiagonal:  {x1: 0.24, y1: 0.86, x2: 0.47, y2: 0.56, widthScale: 0.95},
	segmentLowerRightDiagonal: {x1: 0.76, y1: 0.86, x2: 0.53, y2: 0.56, widthScale: 0.95},
}

var (
	umlautStrokes = []segmentStrokeUnit{
		{x1: 0.32, y1: 0.015, x2: 0.42, y2: 0.015, widthScale: 0.65},
		{x1: 0.58, y1: 0.015, x2: 0.68, y2: 0.015, widthScale: 0.65},
	}
	breveStrokes = []segmentStrokeUnit{
		{x1: 0.38, y1: 0.015, x2: 0.62, y2: 0.015, widthScale: 0.65},
	}
)

var segmentGlyphs = map[rune]segmentGlyph{
	'0': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentLowerRight, segmentBottomRight, segmentBottomLeft, segmentLowerLeft, segmentUpperLeft),
	'1': lineGlyph(segmentUpperRight, segmentLowerRight),
	'2': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentLowerLeft, segmentBottomLeft, segmentBottomRight),
	'3': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'4': lineGlyph(segmentUpperLeft, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight),
	'5': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'6': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'7': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentLowerRight),
	'8': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentLowerRight, segmentBottomRight, segmentBottomLeft, segmentLowerLeft, segmentUpperLeft, segmentMiddleLeft, segmentMiddleRight),
	'9': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentUpperLeft, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),

	'A': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentLowerLeft, segmentLowerRight),
	'B': lineGlyph(segmentTopLeft, segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentMiddleRight, segmentUpperRight, segmentLowerRight, segmentBottomLeft),
	'C': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentBottomLeft, segmentBottomRight),
	'D': lineGlyph(segmentTopLeft, segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentBottomLeft),
	'E': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentMiddleRight, segmentBottomLeft, segmentBottomRight),
	'F': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentMiddleRight),
	'G': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentBottomLeft, segmentBottomRight, segmentLowerRight, segmentMiddleRight),
	'H': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentMiddleLeft, segmentMiddleRight),
	'I': lineGlyph(segmentCenterUpper, segmentCenterLower),
	'J': lineGlyph(segmentUpperRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight, segmentLowerLeft),
	'K': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentUpperRightDiagonal, segmentLowerRightDiagonal),
	'L': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentBottomLeft, segmentBottomRight),
	'M': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentUpperLeftDiagonal, segmentUpperRightDiagonal),
	'N': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentUpperLeftDiagonal, segmentLowerRightDiagonal),
	'O': lineGlyphWithStrokes(
		[]segmentID{segmentUpperRight, segmentLowerRight, segmentLowerLeft, segmentUpperLeft},
		segmentStrokeUnit{x1: 0.16, y1: 0.08, x2: 0.84, y2: 0.08, widthScale: 1},
		segmentStrokeUnit{x1: 0.16, y1: 0.92, x2: 0.84, y2: 0.92, widthScale: 1},
	),
	'P': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight),
	'Q': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentLowerRight, segmentBottomRight, segmentBottomLeft, segmentLowerLeft, segmentUpperLeft, segmentLowerRightDiagonal),
	'R': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentLowerRightDiagonal),
	'S': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'T': lineGlyph(segmentTopLeft, segmentTopRight, segmentCenterUpper, segmentCenterLower),
	'U': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'V': lineGlyphWithStrokes(
		[]segmentID{segmentUpperLeft, segmentUpperRight},
		segmentStrokeUnit{x1: 0.22, y1: 0.56, x2: 0.47, y2: 0.86, widthScale: 0.95},
		segmentStrokeUnit{x1: 0.78, y1: 0.56, x2: 0.53, y2: 0.86, widthScale: 0.95},
	),
	'W': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentLowerLeftDiagonal, segmentLowerRightDiagonal),
	'X': lineGlyph(segmentUpperLeftDiagonal, segmentUpperRightDiagonal, segmentLowerLeftDiagonal, segmentLowerRightDiagonal),
	'Y': lineGlyph(segmentUpperLeftDiagonal, segmentUpperRightDiagonal, segmentCenterLower),
	'Z': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRightDiagonal, segmentLowerLeftDiagonal, segmentBottomLeft, segmentBottomRight),

	'Ä': lineGlyphWithStrokes([]segmentID{segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentLowerLeft, segmentLowerRight}, umlautStrokes...),
	'Ö': lineGlyphWithStrokes([]segmentID{segmentTopLeft, segmentTopRight, segmentUpperRight, segmentLowerRight, segmentBottomRight, segmentBottomLeft, segmentLowerLeft, segmentUpperLeft}, umlautStrokes...),
	'Ü': lineGlyphWithStrokes([]segmentID{segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight}, umlautStrokes...),
	'ẞ': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'ß': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),

	'А': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentLowerLeft, segmentLowerRight),
	'Б': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'В': lineGlyph(segmentTopLeft, segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentMiddleRight, segmentUpperRight, segmentLowerRight, segmentBottomLeft),
	'Г': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft),
	'Ґ': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft),
	'Д': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'Е': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentMiddleRight, segmentBottomLeft, segmentBottomRight),
	'Ё': lineGlyphWithStrokes([]segmentID{segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentMiddleRight, segmentBottomLeft, segmentBottomRight}, umlautStrokes...),
	'Є': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentBottomLeft, segmentBottomRight),
	'Ж': lineGlyph(segmentUpperLeftDiagonal, segmentUpperRightDiagonal, segmentLowerLeftDiagonal, segmentLowerRightDiagonal, segmentMiddleLeft, segmentMiddleRight),
	'З': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'И': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentUpperRightDiagonal, segmentLowerLeftDiagonal),
	'І': lineGlyph(segmentCenterUpper, segmentCenterLower),
	'Ї': lineGlyphWithStrokes([]segmentID{segmentCenterUpper, segmentCenterLower}, umlautStrokes...),
	'Й': lineGlyphWithStrokes([]segmentID{segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentUpperRightDiagonal, segmentLowerLeftDiagonal}, breveStrokes...),
	'К': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentUpperRightDiagonal, segmentLowerRightDiagonal),
	'Л': lineGlyph(segmentLowerLeftDiagonal, segmentUpperRight, segmentLowerRight),
	'М': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentUpperLeftDiagonal, segmentUpperRightDiagonal),
	'Н': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentMiddleLeft, segmentMiddleRight),
	'О': lineGlyphWithStrokes(
		[]segmentID{segmentUpperRight, segmentLowerRight, segmentLowerLeft, segmentUpperLeft},
		segmentStrokeUnit{x1: 0.16, y1: 0.08, x2: 0.84, y2: 0.08, widthScale: 1},
		segmentStrokeUnit{x1: 0.16, y1: 0.92, x2: 0.84, y2: 0.92, widthScale: 1},
	),
	'П': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight),
	'Р': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight),
	'С': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperLeft, segmentLowerLeft, segmentBottomLeft, segmentBottomRight),
	'Т': lineGlyph(segmentTopLeft, segmentTopRight, segmentCenterUpper, segmentCenterLower),
	'У': lineGlyph(segmentUpperLeftDiagonal, segmentUpperRightDiagonal, segmentCenterLower),
	'Ф': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentLowerRight, segmentBottomRight, segmentBottomLeft, segmentLowerLeft, segmentUpperLeft, segmentCenterUpper, segmentCenterLower),
	'Х': lineGlyph(segmentUpperLeftDiagonal, segmentUpperRightDiagonal, segmentLowerLeftDiagonal, segmentLowerRightDiagonal),
	'Ц': lineGlyphWithStrokes([]segmentID{segmentUpperLeft, segmentLowerLeft, segmentUpperRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight}, segmentStrokeUnit{x1: 0.84, y1: 0.92, x2: 0.92, y2: 0.98, widthScale: 0.75}),
	'Ч': lineGlyph(segmentUpperLeft, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight),
	'Ш': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentCenterUpper, segmentCenterLower, segmentUpperRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'Щ': lineGlyphWithStrokes([]segmentID{segmentUpperLeft, segmentLowerLeft, segmentCenterUpper, segmentCenterLower, segmentUpperRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight}, segmentStrokeUnit{x1: 0.84, y1: 0.92, x2: 0.92, y2: 0.98, widthScale: 0.75}),
	'Ъ': lineGlyph(segmentTopLeft, segmentCenterUpper, segmentCenterLower, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'Ы': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight, segmentCenterUpper, segmentCenterLower),
	'Ь': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'Э': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentLowerRight, segmentBottomLeft, segmentBottomRight),
	'Ю': lineGlyph(segmentUpperLeft, segmentLowerLeft, segmentMiddleLeft, segmentTopRight, segmentUpperRight, segmentLowerRight, segmentBottomRight),
	'Я': lineGlyph(segmentTopLeft, segmentTopRight, segmentUpperRight, segmentMiddleLeft, segmentMiddleRight, segmentUpperLeftDiagonal, segmentLowerLeftDiagonal),

	'.': customGlyph(segmentStrokeUnit{x1: 0.43, y1: 0.92, x2: 0.57, y2: 0.92, widthScale: 0.8}),
	',': customGlyph(segmentStrokeUnit{x1: 0.54, y1: 0.88, x2: 0.44, y2: 0.98, widthScale: 0.75}),
	':': customGlyph(
		segmentStrokeUnit{x1: 0.50, y1: 0.32, x2: 0.50, y2: 0.42, widthScale: 0.8},
		segmentStrokeUnit{x1: 0.50, y1: 0.58, x2: 0.50, y2: 0.68, widthScale: 0.8},
	),
	';': customGlyph(
		segmentStrokeUnit{x1: 0.50, y1: 0.32, x2: 0.50, y2: 0.42, widthScale: 0.8},
		segmentStrokeUnit{x1: 0.54, y1: 0.58, x2: 0.44, y2: 0.72, widthScale: 0.75},
	),
	'!': customGlyph(
		segmentStrokeUnit{x1: 0.50, y1: 0.14, x2: 0.50, y2: 0.68, widthScale: 0.85},
		segmentStrokeUnit{x1: 0.43, y1: 0.92, x2: 0.57, y2: 0.92, widthScale: 0.8},
	),
	'?': customGlyph(
		segmentStrokeUnit{x1: 0.25, y1: 0.08, x2: 0.75, y2: 0.08, widthScale: 0.9},
		segmentStrokeUnit{x1: 0.78, y1: 0.14, x2: 0.78, y2: 0.38, widthScale: 0.9},
		segmentStrokeUnit{x1: 0.54, y1: 0.50, x2: 0.78, y2: 0.38, widthScale: 0.9},
		segmentStrokeUnit{x1: 0.50, y1: 0.54, x2: 0.50, y2: 0.70, widthScale: 0.85},
		segmentStrokeUnit{x1: 0.43, y1: 0.92, x2: 0.57, y2: 0.92, widthScale: 0.8},
	),
	'-':  segmentPunctuationGlyph(segmentMiddleLeft, segmentMiddleRight),
	'(':  customGlyph(segmentStrokeUnit{x1: 0.62, y1: 0.12, x2: 0.38, y2: 0.50, widthScale: 0.8}, segmentStrokeUnit{x1: 0.38, y1: 0.50, x2: 0.62, y2: 0.88, widthScale: 0.8}),
	')':  customGlyph(segmentStrokeUnit{x1: 0.38, y1: 0.12, x2: 0.62, y2: 0.50, widthScale: 0.8}, segmentStrokeUnit{x1: 0.62, y1: 0.50, x2: 0.38, y2: 0.88, widthScale: 0.8}),
	'/':  customGlyph(segmentStrokeUnit{x1: 0.25, y1: 0.88, x2: 0.75, y2: 0.12, widthScale: 0.85}),
	'\\': customGlyph(segmentStrokeUnit{x1: 0.25, y1: 0.12, x2: 0.75, y2: 0.88, widthScale: 0.85}),
	'\'': customGlyph(segmentStrokeUnit{x1: 0.50, y1: 0.10, x2: 0.50, y2: 0.28, widthScale: 0.75}),
	'"':  customGlyph(segmentStrokeUnit{x1: 0.38, y1: 0.10, x2: 0.38, y2: 0.28, widthScale: 0.7}, segmentStrokeUnit{x1: 0.62, y1: 0.10, x2: 0.62, y2: 0.28, widthScale: 0.7}),
}

func lineGlyph(segments ...segmentID) segmentGlyph {
	return segmentGlyph{
		segments:  segments,
		backplane: true,
	}
}

func lineGlyphWithStrokes(segments []segmentID, strokes ...segmentStrokeUnit) segmentGlyph {
	return segmentGlyph{
		segments:  segments,
		strokes:   strokes,
		backplane: true,
	}
}

func segmentPunctuationGlyph(segments ...segmentID) segmentGlyph {
	return segmentGlyph{segments: segments}
}

func customGlyph(strokes ...segmentStrokeUnit) segmentGlyph {
	return segmentGlyph{strokes: strokes}
}

func drawSegmentDisplayText(img *image.NRGBA, project Project, lines []string) error {
	layout := computeDisplayLayout(project, lines, img.Bounds().Dx(), img.Bounds().Dy(), displayModeSegment)
	mainColor := parseHexColor(project.Style.MainColor, color.NRGBA{R: 255, G: 96, B: 64, A: 255})
	glowColor := parseHexColor(project.Style.GlowColor, color.NRGBA{R: 255, G: 140, B: 102, A: 255})
	inactiveColor := parseHexColor(project.Style.InactiveColor, color.NRGBA{R: 49, G: 20, B: 14, A: 255})
	inactiveColor.A = uint8(math.Round(clampFloat(project.Style.InactiveVisibility, 0, 100) * 1.3))
	activeGlowAlpha := glowAlpha(project.Style.GlowIntensity, 12, 86)

	totalHeight := len(lines)*layout.charHeight + maxInt(len(lines)-1, 0)*layout.lineGap
	startY := maxInt((img.Bounds().Dy()-totalHeight)/2, layout.padding)
	for lineIndex, line := range lines {
		lineRunes := []rune(line)
		lineWidth := len(lineRunes)*layout.charWidth + maxInt(len(lineRunes)-1, 0)*layout.charGap
		x := alignedX(project.Layout.Alignment, img.Bounds().Dx(), lineWidth, layout.padding)
		y := startY + lineIndex*(layout.charHeight+layout.lineGap)

		for _, r := range lineRunes {
			drawSegmentGlyph(img, x, y, layout, segmentGlyphForRune(r), inactiveColor, mainColor, glowColor, activeGlowAlpha)
			x += layout.charWidth + layout.charGap
		}
	}
	return nil
}

func segmentGlyphForRune(r rune) segmentGlyph {
	if r == ' ' || r == '\t' || r == '\n' {
		return segmentGlyph{}
	}
	if glyph, ok := segmentGlyphs[r]; ok {
		return glyph
	}
	if glyph, ok := segmentGlyphs[unicode.ToUpper(r)]; ok {
		return glyph
	}
	return segmentGlyphs['?']
}

func drawSegmentGlyph(img *image.NRGBA, x, y int, layout displayLayout, glyph segmentGlyph, inactiveColor, activeColor, glowColor color.NRGBA, glowAlpha uint8) {
	thickness := segmentStrokeThickness(layout)
	if glyph.backplane && inactiveColor.A > 0 {
		for _, id := range allSegmentIDs {
			drawSegmentUnitStroke(img, x, y, layout, segmentStrokeByID[id], thickness*0.88, inactiveColor)
		}
	}

	for _, id := range glyph.segments {
		stroke := segmentStrokeByID[id]
		drawSegmentGlowStroke(img, x, y, layout, stroke, thickness, glowColor, glowAlpha)
		drawSegmentUnitStroke(img, x, y, layout, stroke, thickness, activeColor)
	}
	for _, stroke := range glyph.strokes {
		drawSegmentGlowStroke(img, x, y, layout, stroke, thickness, glowColor, glowAlpha)
		drawSegmentUnitStroke(img, x, y, layout, stroke, thickness, activeColor)
	}
}

func segmentStrokeThickness(layout displayLayout) float64 {
	return math.Max(2, math.Round(float64(minInt(layout.charWidth, layout.charHeight))*0.105))
}

func drawSegmentGlowStroke(img *image.NRGBA, originX, originY int, layout displayLayout, stroke segmentStrokeUnit, thickness float64, glowColor color.NRGBA, glowAlpha uint8) {
	if glowAlpha == 0 {
		return
	}
	outer := color.NRGBA{R: glowColor.R, G: glowColor.G, B: glowColor.B, A: glowAlpha / 3}
	inner := color.NRGBA{R: glowColor.R, G: glowColor.G, B: glowColor.B, A: glowAlpha / 2}
	drawSegmentUnitStroke(img, originX, originY, layout, stroke, thickness*2.2, outer)
	drawSegmentUnitStroke(img, originX, originY, layout, stroke, thickness*1.45, inner)
}

func drawSegmentUnitStroke(img *image.NRGBA, originX, originY int, layout displayLayout, stroke segmentStrokeUnit, thickness float64, col color.NRGBA) {
	if col.A == 0 {
		return
	}
	polygon, ok := segmentStrokePolygon(originX, originY, layout, stroke, thickness)
	if !ok {
		return
	}
	fillPolygon(img, polygon, col)
}

func segmentStrokePolygon(originX, originY int, layout displayLayout, stroke segmentStrokeUnit, thickness float64) ([]segmentPoint, bool) {
	widthScale := stroke.widthScale
	if widthScale <= 0 {
		widthScale = 1
	}
	x1 := float64(originX) + stroke.x1*float64(layout.charWidth)
	y1 := float64(originY) + stroke.y1*float64(layout.charHeight)
	x2 := float64(originX) + stroke.x2*float64(layout.charWidth)
	y2 := float64(originY) + stroke.y2*float64(layout.charHeight)
	dx := x2 - x1
	dy := y2 - y1
	length := math.Hypot(dx, dy)
	if length < 0.01 {
		return nil, false
	}

	ux := dx / length
	uy := dy / length
	nx := -uy
	ny := ux
	half := thickness * widthScale / 2
	bevel := math.Min(half*0.95, length*0.34)

	return []segmentPoint{
		{x: x1 + ux*bevel - nx*half, y: y1 + uy*bevel - ny*half},
		{x: x2 - ux*bevel - nx*half, y: y2 - uy*bevel - ny*half},
		{x: x2, y: y2},
		{x: x2 - ux*bevel + nx*half, y: y2 - uy*bevel + ny*half},
		{x: x1 + ux*bevel + nx*half, y: y1 + uy*bevel + ny*half},
		{x: x1, y: y1},
	}, true
}

func fillPolygon(img *image.NRGBA, points []segmentPoint, col color.NRGBA) {
	if len(points) < 3 {
		return
	}

	minX, minY := points[0].x, points[0].y
	maxX, maxY := points[0].x, points[0].y
	for _, p := range points[1:] {
		minX = math.Min(minX, p.x)
		minY = math.Min(minY, p.y)
		maxX = math.Max(maxX, p.x)
		maxY = math.Max(maxY, p.y)
	}

	startX := maxInt(int(math.Floor(minX)), img.Bounds().Min.X)
	endX := minInt(int(math.Ceil(maxX)), img.Bounds().Max.X-1)
	startY := maxInt(int(math.Floor(minY)), img.Bounds().Min.Y)
	endY := minInt(int(math.Ceil(maxY)), img.Bounds().Max.Y-1)

	for py := startY; py <= endY; py++ {
		for px := startX; px <= endX; px++ {
			if pointInPolygon(float64(px)+0.5, float64(py)+0.5, points) {
				blendNRGBA(img, px, py, col)
			}
		}
	}
}

func pointInPolygon(x, y float64, points []segmentPoint) bool {
	inside := false
	j := len(points) - 1
	for i := range points {
		yi := points[i].y
		yj := points[j].y
		if (yi > y) != (yj > y) {
			xIntersect := (points[j].x-points[i].x)*(y-yi)/(yj-yi) + points[i].x
			if x < xIntersect {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}
