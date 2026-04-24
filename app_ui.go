package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type appUI struct {
	app         fyne.App
	window      fyne.Window
	project     Project
	projectPath string
	suspend     bool

	status *widget.Label

	textInput         *widget.Entry
	languageChecks    *widget.CheckGroup
	uppercaseCheck    *widget.Check
	autoReplaceCheck  *widget.Check
	displayType       *widget.RadioGroup
	colorModeSelect   *widget.Select
	mainColorButton   *widget.Button
	glowColorButton   *widget.Button
	inactiveColorBtn  *widget.Button
	glowSlider        *widget.Slider
	inactiveSlider    *widget.Slider
	cellScaleSlider   *widget.Slider
	charSpacingSlider *widget.Slider
	lineSpacingSlider *widget.Slider
	paddingSlider     *widget.Slider
	alignmentSelect   *widget.Select
	statsLabel        *widget.Label
	validationPreview *widget.Label
	validationGrid    *widget.TextGrid
	replacementGuide  *widget.Label
	styleSummary      *widget.Label

	animationTypeSelect *widget.Select
	randomSourceSelect  *widget.Select
	allowInvalidCheck   *widget.Check
	allowEmptyCheck     *widget.Check
	totalDurationSlider *widget.Slider
	introDelaySlider    *widget.Slider
	outroHoldSlider     *widget.Slider
	perCharDelaySlider  *widget.Slider
	switchRateSlider    *widget.Slider
	seedEntry           *widget.Entry
	lockOrderSelect     *widget.Select
	lockModeSelect      *widget.Select
	loopCheck           *widget.Check
	timelineSlider      *widget.Slider
	playbackInfo        *widget.Label
	animationSummary    *widget.Label
	playbackStop        chan struct{}
	playbackRunning     bool

	backgroundModeSelect *widget.Select
	bgColorButton        *widget.Button
	gradientAButton      *widget.Button
	gradientBButton      *widget.Button
	imagePathEntry       *widget.Entry
	bgImageOpacitySlider *widget.Slider
	videoPathEntry       *widget.Entry
	bgFitSelect          *widget.Select
	widthEntry           *widget.Entry
	heightEntry          *widget.Entry
	fpsEntry             *widget.Entry
	startFrameEntry      *widget.Entry
	endFrameEntry        *widget.Entry
	outputFolderEntry    *widget.Entry
	filePrefixEntry      *widget.Entry
	overwriteSelect      *widget.Select
	supersamplingSlider  *widget.Slider
	ffmpegCommand        *widget.Entry
	exportSummary        *widget.Label
	exportProgress       *widget.ProgressBar
	exportETA            *widget.Label
	exportLog            *widget.Entry
	ffmpegStatus         *widget.Label
	cancelRender         chan struct{}
	exportRunning        bool
	renderStartedAt      time.Time

	projectNameEntry *widget.Entry
	notesEntry       *widget.Entry
	tagsEntry        *widget.Entry
	projectPathLabel *widget.Label
	metadataLabel    *widget.Label
	recentProjects   *fyne.Container
	stylePreset      *widget.Select
	animPreset       *widget.Select
	exportPreset     *widget.Select

	staticPreview    *previewCard
	animationPreview *previewCard
	exportPreview    *previewCard
}

func newAppUI(app fyne.App) *appUI {
	ui := &appUI{
		app:     app,
		window:  app.NewWindow("Lola Rennt Animator"),
		project: normalizeProject(defaultProject()),
	}
	ui.ensureSeed()
	ui.window.Resize(fyne.NewSize(1520, 980))
	ui.status = widget.NewLabel("")
	ui.window.SetContent(ui.buildContent())
	ui.applyProjectToWidgets()
	ui.refreshDerivedUI()
	return ui
}

func (ui *appUI) buildContent() fyne.CanvasObject {
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), ui.newProject),
		widget.NewToolbarAction(theme.FolderOpenIcon(), ui.openProjectDialog),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), ui.saveProject),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), ui.refreshDerivedUI),
		widget.NewToolbarAction(theme.ContentCopyIcon(), ui.copyFFmpegCommand),
	)

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Текст і стиль", theme.DocumentCreateIcon(), ui.buildTextAndStyleTab()),
		container.NewTabItemWithIcon("Анімація / Preview", theme.MediaPlayIcon(), ui.buildAnimationTab()),
		container.NewTabItemWithIcon("Експорт", theme.DownloadIcon(), ui.buildExportTab()),
		container.NewTabItemWithIcon("Проєкт / Пресети", theme.SettingsIcon(), ui.buildProjectTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return container.NewBorder(toolbar, ui.status, nil, nil, tabs)
}

func (ui *appUI) buildTextAndStyleTab() fyne.CanvasObject {
	ui.languageChecks = widget.NewCheckGroup(languageOptions, func(options []string) {
		if ui.suspend {
			return
		}
		ui.project.Charset.Languages = sortedLanguages(options)
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.uppercaseCheck = widget.NewCheck("Uppercase only", func(value bool) {
		if ui.suspend {
			return
		}
		ui.project.Text.UppercaseOnly = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.autoReplaceCheck = widget.NewCheck("Автозаміна неподтримуваних символів", func(value bool) {
		if ui.suspend {
			return
		}
		ui.project.Text.AutoReplaceUnsupported = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})

	ui.displayType = widget.NewRadioGroup([]string{displayModeSegment, displayModeDotMatrix}, func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Display.Mode = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})

	ui.textInput = widget.NewMultiLineEntry()
	ui.textInput.Wrapping = fyne.TextWrapWord
	ui.textInput.OnChanged = func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Text.Content = value
		ui.touchProject()
		ui.refreshDerivedUI()
	}

	ui.colorModeSelect = widget.NewSelect([]string{colorModeSingle, colorModePerCharacter}, func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Style.ColorMode = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})

	ui.mainColorButton = widget.NewButton("", func() {
		ui.pickColor("Main color", ui.project.Style.MainColor, func(hex string) {
			ui.project.Style.MainColor = hex
			ui.touchProject()
			ui.refreshDerivedUI()
		})
	})
	ui.glowColorButton = widget.NewButton("", func() {
		ui.pickColor("Glow color", ui.project.Style.GlowColor, func(hex string) {
			ui.project.Style.GlowColor = hex
			ui.touchProject()
			ui.refreshDerivedUI()
		})
	})
	ui.inactiveColorBtn = widget.NewButton("", func() {
		ui.pickColor("Inactive color", ui.project.Style.InactiveColor, func(hex string) {
			ui.project.Style.InactiveColor = hex
			ui.touchProject()
			ui.refreshDerivedUI()
		})
	})

	ui.glowSlider = newSlider(0, 100, ui.bindFloat(func(value float64) {
		ui.project.Style.GlowIntensity = value
	}))
	ui.inactiveSlider = newSlider(0, 100, ui.bindFloat(func(value float64) {
		ui.project.Style.InactiveVisibility = value
	}))
	ui.cellScaleSlider = newSlider(0.6, 2.4, ui.bindFloat(func(value float64) {
		ui.project.Layout.CellScale = value
	}))
	ui.charSpacingSlider = newSlider(0, 40, ui.bindFloat(func(value float64) {
		ui.project.Layout.CharacterSpacing = value
	}))
	ui.lineSpacingSlider = newSlider(0, 50, ui.bindFloat(func(value float64) {
		ui.project.Layout.LineSpacing = value
	}))
	ui.paddingSlider = newSlider(0, 120, ui.bindFloat(func(value float64) {
		ui.project.Layout.Padding = value
	}))
	ui.alignmentSelect = widget.NewSelect([]string{alignmentLeft, alignmentCenter, alignmentRight}, func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Layout.Alignment = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})

	ui.statsLabel = widget.NewLabel("")
	ui.validationPreview = widget.NewLabel("")
	ui.validationPreview.Wrapping = fyne.TextWrapWord
	ui.validationGrid = widget.NewTextGrid()
	ui.validationGrid.Scroll = fyne.ScrollBoth
	ui.replacementGuide = widget.NewLabel("")
	ui.replacementGuide.Wrapping = fyne.TextWrapWord
	ui.styleSummary = widget.NewLabel("")
	ui.staticPreview = newPreviewCard("Static preview", ui.window)
	validationGridScroll := container.NewScroll(ui.validationGrid)
	validationGridScroll.SetMinSize(fyne.NewSize(320, 140))

	left := paneScroll(container.NewVBox(
		sectionCard("Charset / Languages", container.NewVBox(
			ui.languageChecks,
			ui.uppercaseCheck,
			ui.autoReplaceCheck,
			widget.NewButton("Supported charset", ui.showSupportedCharsetDialog),
		)),
		sectionCard("Display type", container.NewVBox(
			ui.displayType,
		)),
		sectionCard("Текст", container.NewVBox(
			ui.textInput,
			container.NewGridWithColumns(3,
				widget.NewButton("Очистити", func() {
					ui.textInput.SetText("")
				}),
				widget.NewButton("Вставити", func() {
					ui.textInput.SetText(ui.window.Clipboard().Content())
				}),
				widget.NewButton("Normalize", func() {
					stats := analyzeText(ui.project.Text.Content, ui.project.Charset.Languages, ui.project.Text.UppercaseOnly, true)
					ui.textInput.SetText(stats.DisplayText)
				}),
			),
			ui.statsLabel,
			ui.validationPreview,
			widget.NewLabel("Validation highlight (unsupported cells are marked after normalization)"),
			validationGridScroll,
			ui.replacementGuide,
		)),
		sectionCard("Колір", container.NewVBox(
			ui.colorModeSelect,
			container.NewGridWithColumns(3, ui.mainColorButton, ui.glowColorButton, ui.inactiveColorBtn),
			widget.NewLabel("Glow intensity"),
			ui.glowSlider,
			widget.NewLabel("Inactive segment visibility"),
			ui.inactiveSlider,
		)),
		sectionCard("Layout", container.NewVBox(
			widget.NewLabel("Cell scale"),
			ui.cellScaleSlider,
			widget.NewLabel("Character spacing"),
			ui.charSpacingSlider,
			widget.NewLabel("Line spacing"),
			ui.lineSpacingSlider,
			widget.NewLabel("Padding"),
			ui.paddingSlider,
			ui.alignmentSelect,
		)),
	))

	right := paneScroll(container.NewVBox(
		ui.staticPreview.object(),
		sectionCard("Current style", ui.styleSummary),
	))

	split := container.NewHSplit(left, right)
	split.Offset = 0.35
	return split
}

func (ui *appUI) buildAnimationTab() fyne.CanvasObject {
	ui.animationTypeSelect = widget.NewSelect([]string{
		"Scramble basic",
		"Scramble with lock",
		"Reveal then scramble then lock",
	}, func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Animation.Type = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.randomSourceSelect = widget.NewSelect([]string{
		"Digits only",
		"Letters only",
		"Alphanumeric",
		"Current charset only",
	}, func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Animation.RandomSource = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.allowInvalidCheck = widget.NewCheck("Дозволити невалідні символи", func(value bool) {
		if ui.suspend {
			return
		}
		ui.project.Animation.AllowInvalidRandomChars = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.allowEmptyCheck = widget.NewCheck("Дозволити порожню комірку", func(value bool) {
		if ui.suspend {
			return
		}
		ui.project.Animation.AllowEmptyCell = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.totalDurationSlider = newSlider(0.5, 12, ui.bindFloat(func(value float64) {
		ui.project.Animation.TotalDuration = value
	}))
	ui.introDelaySlider = newSlider(0, 4, ui.bindFloat(func(value float64) {
		ui.project.Animation.IntroDelay = value
	}))
	ui.outroHoldSlider = newSlider(0, 4, ui.bindFloat(func(value float64) {
		ui.project.Animation.OutroHold = value
	}))
	ui.perCharDelaySlider = newSlider(0, 0.5, ui.bindFloat(func(value float64) {
		ui.project.Animation.PerCharacterDelay = value
	}))
	ui.switchRateSlider = newSlider(1, 60, ui.bindFloat(func(value float64) {
		ui.project.Animation.RandomSwitchRate = value
	}))
	ui.seedEntry = widget.NewEntry()
	ui.seedEntry.OnChanged = func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Animation.Seed = ensureSeed(value)
		ui.suspend = true
		ui.seedEntry.SetText(ui.project.Animation.Seed)
		ui.suspend = false
		ui.touchProject()
		ui.refreshDerivedUI()
	}
	ui.lockOrderSelect = widget.NewSelect([]string{"Left-to-right", "Right-to-left", "Center-out", "Random", "By lines"}, func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Animation.LockOrder = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.lockModeSelect = widget.NewSelect([]string{"Hard lock", "Probabilistic lock"}, func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Animation.LockMode = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.loopCheck = widget.NewCheck("Loop", func(value bool) {
		if ui.suspend {
			return
		}
		ui.project.Animation.Loop = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.timelineSlider = widget.NewSlider(0, 100)
	ui.timelineSlider.Step = 0.1
	ui.timelineSlider.OnChanged = func(_ float64) {
		ui.refreshDerivedUI()
	}
	ui.playbackInfo = widget.NewLabel("")
	ui.animationSummary = widget.NewLabel("")
	ui.animationPreview = newPreviewCard("Animation preview", ui.window)

	left := paneScroll(container.NewVBox(
		sectionCard("Тип анімації", container.NewVBox(
			ui.animationTypeSelect,
			widget.NewLabel("Random source"),
			ui.randomSourceSelect,
			ui.allowInvalidCheck,
			ui.allowEmptyCheck,
		)),
		sectionCard("Timing", container.NewVBox(
			widget.NewLabel("Total duration"),
			ui.totalDurationSlider,
			widget.NewLabel("Intro delay"),
			ui.introDelaySlider,
			widget.NewLabel("Outro hold"),
			ui.outroHoldSlider,
			widget.NewLabel("Per-character delay"),
			ui.perCharDelaySlider,
			widget.NewLabel("Random switch rate"),
			ui.switchRateSlider,
			widget.NewLabel("Seed"),
			ui.seedEntry,
			widget.NewButton("Generate random seed", func() {
				ui.project.Animation.Seed = ensureSeed("")
				ui.touchProject()
				ui.applyProjectToWidgets()
				ui.refreshDerivedUI()
			}),
		)),
		sectionCard("Reveal / Lock logic", container.NewVBox(
			widget.NewLabel("Order"),
			ui.lockOrderSelect,
			widget.NewLabel("Lock mode"),
			ui.lockModeSelect,
			ui.loopCheck,
		)),
	))

	right := paneScroll(container.NewVBox(
		ui.animationPreview.object(),
		sectionCard("Playback", container.NewVBox(
			container.NewGridWithColumns(5,
				widget.NewButtonWithIcon("Play", theme.MediaPlayIcon(), func() {
					ui.startPlayback(false)
				}),
				widget.NewButtonWithIcon("Pause", theme.MediaPauseIcon(), func() {
					ui.pausePlayback()
				}),
				widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {
					ui.stopPlayback(true)
				}),
				widget.NewButtonWithIcon("Restart", theme.ViewRefreshIcon(), func() {
					ui.restartPlayback()
				}),
				widget.NewButtonWithIcon("Loop", theme.ViewRefreshIcon(), func() {
					ui.loopCheck.SetChecked(!ui.loopCheck.Checked)
				}),
			),
			ui.timelineSlider,
			ui.playbackInfo,
		)),
		sectionCard("Current animation", ui.animationSummary),
	))

	split := container.NewHSplit(left, right)
	split.Offset = 0.35
	return split
}

func (ui *appUI) buildExportTab() fyne.CanvasObject {
	ui.backgroundModeSelect = widget.NewSelect([]string{
		backgroundModeTransparent,
		backgroundModeSolid,
		backgroundModeGradient,
		backgroundModeImage,
		backgroundModeVideo,
	}, func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Background.Mode = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.bgColorButton = widget.NewButton("", func() {
		ui.pickColor("Background color", ui.project.Background.SolidColor, func(hex string) {
			ui.project.Background.SolidColor = hex
			ui.touchProject()
			ui.refreshDerivedUI()
		})
	})
	ui.gradientAButton = widget.NewButton("", func() {
		ui.pickColor("Gradient A", ui.project.Background.GradientA, func(hex string) {
			ui.project.Background.GradientA = hex
			ui.touchProject()
			ui.refreshDerivedUI()
		})
	})
	ui.gradientBButton = widget.NewButton("", func() {
		ui.pickColor("Gradient B", ui.project.Background.GradientB, func(hex string) {
			ui.project.Background.GradientB = hex
			ui.touchProject()
			ui.refreshDerivedUI()
		})
	})
	ui.imagePathEntry = widget.NewEntry()
	ui.imagePathEntry.OnChanged = func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Background.ImagePath = value
		ui.touchProject()
		ui.refreshDerivedUI()
	}
	ui.bgImageOpacitySlider = newSlider(1, 100, ui.bindFloat(func(value float64) {
		ui.project.Background.ImageOpacity = value
	}))
	ui.videoPathEntry = widget.NewEntry()
	ui.videoPathEntry.OnChanged = func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Background.VideoPath = value
		ui.touchProject()
		ui.refreshDerivedUI()
	}
	ui.bgFitSelect = widget.NewSelect([]string{"Fit", "Fill", "Center", "Stretch"}, func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Background.FitMode = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})

	ui.widthEntry = newIntEntry(ui.bindInt(func(value int) { ui.project.Export.Width = value }))
	ui.heightEntry = newIntEntry(ui.bindInt(func(value int) { ui.project.Export.Height = value }))
	ui.fpsEntry = newIntEntry(ui.bindInt(func(value int) { ui.project.Export.FPS = value }))
	ui.startFrameEntry = newIntEntry(ui.bindInt(func(value int) { ui.project.Export.StartFrame = value }))
	ui.endFrameEntry = newIntEntry(ui.bindInt(func(value int) { ui.project.Export.EndFrame = value }))
	ui.outputFolderEntry = widget.NewEntry()
	ui.outputFolderEntry.OnChanged = func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Export.OutputFolder = value
		ui.touchProject()
		ui.refreshDerivedUI()
	}
	ui.filePrefixEntry = widget.NewEntry()
	ui.filePrefixEntry.OnChanged = func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Export.FilePrefix = value
		ui.touchProject()
		ui.refreshDerivedUI()
	}
	ui.overwriteSelect = widget.NewSelect([]string{"Ask", "Overwrite", "Create new suffix"}, func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Export.OverwritePolicy = value
		ui.touchProject()
		ui.refreshDerivedUI()
	})
	ui.supersamplingSlider = newSlider(1, 4, ui.bindFloat(func(value float64) {
		ui.project.Export.Supersampling = value
	}))
	ui.ffmpegCommand = widget.NewMultiLineEntry()
	ui.ffmpegCommand.Disable()
	ui.exportLog = widget.NewMultiLineEntry()
	ui.exportLog.Disable()
	ui.exportSummary = widget.NewLabel("")
	ui.exportProgress = widget.NewProgressBar()
	ui.exportETA = widget.NewLabel("ETA: --")
	ui.ffmpegStatus = widget.NewLabel("")
	ui.exportPreview = newPreviewCard("Export preview", ui.window)

	left := paneScroll(container.NewVBox(
		sectionCard("Output", container.NewVBox(
			widget.NewLabel("Output folder"),
			ui.outputFolderEntry,
			container.NewGridWithColumns(2,
				widget.NewButtonWithIcon("Вибрати теку", theme.FolderOpenIcon(), ui.pickOutputFolder),
				widget.NewButton("Open", ui.openOutputFolderAction),
			),
			widget.NewLabel("File prefix"),
			ui.filePrefixEntry,
			widget.NewLabel("Overwrite policy"),
			ui.overwriteSelect,
		)),
		sectionCard("Render size", container.NewVBox(
			container.NewGridWithColumns(2, ui.widthEntry, ui.heightEntry),
			container.NewGridWithColumns(2,
				widget.NewButton("1920x1080", func() { ui.setRenderSize(1920, 1080) }),
				widget.NewButton("1080x1920", func() { ui.setRenderSize(1080, 1920) }),
			),
			widget.NewLabel("Supersampling"),
			ui.supersamplingSlider,
		)),
		sectionCard("Frame settings", container.NewVBox(
			widget.NewLabel("FPS"),
			ui.fpsEntry,
			widget.NewLabel("Start frame"),
			ui.startFrameEntry,
			widget.NewLabel("End frame"),
			ui.endFrameEntry,
		)),
		sectionCard("Background", container.NewVBox(
			ui.backgroundModeSelect,
			container.NewGridWithColumns(3, ui.bgColorButton, ui.gradientAButton, ui.gradientBButton),
			widget.NewLabel("Image path"),
			ui.imagePathEntry,
			widget.NewButtonWithIcon("Вибрати зображення", theme.FolderOpenIcon(), ui.pickBackgroundImage),
			widget.NewLabel("Image opacity"),
			ui.bgImageOpacitySlider,
			widget.NewLabel("Video path"),
			ui.videoPathEntry,
			widget.NewButtonWithIcon("Вибрати відео", theme.MediaVideoIcon(), ui.pickBackgroundVideo),
			widget.NewLabel("Fit mode"),
			ui.bgFitSelect,
		)),
	))

	right := paneScroll(container.NewVBox(
		ui.exportPreview.object(),
		sectionCard("FFmpeg tools", ui.ffmpegStatus),
		sectionCard("FFmpeg command", ui.ffmpegCommand),
		sectionCard("Export summary", ui.exportSummary),
		sectionCard("Render control", container.NewVBox(
			ui.exportProgress,
			ui.exportETA,
			container.NewGridWithColumns(2,
				widget.NewButtonWithIcon("Render PNG sequence", theme.DocumentIcon(), ui.renderPNGSequenceAction),
				widget.NewButtonWithIcon("Copy ffmpeg command", theme.ContentCopyIcon(), ui.copyFFmpegCommand),
			),
			container.NewGridWithColumns(2,
				widget.NewButtonWithIcon("Cancel render", theme.CancelIcon(), ui.cancelRenderAction),
				widget.NewButtonWithIcon("Refresh tools", theme.ViewRefreshIcon(), ui.refreshDerivedUI),
			),
			ui.exportLog,
		)),
	))

	split := container.NewHSplit(left, right)
	split.Offset = 0.35
	return split
}

func (ui *appUI) buildProjectTab() fyne.CanvasObject {
	ui.projectNameEntry = widget.NewEntry()
	ui.projectNameEntry.OnChanged = func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Metadata.ProjectName = value
		ui.touchProject()
		ui.refreshDerivedUI()
	}
	ui.notesEntry = widget.NewMultiLineEntry()
	ui.notesEntry.OnChanged = func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Metadata.Notes = value
		ui.touchProject()
		ui.refreshDerivedUI()
	}
	ui.tagsEntry = widget.NewEntry()
	ui.tagsEntry.OnChanged = func(value string) {
		if ui.suspend {
			return
		}
		ui.project.Metadata.Tags = value
		ui.touchProject()
		ui.refreshDerivedUI()
	}
	ui.projectPathLabel = widget.NewLabel("")
	ui.metadataLabel = widget.NewLabel("")
	ui.recentProjects = container.NewVBox(widget.NewLabel("No recent projects yet"))
	ui.stylePreset = widget.NewSelect(presetNames(stylePresets), nil)
	ui.animPreset = widget.NewSelect(presetNames(animationPresets), nil)
	ui.exportPreset = widget.NewSelect(presetNames(exportPresets), nil)
	ui.stylePreset.SetSelected(stylePresets[0].Name)
	ui.animPreset.SetSelected(animationPresets[0].Name)
	ui.exportPreset.SetSelected(exportPresets[0].Name)

	left := paneScroll(container.NewVBox(
		sectionCard("Project", container.NewGridWithColumns(2,
			widget.NewButtonWithIcon("New", theme.DocumentCreateIcon(), ui.newProject),
			widget.NewButtonWithIcon("Open", theme.FolderOpenIcon(), ui.openProjectDialog),
			widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), ui.saveProject),
			widget.NewButtonWithIcon("Save As", theme.DocumentSaveIcon(), ui.saveProjectAsDialog),
		)),
		sectionCard("Metadata", container.NewVBox(
			widget.NewLabel("Project name"),
			ui.projectNameEntry,
			widget.NewLabel("Notes"),
			ui.notesEntry,
			widget.NewLabel("Tags"),
			ui.tagsEntry,
			ui.metadataLabel,
		)),
	))

	right := paneScroll(container.NewVBox(
		sectionCard("Current file", ui.projectPathLabel),
		sectionCard("Recent projects", ui.recentProjects),
		sectionCard("Presets", container.NewVBox(
			widget.NewLabel("Style preset"),
			ui.stylePreset,
			widget.NewButton("Apply style preset", func() {
				if applyPresetByName(&ui.project, ui.stylePreset.Selected, stylePresets) {
					ui.touchProject()
					ui.applyProjectToWidgets()
					ui.refreshDerivedUI()
				}
			}),
			widget.NewLabel("Animation preset"),
			ui.animPreset,
			widget.NewButton("Apply animation preset", func() {
				if applyPresetByName(&ui.project, ui.animPreset.Selected, animationPresets) {
					ui.touchProject()
					ui.applyProjectToWidgets()
					ui.refreshDerivedUI()
				}
			}),
			widget.NewLabel("Export preset"),
			ui.exportPreset,
			widget.NewButton("Apply export preset", func() {
				if applyPresetByName(&ui.project, ui.exportPreset.Selected, exportPresets) {
					ui.touchProject()
					ui.applyProjectToWidgets()
					ui.refreshDerivedUI()
				}
			}),
		)),
	))

	split := container.NewHSplit(left, right)
	split.Offset = 0.45
	return split
}

func sectionCard(title string, content fyne.CanvasObject) fyne.CanvasObject {
	return widget.NewCard(title, "", content)
}

func paneScroll(content fyne.CanvasObject) *container.Scroll {
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(180, 180))
	return scroll
}

func newSlider(min, max float64, onChange func(float64)) *widget.Slider {
	slider := widget.NewSlider(min, max)
	slider.Step = (max - min) / 100
	slider.OnChanged = func(value float64) {
		onChange(value)
	}
	return slider
}

func newIntEntry(onChange func(int)) *widget.Entry {
	entry := widget.NewEntry()
	entry.OnChanged = func(value string) {
		if value == "" {
			return
		}
		parsed, err := strconv.Atoi(value)
		if err == nil {
			onChange(parsed)
		}
	}
	return entry
}

func (ui *appUI) bindFloat(assign func(float64)) func(float64) {
	return func(value float64) {
		if ui.suspend {
			return
		}
		assign(value)
		ui.touchProject()
		ui.refreshDerivedUI()
	}
}

func (ui *appUI) bindInt(assign func(int)) func(int) {
	return func(value int) {
		if ui.suspend {
			return
		}
		assign(value)
		ui.touchProject()
		ui.refreshDerivedUI()
	}
}

func (ui *appUI) applyProjectToWidgets() {
	ui.suspend = true
	defer func() { ui.suspend = false }()

	ui.textInput.SetText(ui.project.Text.Content)
	ui.languageChecks.SetSelected(ui.project.Charset.Languages)
	ui.uppercaseCheck.SetChecked(ui.project.Text.UppercaseOnly)
	ui.autoReplaceCheck.SetChecked(ui.project.Text.AutoReplaceUnsupported)
	ui.displayType.SetSelected(ui.project.Display.Mode)
	ui.colorModeSelect.SetSelected(ui.project.Style.ColorMode)
	ui.glowSlider.SetValue(ui.project.Style.GlowIntensity)
	ui.inactiveSlider.SetValue(ui.project.Style.InactiveVisibility)
	ui.cellScaleSlider.SetValue(ui.project.Layout.CellScale)
	ui.charSpacingSlider.SetValue(ui.project.Layout.CharacterSpacing)
	ui.lineSpacingSlider.SetValue(ui.project.Layout.LineSpacing)
	ui.paddingSlider.SetValue(ui.project.Layout.Padding)
	ui.alignmentSelect.SetSelected(ui.project.Layout.Alignment)

	ui.animationTypeSelect.SetSelected(ui.project.Animation.Type)
	ui.randomSourceSelect.SetSelected(ui.project.Animation.RandomSource)
	ui.allowInvalidCheck.SetChecked(ui.project.Animation.AllowInvalidRandomChars)
	ui.allowEmptyCheck.SetChecked(ui.project.Animation.AllowEmptyCell)
	ui.totalDurationSlider.SetValue(ui.project.Animation.TotalDuration)
	ui.introDelaySlider.SetValue(ui.project.Animation.IntroDelay)
	ui.outroHoldSlider.SetValue(ui.project.Animation.OutroHold)
	ui.perCharDelaySlider.SetValue(ui.project.Animation.PerCharacterDelay)
	ui.switchRateSlider.SetValue(ui.project.Animation.RandomSwitchRate)
	ui.seedEntry.SetText(ui.project.Animation.Seed)
	ui.lockOrderSelect.SetSelected(ui.project.Animation.LockOrder)
	ui.lockModeSelect.SetSelected(ui.project.Animation.LockMode)
	ui.loopCheck.SetChecked(ui.project.Animation.Loop)

	ui.backgroundModeSelect.SetSelected(ui.project.Background.Mode)
	ui.imagePathEntry.SetText(ui.project.Background.ImagePath)
	ui.bgImageOpacitySlider.SetValue(ui.project.Background.ImageOpacity)
	ui.videoPathEntry.SetText(ui.project.Background.VideoPath)
	ui.bgFitSelect.SetSelected(ui.project.Background.FitMode)
	ui.widthEntry.SetText(strconv.Itoa(ui.project.Export.Width))
	ui.heightEntry.SetText(strconv.Itoa(ui.project.Export.Height))
	ui.fpsEntry.SetText(strconv.Itoa(ui.project.Export.FPS))
	ui.startFrameEntry.SetText(strconv.Itoa(ui.project.Export.StartFrame))
	ui.endFrameEntry.SetText(strconv.Itoa(ui.project.Export.EndFrame))
	ui.outputFolderEntry.SetText(ui.project.Export.OutputFolder)
	ui.filePrefixEntry.SetText(ui.project.Export.FilePrefix)
	ui.overwriteSelect.SetSelected(ui.project.Export.OverwritePolicy)
	ui.supersamplingSlider.SetValue(ui.project.Export.Supersampling)

	ui.projectNameEntry.SetText(ui.project.Metadata.ProjectName)
	ui.notesEntry.SetText(ui.project.Metadata.Notes)
	ui.tagsEntry.SetText(ui.project.Metadata.Tags)
}

func (ui *appUI) refreshDerivedUI() {
	if ui.window == nil {
		return
	}

	stats := analyzeText(ui.project.Text.Content, ui.project.Charset.Languages, ui.project.Text.UppercaseOnly, ui.project.Text.AutoReplaceUnsupported)
	ui.staticPreview.applyProject(ui.project, stats, ui.project.Export.StartFrame)
	previewFrame := ui.currentPreviewFrame()
	ui.animationPreview.applyProject(ui.project, stats, previewFrame)
	ui.exportPreview.applyProject(ui.project, stats, ui.project.Export.EndFrame)

	ui.statsLabel.SetText(fmt.Sprintf("Символів: %d • Рядків: %d • Unsupported: %s", stats.CharacterCount, stats.LineCount, formatUnsupportedRunes(stats.UnsupportedUnique)))
	ui.validationPreview.SetText("Validation preview:\n" + buildValidationPreview(ui.project.Text.Content, ui.project.Charset.Languages, ui.project.Text.UppercaseOnly))
	ui.refreshValidationGrid()
	ui.replacementGuide.SetText(formatReplacementGuidance(stats.UnsupportedUnique, ui.project.Text.AutoReplaceUnsupported))
	ui.styleSummary.SetText(fmt.Sprintf(
		"Mode: %s\nLanguages: %s\nMain: %s\nGlow: %.0f%%\nInactive: %.0f%%\nAlignment: %s",
		ui.project.Display.Mode,
		strings.Join(sortedLanguages(ui.project.Charset.Languages), ", "),
		ui.project.Style.MainColor,
		ui.project.Style.GlowIntensity,
		ui.project.Style.InactiveVisibility,
		ui.project.Layout.Alignment,
	))

	ui.playbackInfo.SetText(ui.playbackSummary())
	ui.animationSummary.SetText(fmt.Sprintf(
		"Type: %s\nRandom source: %s\nSeed: %s\nLock: %s / %s",
		ui.project.Animation.Type,
		ui.project.Animation.RandomSource,
		ui.project.Animation.Seed,
		ui.project.Animation.LockOrder,
		ui.project.Animation.LockMode,
	))

	ui.exportSummary.SetText(fmt.Sprintf(
		"Canvas: %dx%d\nFrames: %d-%d\nFPS: %d\nBackground: %s\nFit: %s",
		ui.project.Export.Width,
		ui.project.Export.Height,
		ui.project.Export.StartFrame,
		ui.project.Export.EndFrame,
		ui.project.Export.FPS,
		ui.project.Background.Mode,
		ui.project.Background.FitMode,
	))
	ui.ffmpegCommand.SetText(ui.buildFFmpegCommand())
	ui.ffmpegStatus.SetText(ui.ffmpegStatusText())

	ui.setColorButtonLabel(ui.mainColorButton, "Main", ui.project.Style.MainColor)
	ui.setColorButtonLabel(ui.glowColorButton, "Glow", ui.project.Style.GlowColor)
	ui.setColorButtonLabel(ui.inactiveColorBtn, "Inactive", ui.project.Style.InactiveColor)
	ui.setColorButtonLabel(ui.bgColorButton, "BG", ui.project.Background.SolidColor)
	ui.setColorButtonLabel(ui.gradientAButton, "Grad A", ui.project.Background.GradientA)
	ui.setColorButtonLabel(ui.gradientBButton, "Grad B", ui.project.Background.GradientB)

	projectPath := "Unsaved project"
	if ui.projectPath != "" {
		projectPath = ui.projectPath
	}
	ui.projectPathLabel.SetText(projectPath)
	ui.metadataLabel.SetText(fmt.Sprintf("Created: %s\nUpdated: %s", ui.project.Metadata.CreatedAt, ui.project.Metadata.UpdatedAt))
	ui.refreshRecentProjects()
	if ui.playbackRunning {
		ui.setStatus(fmt.Sprintf("Playback running • Frame %d / %d", ui.currentPreviewFrame(), ui.playbackEndFrame()))
	} else {
		ui.setStatus(fmt.Sprintf("Project ready • %s • Seed %s", ui.project.Metadata.ProjectName, ui.project.Animation.Seed))
	}
}

func (ui *appUI) playbackSummary() string {
	fps := maxInt(ui.project.Export.FPS, 1)
	currentFrame := ui.currentPreviewFrame()
	currentTime := float64(currentFrame-ui.project.Export.StartFrame) / float64(fps)
	state := "Paused"
	if ui.playbackRunning {
		state = "Playing"
	}
	return fmt.Sprintf("%s • Current time: %.2f s • Frame: %d / %d • Preview FPS: %d", state, currentTime, currentFrame, ui.playbackEndFrame(), fps)
}

func (ui *appUI) currentPreviewFrame() int {
	startFrame := ui.project.Export.StartFrame
	endFrame := ui.playbackEndFrame()
	totalFrames := endFrame - startFrame + 1
	if totalFrames <= 1 {
		return startFrame
	}

	progress := ui.timelineSlider.Value / 100.0
	frameOffset := int(progress*float64(totalFrames-1) + 0.5)
	frame := startFrame + frameOffset
	if frame > endFrame {
		return endFrame
	}
	return frame
}

func (ui *appUI) playbackEndFrame() int {
	if ui.project.Export.EndFrame < ui.project.Export.StartFrame {
		return ui.project.Export.StartFrame
	}
	return ui.project.Export.EndFrame
}

func (ui *appUI) setPreviewFrame(frame int) {
	startFrame := ui.project.Export.StartFrame
	endFrame := ui.playbackEndFrame()
	if frame < startFrame {
		frame = startFrame
	}
	if frame > endFrame {
		frame = endFrame
	}

	totalFrames := endFrame - startFrame + 1
	if totalFrames <= 1 {
		ui.timelineSlider.SetValue(0)
		return
	}

	progress := float64(frame-startFrame) / float64(totalFrames-1)
	ui.timelineSlider.SetValue(progress * 100.0)
}

func (ui *appUI) startPlayback(restart bool) {
	if restart {
		ui.setPreviewFrame(ui.project.Export.StartFrame)
	}
	if ui.playbackRunning {
		return
	}
	if ui.currentPreviewFrame() >= ui.playbackEndFrame() {
		ui.setPreviewFrame(ui.project.Export.StartFrame)
	}

	fps := maxInt(ui.project.Export.FPS, 1)
	interval := time.Second / time.Duration(fps)
	stopCh := make(chan struct{})
	ui.playbackStop = stopCh
	ui.playbackRunning = true
	ui.refreshDerivedUI()

	go func(stopCh chan struct{}) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				fyne.Do(func() {
					if ui.playbackStop != stopCh {
						return
					}
					ui.advancePlaybackFrame()
				})
			}
		}
	}(stopCh)
}

func (ui *appUI) advancePlaybackFrame() {
	nextFrame := ui.currentPreviewFrame() + 1
	if nextFrame > ui.playbackEndFrame() {
		if ui.project.Animation.Loop {
			ui.setPreviewFrame(ui.project.Export.StartFrame)
			return
		}
		ui.stopPlaybackLoop()
		ui.setPreviewFrame(ui.playbackEndFrame())
		ui.refreshDerivedUI()
		ui.setStatus("Playback finished.")
		return
	}
	ui.setPreviewFrame(nextFrame)
}

func (ui *appUI) pausePlayback() {
	if !ui.playbackRunning {
		ui.setStatus("Playback is not running.")
		return
	}
	ui.stopPlaybackLoop()
	ui.refreshDerivedUI()
	ui.setStatus("Playback paused.")
}

func (ui *appUI) stopPlayback(resetToStart bool) {
	wasRunning := ui.playbackRunning
	ui.stopPlaybackLoop()
	if resetToStart {
		ui.setPreviewFrame(ui.project.Export.StartFrame)
	}
	ui.refreshDerivedUI()
	if wasRunning {
		ui.setStatus("Playback stopped.")
	} else {
		ui.setStatus("Preview reset to the first frame.")
	}
}

func (ui *appUI) restartPlayback() {
	ui.stopPlaybackLoop()
	ui.setPreviewFrame(ui.project.Export.StartFrame)
	ui.startPlayback(true)
}

func (ui *appUI) stopPlaybackLoop() {
	if ui.playbackStop != nil {
		close(ui.playbackStop)
		ui.playbackStop = nil
	}
	ui.playbackRunning = false
}

func (ui *appUI) setColorButtonLabel(button *widget.Button, title, hex string) {
	if button == nil {
		return
	}
	button.SetText(fmt.Sprintf("%s %s", title, hex))
}

func (ui *appUI) pickColor(title, current string, apply func(string)) {
	picker := dialog.NewColorPicker(title, "Select a color", func(c color.Color) {
		apply(colorToHex(c))
	}, ui.window)
	picker.Advanced = true
	picker.Show()
	picker.SetColor(parseHexColor(current, color.NRGBA{R: 255, G: 96, B: 64, A: 255}))
}

func (ui *appUI) showSupportedCharsetDialog() {
	viewer := widget.NewMultiLineEntry()
	viewer.SetText(buildSupportedCharsetSummary(ui.project.Charset.Languages))
	viewer.Wrapping = fyne.TextWrapWord
	viewer.Disable()
	viewer.SetMinRowsVisible(12)

	dialog.NewCustom(
		"Supported charset",
		"Close",
		container.NewScroll(viewer),
		ui.window,
	).Show()
}

func (ui *appUI) refreshValidationGrid() {
	if ui.validationGrid == nil {
		return
	}

	normalized := normalizeValidationText(ui.project.Text.Content, ui.project.Text.UppercaseOnly)
	ui.validationGrid.SetText(normalized)

	unsupportedStyle := &widget.CustomTextGridStyle{
		TextStyle: fyne.TextStyle{Bold: true},
		FGColor:   color.NRGBA{R: 255, G: 245, B: 245, A: 255},
		BGColor:   color.NRGBA{R: 158, G: 36, B: 36, A: 255},
	}
	supported := supportedRunes(ui.project.Charset.Languages)

	row := 0
	col := 0
	for _, r := range normalized {
		if r == '\n' {
			row++
			col = 0
			continue
		}
		if !supported[r] {
			ui.validationGrid.SetStyle(row, col, unsupportedStyle)
		}
		col++
	}
	ui.validationGrid.Refresh()
}

func (ui *appUI) touchProject() {
	ui.project.Metadata.UpdatedAt = timeNowString()
}

func (ui *appUI) ensureSeed() {
	ui.project.Animation.Seed = ensureSeed(ui.project.Animation.Seed)
}

func (ui *appUI) setStatus(message string) {
	if ui.status != nil {
		ui.status.SetText(message)
	}
}

func (ui *appUI) newProject() {
	ui.stopPlaybackLoop()
	ui.project = normalizeProject(defaultProject())
	ui.projectPath = ""
	ui.ensureSeed()
	ui.applyProjectToWidgets()
	ui.refreshDerivedUI()
}

func (ui *appUI) openProjectDialog() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		raw, err := io.ReadAll(reader)
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		var loaded Project
		if err := json.Unmarshal(raw, &loaded); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		ui.stopPlaybackLoop()
		ui.project = normalizeProject(loaded)
		ui.ensureSeed()
		if reader.URI() != nil {
			ui.projectPath = reader.URI().Path()
			ui.recordRecentProject(ui.projectPath)
		}
		ui.applyProjectToWidgets()
		ui.refreshDerivedUI()
	}, ui.window)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	fd.Show()
}

func (ui *appUI) saveProject() {
	ui.ensureSeed()
	if ui.projectPath == "" {
		ui.saveProjectAsDialog()
		return
	}
	if err := ui.saveToPath(ui.projectPath); err != nil {
		dialog.ShowError(err, ui.window)
	}
}

func (ui *appUI) saveProjectAsDialog() {
	ui.ensureSeed()
	fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		ui.touchProject()
		raw, err := ui.project.Marshal()
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if _, err := writer.Write(raw); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if writer.URI() != nil {
			ui.projectPath = writer.URI().Path()
			ui.recordRecentProject(ui.projectPath)
		}
		ui.refreshDerivedUI()
	}, ui.window)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	fd.SetFileName(safeFileName(ui.project.Metadata.ProjectName) + ".json")
	fd.Show()
}

func (ui *appUI) saveToPath(path string) error {
	ui.touchProject()
	raw, err := ui.project.Marshal()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return err
	}
	ui.projectPath = path
	ui.recordRecentProject(path)
	ui.refreshDerivedUI()
	return nil
}

func (ui *appUI) buildFFmpegCommand() string {
	tools := detectFFmpegTools()
	fps := maxInt(ui.project.Export.FPS, 1)
	outputFolder := strings.TrimSpace(ui.project.Export.OutputFolder)
	if outputFolder == "" {
		outputFolder = "."
	}
	prefix := strings.TrimSpace(ui.project.Export.FilePrefix)
	if prefix == "" {
		prefix = "frame"
	}

	framePattern := filepath.ToSlash(filepath.Join(outputFolder, prefix+"_%05d.png"))
	outputPath := filepath.ToSlash(filepath.Join(outputFolder, prefix+".mp4"))
	ffmpegBin := "ffmpeg"
	if tools.FFmpegPath != "" {
		ffmpegBin = tools.FFmpegPath
	}
	return fmt.Sprintf(
		"\"%s\" -framerate %d -i \"%s\" -c:v libx264 -pix_fmt yuv420p \"%s\"",
		ffmpegBin,
		fps,
		framePattern,
		outputPath,
	)
}

func (ui *appUI) copyFFmpegCommand() {
	ui.window.Clipboard().SetContent(ui.buildFFmpegCommand())
	ui.setStatus("FFmpeg command copied to clipboard.")
}

func (ui *appUI) renderPNGSequenceAction() {
	if ui.exportRunning {
		ui.setStatus("Render is already running.")
		return
	}
	prepared, err := prepareRenderProject(ui.project, false)
	if err != nil {
		if conflict, ok := err.(overwriteConflictError); ok {
			dialog.NewConfirm(
				"Overwrite existing frames?",
				fmt.Sprintf("Existing output detected:\n%s\n\nOverwrite existing PNG frames?", conflict.Path),
				func(confirm bool) {
					if !confirm {
						ui.setStatus("Render cancelled before start.")
						return
					}
					project, prepErr := prepareRenderProject(ui.project, true)
					if prepErr != nil {
						dialog.ShowError(prepErr, ui.window)
						return
					}
					ui.startRender(project)
				},
				ui.window,
			).Show()
			return
		}
		dialog.ShowError(err, ui.window)
		return
	}
	ui.startRender(prepared)
}

func (ui *appUI) startRender(project Project) {
	ui.exportRunning = true
	ui.cancelRender = make(chan struct{})
	ui.renderStartedAt = time.Now()
	ui.exportProgress.SetValue(0)
	ui.exportETA.SetText("ETA: calculating...")
	ui.exportLog.SetText("")
	ui.setStatus("PNG render started.")
	if project.Export.FilePrefix != ui.project.Export.FilePrefix {
		ui.appendExportLog(fmt.Sprintf("Using output prefix: %s", project.Export.FilePrefix))
	}

	go func() {
		outputFolder, count, err := renderPNGSequence(project, ui.cancelRender, func(step renderProgress) {
			fyne.Do(func() {
				if step.TotalFrames > 0 {
					ui.exportProgress.SetValue(float64(step.CurrentFrame) / float64(step.TotalFrames))
				}
				ui.updateRenderETA(step.CurrentFrame, step.TotalFrames)
				ui.appendExportLog(fmt.Sprintf("[%d/%d] %s", step.CurrentFrame, step.TotalFrames, filepathBase(step.Filename)))
			})
		})

		fyne.Do(func() {
			ui.exportRunning = false
			ui.cancelRender = nil
			if err != nil {
				if err == errRenderCancelled {
					ui.exportETA.SetText("ETA: cancelled")
					ui.appendExportLog("Render cancelled by user.")
					ui.setStatus("Render cancelled.")
					return
				}
				ui.exportETA.SetText("ETA: failed")
				ui.appendExportLog("Render failed: " + err.Error())
				dialog.ShowError(err, ui.window)
				return
			}
			ui.exportProgress.SetValue(1)
			ui.exportETA.SetText("ETA: complete")
			ui.appendExportLog(fmt.Sprintf("Render complete: %d frames in %s", count, outputFolder))
			ui.setStatus(fmt.Sprintf("Rendered %d PNG frames to %s", count, outputFolder))
		})
	}()
}

func (ui *appUI) cancelRenderAction() {
	if !ui.exportRunning || ui.cancelRender == nil {
		ui.setStatus("No active render to cancel.")
		return
	}
	close(ui.cancelRender)
	ui.cancelRender = nil
}

func (ui *appUI) updateRenderETA(currentFrame, totalFrames int) {
	if ui.exportETA == nil {
		return
	}
	if currentFrame <= 0 || totalFrames <= 0 || ui.renderStartedAt.IsZero() {
		ui.exportETA.SetText("ETA: calculating...")
		return
	}

	elapsed := time.Since(ui.renderStartedAt)
	if currentFrame >= totalFrames {
		ui.exportETA.SetText("ETA: finishing...")
		return
	}

	avgPerFrame := elapsed / time.Duration(currentFrame)
	remainingFrames := totalFrames - currentFrame
	remaining := avgPerFrame * time.Duration(remainingFrames)
	ui.exportETA.SetText(fmt.Sprintf("ETA: %s • Elapsed: %s", formatShortDuration(remaining), formatShortDuration(elapsed)))
}

func formatShortDuration(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	totalSeconds := int(d.Round(time.Second).Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %02ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func (ui *appUI) pickOutputFolder() {
	dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if list == nil {
			return
		}
		ui.project.Export.OutputFolder = list.Path()
		ui.touchProject()
		ui.applyProjectToWidgets()
		ui.refreshDerivedUI()
	}, ui.window)
}

func (ui *appUI) openOutputFolderAction() {
	outputFolder := strings.TrimSpace(ui.project.Export.OutputFolder)
	if outputFolder == "" {
		outputFolder = "."
	}

	absPath, err := filepath.Abs(outputFolder)
	if err != nil {
		dialog.ShowError(err, ui.window)
		return
	}
	info, err := os.Stat(absPath)
	if err != nil {
		dialog.ShowError(err, ui.window)
		return
	}
	if !info.IsDir() {
		dialog.ShowError(fmt.Errorf("output path is not a directory: %s", absPath), ui.window)
		return
	}

	if err := openPathInFileManager(absPath); err != nil {
		dialog.ShowError(err, ui.window)
		return
	}
	ui.setStatus("Opened output folder.")
}

func openPathInFileManager(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", filepath.Clean(path))
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func (ui *appUI) pickBackgroundImage() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		if reader.URI() != nil {
			ui.project.Background.ImagePath = reader.URI().Path()
			ui.project.Background.Mode = backgroundModeImage
			ui.touchProject()
			ui.applyProjectToWidgets()
			ui.refreshDerivedUI()
		}
	}, ui.window)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".webp"}))
	fd.Show()
}

func (ui *appUI) pickBackgroundVideo() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		if reader.URI() != nil {
			ui.project.Background.VideoPath = reader.URI().Path()
			ui.project.Background.Mode = backgroundModeVideo
			ui.touchProject()
			ui.applyProjectToWidgets()
			ui.refreshDerivedUI()
		}
	}, ui.window)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".mp4", ".mov", ".mkv", ".webm"}))
	fd.Show()
}

func (ui *appUI) setRenderSize(width, height int) {
	ui.project.Export.Width = width
	ui.project.Export.Height = height
	ui.touchProject()
	ui.applyProjectToWidgets()
	ui.refreshDerivedUI()
}

func safeFileName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "project"
	}
	replacer := strings.NewReplacer(" ", "_", "/", "_", "\\", "_", ":", "_")
	return replacer.Replace(trimmed)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (ui *appUI) appendExportLog(line string) {
	current := strings.TrimSpace(ui.exportLog.Text)
	if current == "" {
		ui.exportLog.SetText(line)
		return
	}
	ui.exportLog.SetText(current + "\n" + line)
}

func (ui *appUI) ffmpegStatusText() string {
	tools := detectFFmpegTools()
	ffmpeg := "not found"
	ffprobe := "not found"
	if tools.FFmpegPath != "" {
		ffmpeg = tools.FFmpegPath
	}
	if tools.FFprobePath != "" {
		ffprobe = tools.FFprobePath
	}
	return fmt.Sprintf("ffmpeg: %s\nffprobe: %s", ffmpeg, ffprobe)
}

func (ui *appUI) recentProjectPaths() []string {
	raw := ui.app.Preferences().StringList("recent_projects")
	unique := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, path := range raw {
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		unique = append(unique, path)
	}
	return unique
}

func (ui *appUI) recordRecentProject(path string) {
	if path == "" {
		return
	}
	items := []string{path}
	for _, existing := range ui.recentProjectPaths() {
		if existing != path {
			items = append(items, existing)
		}
		if len(items) >= 8 {
			break
		}
	}
	ui.app.Preferences().SetStringList("recent_projects", items)
}

func (ui *appUI) refreshRecentProjects() {
	if ui.recentProjects == nil {
		return
	}
	items := ui.recentProjectPaths()
	objects := make([]fyne.CanvasObject, 0, len(items))
	if len(items) == 0 {
		objects = append(objects, widget.NewLabel("No recent projects yet"))
	} else {
		for _, path := range items {
			currentPath := path
			objects = append(objects, widget.NewButton(filepathBase(currentPath), func() {
				if err := ui.loadProjectFromPath(currentPath); err != nil {
					dialog.ShowError(err, ui.window)
				}
			}))
		}
	}
	ui.recentProjects.Objects = objects
	ui.recentProjects.Refresh()
}

func (ui *appUI) loadProjectFromPath(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var loaded Project
	if err := json.Unmarshal(raw, &loaded); err != nil {
		return err
	}
	ui.stopPlaybackLoop()
	ui.project = normalizeProject(loaded)
	ui.ensureSeed()
	ui.projectPath = path
	ui.recordRecentProject(path)
	ui.applyProjectToWidgets()
	ui.refreshDerivedUI()
	return nil
}
