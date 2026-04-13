package main

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type previewCard struct {
	card     *widget.Card
	bg       *canvas.Rectangle
	title    *canvas.Text
	body     *canvas.Text
	subtitle *canvas.Text
}

func newPreviewCard(cardTitle string) *previewCard {
	bg := canvas.NewRectangle(color.NRGBA{R: 5, G: 6, B: 8, A: 255})
	bg.SetMinSize(fyne.NewSize(0, 430))

	title := canvas.NewText("", color.NRGBA{R: 255, G: 186, B: 128, A: 255})
	title.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	title.TextSize = 16
	title.Alignment = fyne.TextAlignCenter

	body := canvas.NewText("", color.NRGBA{R: 255, G: 96, B: 64, A: 255})
	body.TextStyle = fyne.TextStyle{Monospace: true}
	body.TextSize = 54
	body.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("", color.NRGBA{R: 130, G: 136, B: 145, A: 255})
	subtitle.TextSize = 15
	subtitle.Alignment = fyne.TextAlignCenter

	frame := container.NewMax(
		bg,
		container.NewPadded(
			container.NewCenter(container.NewVBox(
				layout.NewSpacer(),
				title,
				body,
				subtitle,
				layout.NewSpacer(),
			)),
		),
	)

	return &previewCard{
		card:     widget.NewCard(cardTitle, "", frame),
		bg:       bg,
		title:    title,
		body:     body,
		subtitle: subtitle,
	}
}

func (p *previewCard) object() fyne.CanvasObject {
	return p.card
}

func (p *previewCard) applyProject(project Project, stats textStats, label string) {
	p.title.Text = label
	p.title.Refresh()

	displayText := stats.DisplayText
	if strings.TrimSpace(displayText) == "" {
		displayText = " "
	}
	p.body.Text = displayText
	p.body.Color = parseHexColor(project.Style.MainColor, color.NRGBA{R: 255, G: 96, B: 64, A: 255})
	p.body.TextStyle = fyne.TextStyle{Monospace: true}
	if project.Display.Mode == displayModeDotMatrix {
		p.body.TextSize = 46
	} else {
		p.body.TextSize = 54
	}
	p.body.Alignment = textAlign(project.Layout.Alignment)
	p.body.Refresh()

	p.subtitle.Text = previewSubtitle(project, stats)
	p.subtitle.Refresh()
	p.bg.FillColor = previewBackground(project)
	p.bg.Refresh()
}

func previewSubtitle(project Project, stats textStats) string {
	return strings.Join([]string{
		project.Display.Mode,
		strings.Join(sortedLanguages(project.Charset.Languages), " + "),
		project.Background.Mode,
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

func textAlign(alignment string) fyne.TextAlign {
	switch alignment {
	case alignmentLeft:
		return fyne.TextAlignLeading
	case alignmentRight:
		return fyne.TextAlignTrailing
	default:
		return fyne.TextAlignCenter
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
