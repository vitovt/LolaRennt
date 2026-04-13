package main

import (
	"encoding/json"
	"time"
)

const (
	displayModeSegment   = "Segment"
	displayModeDotMatrix = "Dot-matrix"

	colorModeSingle       = "Single color"
	colorModePerCharacter = "Per-character colors"

	backgroundModeTransparent = "Transparent"
	backgroundModeSolid       = "Solid color"
	backgroundModeGradient    = "Gradient"
	backgroundModeImage       = "Image"
	backgroundModeVideo       = "Video"

	alignmentLeft   = "Left"
	alignmentCenter = "Center"
	alignmentRight  = "Right"
)

type Project struct {
	Text       TextSettings       `json:"text"`
	Charset    CharsetSettings    `json:"charset"`
	Display    DisplaySettings    `json:"display"`
	Style      StyleSettings      `json:"style"`
	Layout     LayoutSettings     `json:"layout"`
	Animation  AnimationSettings  `json:"animation"`
	Background BackgroundSettings `json:"background"`
	Export     ExportSettings     `json:"export"`
	Metadata   MetadataSettings   `json:"metadata"`
}

type TextSettings struct {
	Content                string `json:"content"`
	UppercaseOnly          bool   `json:"uppercase_only"`
	AutoReplaceUnsupported bool   `json:"auto_replace_unsupported"`
}

type CharsetSettings struct {
	Languages []string `json:"languages"`
}

type DisplaySettings struct {
	Mode string `json:"mode"`
}

type StyleSettings struct {
	ColorMode          string  `json:"color_mode"`
	MainColor          string  `json:"main_color"`
	GlowColor          string  `json:"glow_color"`
	InactiveColor      string  `json:"inactive_color"`
	GlowIntensity      float64 `json:"glow_intensity"`
	InactiveVisibility float64 `json:"inactive_visibility"`
}

type LayoutSettings struct {
	CellScale        float64 `json:"cell_scale"`
	CharacterSpacing float64 `json:"character_spacing"`
	LineSpacing      float64 `json:"line_spacing"`
	Alignment        string  `json:"alignment"`
	Padding          float64 `json:"padding"`
}

type AnimationSettings struct {
	Type                    string  `json:"type"`
	RandomSource            string  `json:"random_source"`
	AllowInvalidRandomChars bool    `json:"allow_invalid_random_chars"`
	AllowEmptyCell          bool    `json:"allow_empty_cell"`
	TotalDuration           float64 `json:"total_duration"`
	IntroDelay              float64 `json:"intro_delay"`
	OutroHold               float64 `json:"outro_hold"`
	PerCharacterDelay       float64 `json:"per_character_delay"`
	RandomSwitchRate        float64 `json:"random_switch_rate"`
	Seed                    string  `json:"seed"`
	LockOrder               string  `json:"lock_order"`
	LockMode                string  `json:"lock_mode"`
	Loop                    bool    `json:"loop"`
}

type BackgroundSettings struct {
	Mode       string `json:"mode"`
	SolidColor string `json:"solid_color"`
	GradientA  string `json:"gradient_a"`
	GradientB  string `json:"gradient_b"`
	ImagePath  string `json:"image_path"`
	VideoPath  string `json:"video_path"`
	FitMode    string `json:"fit_mode"`
}

type ExportSettings struct {
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	FPS             int     `json:"fps"`
	StartFrame      int     `json:"start_frame"`
	EndFrame        int     `json:"end_frame"`
	OutputFolder    string  `json:"output_folder"`
	FilePrefix      string  `json:"file_prefix"`
	OverwritePolicy string  `json:"overwrite_policy"`
	Supersampling   float64 `json:"supersampling"`
}

type MetadataSettings struct {
	ProjectName string `json:"project_name"`
	Notes       string `json:"notes"`
	Tags        string `json:"tags"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func defaultProject() Project {
	now := timeNowString()
	return Project{
		Text: TextSettings{
			Content:                "LOLA: RENNT\nDIE: DEIN: LEBEN\nVERANDERN: KANN",
			UppercaseOnly:          true,
			AutoReplaceUnsupported: true,
		},
		Charset: CharsetSettings{
			Languages: []string{"English", "German", "Ukrainian", "Russian"},
		},
		Display: DisplaySettings{
			Mode: displayModeSegment,
		},
		Style: StyleSettings{
			ColorMode:          colorModeSingle,
			MainColor:          "#FF6040",
			GlowColor:          "#FF8C66",
			InactiveColor:      "#31140E",
			GlowIntensity:      58,
			InactiveVisibility: 18,
		},
		Layout: LayoutSettings{
			CellScale:        1.0,
			CharacterSpacing: 10,
			LineSpacing:      18,
			Alignment:        alignmentCenter,
			Padding:          24,
		},
		Animation: AnimationSettings{
			Type:                    "Scramble with lock",
			RandomSource:            "Current charset only",
			AllowInvalidRandomChars: false,
			AllowEmptyCell:          false,
			TotalDuration:           3.2,
			IntroDelay:              0.2,
			OutroHold:               0.6,
			PerCharacterDelay:       0.04,
			RandomSwitchRate:        20,
			Seed:                    "",
			LockOrder:               "Center-out",
			LockMode:                "Probabilistic lock",
			Loop:                    true,
		},
		Background: BackgroundSettings{
			Mode:       backgroundModeTransparent,
			SolidColor: "#050608",
			GradientA:  "#0B0D10",
			GradientB:  "#18262B",
			FitMode:    "Fit",
		},
		Export: ExportSettings{
			Width:           1920,
			Height:          1080,
			FPS:             30,
			StartFrame:      0,
			EndFrame:        119,
			OutputFolder:    ".",
			FilePrefix:      "lola_frame",
			OverwritePolicy: "Ask",
			Supersampling:   1,
		},
		Metadata: MetadataSettings{
			ProjectName: "Lola Rennt Intro",
			Notes:       "Initial v1.0 implementation slice",
			Tags:        "segment, scramble, intro",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

func normalizeProject(project Project) Project {
	def := defaultProject()

	if project.Text.Content == "" {
		project.Text.Content = def.Text.Content
	}
	if len(project.Charset.Languages) == 0 {
		project.Charset.Languages = def.Charset.Languages
	}
	if project.Display.Mode == "" {
		project.Display.Mode = def.Display.Mode
	}
	if project.Style.ColorMode == "" {
		project.Style.ColorMode = def.Style.ColorMode
	}
	if project.Style.MainColor == "" {
		project.Style.MainColor = def.Style.MainColor
	}
	if project.Style.GlowColor == "" {
		project.Style.GlowColor = def.Style.GlowColor
	}
	if project.Style.InactiveColor == "" {
		project.Style.InactiveColor = def.Style.InactiveColor
	}
	if project.Layout.Alignment == "" {
		project.Layout.Alignment = def.Layout.Alignment
	}
	if project.Layout.CellScale == 0 {
		project.Layout.CellScale = def.Layout.CellScale
	}
	if project.Background.Mode == "" {
		project.Background.Mode = def.Background.Mode
	}
	if project.Background.SolidColor == "" {
		project.Background.SolidColor = def.Background.SolidColor
	}
	if project.Background.GradientA == "" {
		project.Background.GradientA = def.Background.GradientA
	}
	if project.Background.GradientB == "" {
		project.Background.GradientB = def.Background.GradientB
	}
	if project.Background.FitMode == "" {
		project.Background.FitMode = def.Background.FitMode
	}
	if project.Animation.Type == "" {
		project.Animation.Type = def.Animation.Type
	}
	if project.Animation.RandomSource == "" {
		project.Animation.RandomSource = def.Animation.RandomSource
	}
	if project.Animation.TotalDuration == 0 {
		project.Animation.TotalDuration = def.Animation.TotalDuration
	}
	if project.Animation.RandomSwitchRate == 0 {
		project.Animation.RandomSwitchRate = def.Animation.RandomSwitchRate
	}
	if project.Animation.LockOrder == "" {
		project.Animation.LockOrder = def.Animation.LockOrder
	}
	if project.Animation.LockMode == "" {
		project.Animation.LockMode = def.Animation.LockMode
	}
	if project.Export.Width == 0 {
		project.Export.Width = def.Export.Width
	}
	if project.Export.Height == 0 {
		project.Export.Height = def.Export.Height
	}
	if project.Export.FPS == 0 {
		project.Export.FPS = def.Export.FPS
	}
	if project.Export.EndFrame == 0 {
		project.Export.EndFrame = def.Export.EndFrame
	}
	if project.Export.FilePrefix == "" {
		project.Export.FilePrefix = def.Export.FilePrefix
	}
	if project.Export.OverwritePolicy == "" {
		project.Export.OverwritePolicy = def.Export.OverwritePolicy
	}
	if project.Export.Supersampling == 0 {
		project.Export.Supersampling = def.Export.Supersampling
	}
	if project.Metadata.ProjectName == "" {
		project.Metadata.ProjectName = def.Metadata.ProjectName
	}
	if project.Metadata.CreatedAt == "" {
		project.Metadata.CreatedAt = timeNowString()
	}
	if project.Metadata.UpdatedAt == "" {
		project.Metadata.UpdatedAt = project.Metadata.CreatedAt
	}
	return project
}

func (p Project) Marshal() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

func timeNowString() string {
	return time.Now().Format(time.RFC3339)
}
