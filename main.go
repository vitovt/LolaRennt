package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Segment Text Animator — Fyne Mockup")
	w.Resize(fyne.NewSize(1460, 920))

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Текст і стиль", theme.DocumentCreateIcon(), buildTextAndStyleTab()),
		container.NewTabItemWithIcon("Анімація / Preview", theme.MediaPlayIcon(), buildAnimationTab()),
		container.NewTabItemWithIcon("Експорт", theme.DownloadIcon(), buildExportTab()),
		container.NewTabItemWithIcon("Проєкт / Пресети", theme.SettingsIcon(), buildProjectTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	status := widget.NewLabel("Готово • Mockup інтерфейсу • Go + Fyne")
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {}),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {}),
		widget.NewToolbarAction(theme.MediaPlayIcon(), func() {}),
		widget.NewToolbarAction(theme.MediaStopIcon(), func() {}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DownloadIcon(), func() {}),
	)

	content := container.NewBorder(toolbar, status, nil, nil, tabs)
	w.SetContent(content)
	w.ShowAndRun()
}

func buildTextAndStyleTab() fyne.CanvasObject {
	charsetSelect := widget.NewSelect([]string{
		"English only",
		"German only",
		"Cyrillic only",
		"Cyrillic + German",
		"Cyrillic + English",
		"Custom",
	}, func(string) {})
	charsetSelect.SetSelected("Cyrillic + German")

	displayType := widget.NewRadioGroup([]string{"Segment", "Dot-matrix"}, func(string) {})
	displayType.SetSelected("Segment")
	glyphPreset := widget.NewSelect([]string{"Segment 16 + umlaut", "Segment 20", "Matrix 5x7", "Matrix 5x8"}, func(string) {})
	glyphPreset.SetSelected("Segment 16 + umlaut")

	textInput := widget.NewMultiLineEntry()
	textInput.SetText("LOLA: RENNT\nDIE: DEIN: LEBEN\nVERÄNDERN: KANN")
	textInput.Wrapping = fyne.TextWrapWord

	mainColor := widget.NewButton("Основний колір", func() {})
	glowColor := widget.NewButton("Glow", func() {})
	inactiveColor := widget.NewButton("Неактивні сегменти", func() {})

	left := container.NewVScroll(container.NewVBox(
		sectionCard("Charset / Font profile", container.NewVBox(
			widget.NewCheck("Uppercase only", func(bool) {}),
			widget.NewCheck("Автозаміна неподтримуваних символів", func(bool) {}),
			widget.NewLabel("Профіль символів"),
			charsetSelect,
			widget.NewButton("Показати таблицю підтримуваних символів", func() {}),
		)),
		sectionCard("Display type", container.NewVBox(
			widget.NewLabel("Тип індикатора"),
			displayType,
			widget.NewLabel("Glyph preset"),
			glyphPreset,
		)),
		sectionCard("Текст", container.NewVBox(
			textInput,
			container.NewGridWithColumns(3,
				widget.NewButton("Очистити", func() {}),
				widget.NewButton("Вставити", func() {}),
				widget.NewButton("Normalize", func() {}),
			),
			widget.NewLabel("Символів: 48 • Рядків: 3 • Unsupported: 0"),
		)),
		sectionCard("Колір", container.NewVBox(
			widget.NewSelect([]string{"Single color", "Per-character colors", "Gradient", "Palette cycle"}, func(string) {}),
			container.NewGridWithColumns(3, mainColor, glowColor, inactiveColor),
			widget.NewLabel("Opacity"),
			widget.NewSlider(0, 100),
			widget.NewLabel("Glow intensity"),
			widget.NewSlider(0, 100),
			widget.NewLabel("Inactive segment visibility"),
			widget.NewSlider(0, 100),
		)),
		sectionCard("Layout", container.NewVBox(
			widget.NewLabel("Cell scale"),
			widget.NewSlider(0.5, 3.0),
			widget.NewLabel("Character spacing"),
			widget.NewSlider(0, 60),
			widget.NewLabel("Line spacing"),
			widget.NewSlider(0, 80),
			widget.NewSelect([]string{"Left", "Center", "Right"}, func(string) {}),
			widget.NewCheck("Auto-wrap", func(bool) {}),
			widget.NewCheck("Fixed cell width", func(bool) {}),
		)),
	))
	left.SetMinSize(fyne.NewSize(410, 0))

	previewBox := previewPanel("Статичне preview", "Миттєве оновлення під час вводу")
	rightTools := sectionCard("Швидкі дії", container.NewGridWithColumns(3,
		widget.NewButtonWithIcon("Оновити", theme.ViewRefreshIcon(), func() {}),
		widget.NewButtonWithIcon("Запустити", theme.MediaPlayIcon(), func() {}),
		widget.NewButtonWithIcon("Стоп", theme.MediaStopIcon(), func() {}),
	))
	rightInfo := sectionCard("Поточний стиль", container.NewVBox(
		widget.NewLabel("Preset: Red LED Classic"),
		widget.NewLabel("Mode: Segment 16 + umlaut"),
		widget.NewLabel("Charset: Cyrillic + German"),
		widget.NewLabel("Preview background: checkerboard"),
	))

	right := container.NewBorder(nil, nil, nil, nil, container.NewVBox(previewBox, rightTools, rightInfo))

	split := container.NewHSplit(left, right)
	split.Offset = 0.34
	return split
}

func buildAnimationTab() fyne.CanvasObject {
	preview := previewPanel("Animation preview", "Scramble → lock → final text")
	timeline := widget.NewSlider(0, 100)
	timeline.Value = 37

	left := container.NewVScroll(container.NewVBox(
		sectionCard("Тип анімації", container.NewVBox(
			widget.NewSelect([]string{
				"Scramble basic",
				"Scramble with lock",
				"Reveal then scramble then lock",
				"Type-on + scramble",
				"Cell reveal left-to-right",
				"Cell reveal random",
			}, func(string) {}),
			widget.NewLabel("Random source"),
			widget.NewSelect([]string{"Digits only", "Letters only", "Alphanumeric", "Current charset only", "Custom pool"}, func(string) {}),
			widget.NewCheck("Дозволити невалідні символи", func(bool) {}),
			widget.NewCheck("Дозволити порожню комірку", func(bool) {}),
		)),
		sectionCard("Timing", container.NewVBox(
			widget.NewLabel("Total duration"),
			widget.NewSlider(0.2, 12),
			widget.NewLabel("Intro delay"),
			widget.NewSlider(0, 3),
			widget.NewLabel("Per-character delay"),
			widget.NewSlider(0, 1.5),
			widget.NewLabel("Random switch rate"),
			widget.NewSlider(1, 60),
			widget.NewLabel("Seed"),
			widget.NewEntry(),
			widget.NewButton("Randomize seed", func() {}),
		)),
		sectionCard("Reveal / Lock logic", container.NewVBox(
			widget.NewLabel("Order"),
			widget.NewSelect([]string{"Left-to-right", "Right-to-left", "Center-out", "Random", "By words", "By lines"}, func(string) {}),
			widget.NewLabel("Lock mode"),
			widget.NewSelect([]string{"Hard lock", "Probabilistic lock", "Overshoot flicker then lock"}, func(string) {}),
			widget.NewCheck("Всі символи фіксуються одночасно в кінці", func(bool) {}),
			widget.NewCheck("Фіксувати розділові знаки одразу", func(bool) {}),
		)),
		sectionCard("Visual effects", container.NewVBox(
			labeledSlider("Glow", 0, 100),
			labeledSlider("Bloom", 0, 100),
			labeledSlider("Flicker", 0, 100),
			labeledSlider("Noise", 0, 100),
			labeledSlider("Ghost trail", 0, 100),
			widget.NewCheck("CRT-style softness", func(bool) {}),
			widget.NewCheck("Scanline effect", func(bool) {}),
		)),
	))
	left.SetMinSize(fyne.NewSize(410, 0))

	transport := sectionCard("Playback", container.NewVBox(
		container.NewGridWithColumns(4,
			widget.NewButtonWithIcon("Play", theme.MediaPlayIcon(), func() {}),
			widget.NewButtonWithIcon("Pause", theme.MediaPauseIcon(), func() {}),
			widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {}),
			widget.NewButtonWithIcon("Loop", theme.ViewRefreshIcon(), func() {}),
		),
		widget.NewLabel("Timeline"),
		timeline,
		widget.NewLabel("Current time: 01.23 s • Frame: 37 / 120 • Preview FPS: 30"),
	))

	meta := sectionCard("Стан анімації", container.NewVBox(
		widget.NewLabel("Preset: Dramatic lock"),
		widget.NewLabel("Random source: Alphanumeric"),
		widget.NewLabel("Lock order: Center-out"),
		widget.NewLabel("Target text fixed at t=82%"),
	))

	right := container.NewVBox(preview, transport, meta)
	split := container.NewHSplit(left, right)
	split.Offset = 0.34
	return split
}

func buildExportTab() fyne.CanvasObject {
	left := container.NewVScroll(container.NewVBox(
		sectionCard("Output", container.NewVBox(
			widget.NewSelect([]string{"PNG sequence", "Single PNG current frame", "MP4 with baked background", "MOV ProRes 4444"}, func(string) {}),
			widget.NewEntry(),
			widget.NewLabel("Output folder"),
			widget.NewButtonWithIcon("Вибрати теку", theme.FolderOpenIcon(), func() {}),
			widget.NewLabel("Overwrite policy"),
			widget.NewSelect([]string{"Ask", "Overwrite", "Create new suffix"}, func(string) {}),
		)),
		sectionCard("Render size", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewEntry(),
				widget.NewEntry(),
			),
			widget.NewSelect([]string{"1280x720", "1920x1080", "2560x1440", "3840x2160", "Custom"}, func(string) {}),
			widget.NewLabel("Supersampling"),
			widget.NewSlider(1, 4),
			widget.NewLabel("Anti-aliasing"),
			widget.NewSelect([]string{"Off", "2x", "4x", "8x"}, func(string) {}),
		)),
		sectionCard("Frame settings", container.NewVBox(
			widget.NewLabel("FPS"),
			widget.NewSlider(1, 120),
			widget.NewLabel("Start frame"),
			widget.NewEntry(),
			widget.NewLabel("End frame"),
			widget.NewEntry(),
			widget.NewCheck("Export full canvas", func(bool) {}),
		)),
		sectionCard("Background", container.NewVBox(
			widget.NewSelect([]string{"Transparent", "Solid color", "Gradient", "Image", "Image + blur"}, func(string) {}),
			container.NewGridWithColumns(3,
				widget.NewButton("BG Color", func() {}),
				widget.NewButton("Gradient A", func() {}),
				widget.NewButton("Gradient B", func() {}),
			),
			widget.NewButtonWithIcon("Вибрати зображення", theme.FolderOpenIcon(), func() {}),
			widget.NewLabel("Image fit mode"),
			widget.NewSelect([]string{"Fit", "Fill", "Center", "Stretch"}, func(string) {}),
			widget.NewLabel("Blur"),
			widget.NewSlider(0, 100),
		)),
		sectionCard("Encoding", container.NewVBox(
			widget.NewLabel("Codec preset"),
			widget.NewSelect([]string{"H.264 Preview", "H.265", "ProRes 422", "ProRes 4444"}, func(string) {}),
			widget.NewLabel("FFmpeg path"),
			widget.NewEntry(),
			widget.NewCheck("Delete PNG after encode", func(bool) {}),
			widget.NewCheck("Keep intermediate frames", func(bool) {}),
		)),
	))
	left.SetMinSize(fyne.NewSize(410, 0))

	preview := previewPanel("Export preview", "Фінальний кадр / вибраний фон / safe area")
	queue := sectionCard("Render control", container.NewVBox(
		container.NewGridWithColumns(3,
			widget.NewButtonWithIcon("PNG sequence", theme.DocumentIcon(), func() {}),
			widget.NewButtonWithIcon("Encode video", theme.MediaPlayIcon(), func() {}),
			widget.NewButtonWithIcon("Render + Encode", theme.DownloadIcon(), func() {}),
		),
		widget.NewProgressBar(),
		widget.NewLabel("ETA: 00:43 • Status: waiting for render"),
		widget.NewMultiLineEntry(),
		container.NewGridWithColumns(2,
			widget.NewButton("Cancel", func() {}),
			widget.NewButton("Open output folder", func() {}),
		),
	))
	alpha := sectionCard("Alpha / transparency", container.NewVBox(
		widget.NewCheck("Render alpha channel", func(bool) {}),
		widget.NewSelect([]string{"Straight alpha", "Premultiplied alpha"}, func(string) {}),
		widget.NewLabel("MP4 = delivery with baked background"),
		widget.NewLabel("PNG / ProRes 4444 = монтажний master"),
	))

	right := container.NewVBox(preview, queue, alpha)
	split := container.NewHSplit(left, right)
	split.Offset = 0.34
	return split
}

func buildProjectTab() fyne.CanvasObject {
	left := container.NewVScroll(container.NewVBox(
		sectionCard("Project", container.NewGridWithColumns(2,
			widget.NewButtonWithIcon("New", theme.DocumentCreateIcon(), func() {}),
			widget.NewButtonWithIcon("Open", theme.FolderOpenIcon(), func() {}),
			widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {}),
			widget.NewButtonWithIcon("Save As", theme.DocumentSaveIcon(), func() {}),
		)),
		sectionCard("Metadata", container.NewVBox(
			widget.NewLabel("Project name"),
			widget.NewEntry(),
			widget.NewLabel("Notes"),
			widget.NewMultiLineEntry(),
			widget.NewLabel("Tags"),
			widget.NewEntry(),
			widget.NewCheck("Autosave", func(bool) {}),
		)),
		sectionCard("Presets", container.NewVBox(
			widget.NewButton("Save style preset", func() {}),
			widget.NewButton("Save animation preset", func() {}),
			widget.NewButton("Save export preset", func() {}),
			widget.NewButton("Import preset", func() {}),
		)),
	))
	left.SetMinSize(fyne.NewSize(380, 0))

	recent := widget.NewList(
		func() int { return 6 },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Template Project"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*fyne.Container).Objects[1].(*widget.Label)
			switch id {
			case 0:
				label.SetText("Lola_Rennt_Title.json")
			case 1:
				label.SetText("German_Phrase_Orange_VFD.json")
			case 2:
				label.SetText("Cyrillic_Green_Matrix.json")
			case 3:
				label.SetText("4K_Transparent_Master.json")
			case 4:
				label.SetText("Shorts_Vertical_Red_LED.json")
			case 5:
				label.SetText("Cyber_Green_Preset.json")
			}
		},
	)
	recentBox := container.NewGridWrap(fyne.NewSize(300, 220), recent)

	styles := widget.NewList(
		func() int { return 5 },
		func() fyne.CanvasObject {
			return widget.NewLabel("Preset")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			names := []string{"Red LED Classic", "Orange VFD", "Cold White LCD", "Neon Cyan", "Cyber Green"}
			label.SetText(names[id])
		},
	)
	stylesBox := container.NewGridWrap(fyne.NewSize(300, 160), styles)

	right := container.NewVScroll(container.NewVBox(
		sectionCard("Recent projects", recentBox),
		sectionCard("Available presets", stylesBox),
		sectionCard("Selected item", container.NewVBox(
			widget.NewLabel("Name: Lola_Rennt_Title.json"),
			widget.NewLabel("Updated: today, 18:20"),
			widget.NewLabel("Style: Red LED Classic"),
			widget.NewLabel("Export: 1920x1080 / 30 fps / transparent PNG"),
			container.NewGridWithColumns(3,
				widget.NewButton("Apply", func() {}),
				widget.NewButton("Duplicate", func() {}),
				widget.NewButton("Delete", func() {}),
			),
		)),
	))

	split := container.NewHSplit(left, right)
	split.Offset = 0.31
	return split
}

func sectionCard(title string, content fyne.CanvasObject) fyne.CanvasObject {
	return widget.NewCard(title, "", content)
}

func labeledSlider(name string, min, max float64) fyne.CanvasObject {
	s := widget.NewSlider(min, max)
	s.Value = (max - min) * 0.45
	return container.NewVBox(widget.NewLabel(name), s)
}

func previewPanel(title, subtitle string) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.RGBA{R: 8, G: 8, B: 8, A: 255})
	bg.SetMinSize(fyne.NewSize(860, 430))

	glow := canvas.NewText("LOLA: RENNT", color.RGBA{R: 255, G: 96, B: 64, A: 255})
	glow.TextSize = 54
	glow.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText(subtitle, color.RGBA{R: 140, G: 140, B: 140, A: 255})
	sub.TextSize = 16
	sub.Alignment = fyne.TextAlignCenter

	header := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	toolbar := container.NewHBox(
		widget.NewButtonWithIcon("100%", theme.SearchIcon(), func() {}),
		widget.NewButtonWithIcon("Checker", theme.SettingsIcon(), func() {}),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("PNG", theme.DocumentIcon(), func() {}),
	)

	frame := container.NewMax(
		bg,
		container.NewCenter(container.NewVBox(
			layout.NewSpacer(),
			glow,
			sub,
			layout.NewSpacer(),
		)),
	)

	return widget.NewCard("Preview", "", container.NewBorder(header, toolbar, nil, nil, frame))
}
