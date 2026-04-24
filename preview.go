package main

import (
	"image"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const (
	previewBaseWidth  = 960
	previewBaseHeight = 430

	previewBGProject = "Project"
	previewBGBlack   = "Black"
	previewBGCustom  = "Custom"
)

type previewCard struct {
	card              *widget.Card
	scroll            *container.Scroll
	content           *fyne.Container
	bg                *canvas.Rectangle
	checker           *canvas.Raster
	image             *canvas.Image
	safeAreaLines     [4]*canvas.Line
	subtitle          *canvas.Text
	zoomSlider        *widget.Slider
	bgModeSelect      *widget.Select
	customColorBtn    *widget.Button
	checkerCheck      *widget.Check
	safeAreaCheck     *widget.Check
	window            fyne.Window
	project           Project
	zoom              float64
	bgMode            string
	customBGColor     color.NRGBA
	checkerboardOn    bool
	safeAreaVisible   bool
	onViewportChanged func(ExportRegion)
}

func newPreviewCard(cardTitle string, window fyne.Window) *previewCard {
	card := &previewCard{
		window:        window,
		zoom:          1,
		bgMode:        previewBGProject,
		customBGColor: color.NRGBA{R: 24, G: 28, B: 34, A: 255},
	}

	card.bg = canvas.NewRectangle(color.NRGBA{R: 5, G: 6, B: 8, A: 255})
	card.checker = canvas.NewRasterWithPixels(func(x, y, _, _ int) color.Color {
		cell := 18
		light := color.NRGBA{R: 114, G: 118, B: 126, A: 255}
		dark := color.NRGBA{R: 76, G: 80, B: 88, A: 255}
		if ((x/cell)+(y/cell))%2 == 0 {
			return light
		}
		return dark
	})
	card.checker.Hide()

	card.image = canvas.NewImageFromImage(image.NewNRGBA(image.Rect(0, 0, previewBaseWidth, previewBaseHeight)))
	card.image.FillMode = canvas.ImageFillStretch
	card.image.ScaleMode = canvas.ImageScaleFastest

	card.content = container.NewWithoutLayout(card.bg, card.checker, card.image)
	for i := range card.safeAreaLines {
		line := canvas.NewLine(color.NRGBA{R: 255, G: 214, B: 102, A: 230})
		line.StrokeWidth = 1
		line.Hide()
		card.safeAreaLines[i] = line
		card.content.Add(line)
	}
	card.layoutPreview()

	card.scroll = container.NewScroll(card.content)
	card.scroll.SetMinSize(fyne.NewSize(previewBaseWidth, previewBaseHeight))
	card.scroll.OnScrolled = func(_ fyne.Position) {
		card.notifyViewportChanged()
	}

	card.subtitle = canvas.NewText("", color.NRGBA{R: 130, G: 136, B: 145, A: 255})
	card.subtitle.TextSize = 15
	card.subtitle.Alignment = fyne.TextAlignCenter

	card.zoomSlider = widget.NewSlider(1, 3)
	card.zoomSlider.Step = 0.1
	card.zoomSlider.SetValue(1)
	card.zoomSlider.OnChanged = func(value float64) {
		card.zoom = value
		card.layoutPreview()
		card.notifyViewportChanged()
	}

	card.bgModeSelect = widget.NewSelect([]string{previewBGProject, previewBGBlack, previewBGCustom}, func(value string) {
		card.bgMode = value
		card.refreshBackground()
		card.refreshControls()
	})
	card.bgModeSelect.SetSelected(previewBGProject)

	card.customColorBtn = widget.NewButton("", func() {
		if card.window == nil {
			return
		}
		picker := dialog.NewColorPicker("Preview background", "Select a preview background color", func(c color.Color) {
			card.customBGColor = parseHexColor(colorToHex(c), card.customBGColor)
			card.refreshBackground()
			card.refreshControls()
		}, card.window)
		picker.Advanced = true
		picker.Show()
		picker.SetColor(card.customBGColor)
	})

	card.checkerCheck = widget.NewCheck("Checkerboard", func(value bool) {
		card.checkerboardOn = value
		card.refreshBackground()
	})
	card.safeAreaCheck = widget.NewCheck("Safe area", func(value bool) {
		card.safeAreaVisible = value
		card.refreshSafeArea()
	})

	card.refreshControls()
	card.refreshBackground()
	card.refreshSafeArea()

	controls := container.NewVBox(
		container.NewGridWithColumns(4,
			widget.NewLabel("Zoom"),
			card.zoomSlider,
			card.bgModeSelect,
			card.customColorBtn,
		),
		container.NewGridWithColumns(2, card.checkerCheck, card.safeAreaCheck),
	)

	card.card = widget.NewCard(cardTitle, "", container.NewBorder(
		controls,
		container.NewPadded(card.subtitle),
		nil,
		nil,
		card.scroll,
	))
	return card
}

func (p *previewCard) object() fyne.CanvasObject {
	return p.card
}

func (p *previewCard) applyProject(project Project, stats textStats, frame int) {
	p.project = project

	renderWidth := previewBaseWidth
	renderHeight := previewBaseHeight
	if p.zoom > 1 {
		renderWidth = int(float64(previewBaseWidth) * p.zoom)
		renderHeight = int(float64(previewBaseHeight) * p.zoom)
	}

	rendered, err := renderImage(project, stats, frame, renderWidth, renderHeight)
	if err == nil {
		p.image.Image = rendered
		p.image.Refresh()
	}
	p.subtitle.Text = previewSubtitle(project, stats, frame)
	p.subtitle.Refresh()
	p.refreshBackground()
}

func (p *previewCard) layoutPreview() {
	size := fyne.NewSize(float32(float64(previewBaseWidth)*p.zoom), float32(float64(previewBaseHeight)*p.zoom))
	p.content.Resize(size)

	p.bg.Move(fyne.NewPos(0, 0))
	p.bg.Resize(size)

	p.checker.Move(fyne.NewPos(0, 0))
	p.checker.Resize(size)

	p.image.Move(fyne.NewPos(0, 0))
	p.image.Resize(size)

	p.refreshSafeArea()
	p.content.Refresh()
	if p.scroll != nil {
		p.scroll.Refresh()
	}
}

func (p *previewCard) visibleRegion() ExportRegion {
	if p.scroll == nil || p.content == nil {
		return fullExportRegion()
	}

	contentSize := p.content.Size()
	if contentSize.Width <= 0 || contentSize.Height <= 0 {
		return fullExportRegion()
	}

	viewportSize := p.scroll.Size()
	if viewportSize.Width <= 0 || viewportSize.Height <= 0 {
		return fullExportRegion()
	}

	region := ExportRegion{
		X:      float64(p.scroll.Offset.X / contentSize.Width),
		Y:      float64(p.scroll.Offset.Y / contentSize.Height),
		Width:  float64(minFloat32(viewportSize.Width, contentSize.Width) / contentSize.Width),
		Height: float64(minFloat32(viewportSize.Height, contentSize.Height) / contentSize.Height),
	}
	return normalizeExportRegion(region)
}

func (p *previewCard) notifyViewportChanged() {
	if p.onViewportChanged != nil {
		p.onViewportChanged(p.visibleRegion())
	}
}

func (p *previewCard) refreshBackground() {
	p.bg.FillColor = p.backgroundColor()
	p.bg.Refresh()
	if p.checkerboardOn {
		p.checker.Show()
	} else {
		p.checker.Hide()
	}
	p.checker.Refresh()
}

func (p *previewCard) backgroundColor() color.NRGBA {
	switch p.bgMode {
	case previewBGBlack:
		return color.NRGBA{A: 255}
	case previewBGCustom:
		return p.customBGColor
	default:
		return previewBackground(p.project)
	}
}

func (p *previewCard) refreshControls() {
	if p.customColorBtn == nil {
		return
	}
	p.customColorBtn.SetText("BG " + colorToHex(p.customBGColor))
	if p.bgMode == previewBGCustom {
		p.customColorBtn.Enable()
	} else {
		p.customColorBtn.Disable()
	}
}

func (p *previewCard) refreshSafeArea() {
	if p.content == nil {
		return
	}
	size := p.content.Size()
	x1 := size.Width * 0.05
	y1 := size.Height * 0.05
	x2 := size.Width - x1
	y2 := size.Height - y1

	p.safeAreaLines[0].Position1 = fyne.NewPos(x1, y1)
	p.safeAreaLines[0].Position2 = fyne.NewPos(x2, y1)
	p.safeAreaLines[1].Position1 = fyne.NewPos(x2, y1)
	p.safeAreaLines[1].Position2 = fyne.NewPos(x2, y2)
	p.safeAreaLines[2].Position1 = fyne.NewPos(x2, y2)
	p.safeAreaLines[2].Position2 = fyne.NewPos(x1, y2)
	p.safeAreaLines[3].Position1 = fyne.NewPos(x1, y2)
	p.safeAreaLines[3].Position2 = fyne.NewPos(x1, y1)

	for _, line := range p.safeAreaLines {
		if p.safeAreaVisible {
			line.Show()
		} else {
			line.Hide()
		}
		line.Refresh()
	}
}

func previewSubtitle(project Project, stats textStats, frame int) string {
	frameState := buildAnimatedFrame(project, stats, frame)
	return strings.Join([]string{
		project.Display.Mode,
		strings.Join(sortedLanguages(project.Charset.Languages), " + "),
		project.Background.Mode,
		"Frame " + strconv.Itoa(frameState.Frame),
		"Unsupported: " + formatUnsupportedRunes(stats.UnsupportedUnique),
	}, " • ")
}

func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func previewBackground(project Project) color.NRGBA {
	switch project.Background.Mode {
	case backgroundModeSolid:
		return parseHexColor(project.Background.SolidColor, color.NRGBA{R: 5, G: 6, B: 8, A: 255})
	case backgroundModeGradient:
		return mixColors(
			parseHexColor(project.Background.GradientA, color.NRGBA{R: 6, G: 10, B: 13, A: 255}),
			parseHexColor(project.Background.GradientB, color.NRGBA{R: 18, G: 38, B: 43, A: 255}),
		)
	case backgroundModeImage:
		return color.NRGBA{R: 18, G: 22, B: 28, A: 255}
	case backgroundModeVideo:
		return color.NRGBA{R: 24, G: 18, B: 24, A: 255}
	default:
		return color.NRGBA{R: 5, G: 6, B: 8, A: 255}
	}
}

func mixColors(a, b color.NRGBA) color.NRGBA {
	return color.NRGBA{
		R: uint8((uint16(a.R) + uint16(b.R)) / 2),
		G: uint8((uint16(a.G) + uint16(b.G)) / 2),
		B: uint8((uint16(a.B) + uint16(b.B)) / 2),
		A: 255,
	}
}
