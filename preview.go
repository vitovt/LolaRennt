package main

import (
	"image"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type previewCard struct {
	card     *widget.Card
	bg       *canvas.Rectangle
	image    *canvas.Image
	subtitle *canvas.Text
}

func newPreviewCard(cardTitle string) *previewCard {
	bg := canvas.NewRectangle(color.NRGBA{R: 5, G: 6, B: 8, A: 255})
	bg.SetMinSize(fyne.NewSize(0, 430))

	img := canvas.NewImageFromImage(image.NewNRGBA(image.Rect(0, 0, 960, 430)))
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest

	subtitle := canvas.NewText("", color.NRGBA{R: 130, G: 136, B: 145, A: 255})
	subtitle.TextSize = 15
	subtitle.Alignment = fyne.TextAlignCenter

	frame := container.NewMax(
		bg,
		container.NewBorder(nil, container.NewPadded(subtitle), nil, nil, container.NewPadded(img)),
	)

	return &previewCard{
		card:     widget.NewCard(cardTitle, "", frame),
		bg:       bg,
		image:    img,
		subtitle: subtitle,
	}
}

func (p *previewCard) object() fyne.CanvasObject {
	return p.card
}

func (p *previewCard) applyProject(project Project, stats textStats, frame int) {
	rendered, err := renderImage(project, stats, frame, 960, 430)
	if err == nil {
		p.image.Image = rendered
		p.image.Refresh()
	}
	p.subtitle.Text = previewSubtitle(project, stats, frame)
	p.subtitle.Refresh()
	p.bg.FillColor = previewBackground(project)
	p.bg.Refresh()
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
