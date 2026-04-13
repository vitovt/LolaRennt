package main

type namedPreset struct {
	Name  string
	Apply func(*Project)
}

var stylePresets = []namedPreset{
	{
		Name: "Red LED Classic",
		Apply: func(p *Project) {
			p.Style.MainColor = "#FF6040"
			p.Style.GlowColor = "#FF8C66"
			p.Style.InactiveColor = "#31140E"
			p.Style.GlowIntensity = 58
			p.Style.InactiveVisibility = 18
			p.Display.Mode = displayModeSegment
		},
	},
	{
		Name: "Orange VFD",
		Apply: func(p *Project) {
			p.Style.MainColor = "#FFB347"
			p.Style.GlowColor = "#FFD28B"
			p.Style.InactiveColor = "#3C260B"
			p.Style.GlowIntensity = 42
			p.Style.InactiveVisibility = 12
			p.Display.Mode = displayModeSegment
		},
	},
	{
		Name: "Neon Cyan",
		Apply: func(p *Project) {
			p.Style.MainColor = "#4FF2FF"
			p.Style.GlowColor = "#A5FBFF"
			p.Style.InactiveColor = "#12333A"
			p.Style.GlowIntensity = 76
			p.Style.InactiveVisibility = 16
			p.Display.Mode = displayModeDotMatrix
		},
	},
	{
		Name: "Cyber Green",
		Apply: func(p *Project) {
			p.Style.MainColor = "#7DFF72"
			p.Style.GlowColor = "#C3FFBC"
			p.Style.InactiveColor = "#163117"
			p.Style.GlowIntensity = 68
			p.Style.InactiveVisibility = 14
			p.Display.Mode = displayModeDotMatrix
		},
	},
}

var animationPresets = []namedPreset{
	{
		Name: "Fast scramble",
		Apply: func(p *Project) {
			p.Animation.Type = "Scramble basic"
			p.Animation.TotalDuration = 1.8
			p.Animation.IntroDelay = 0
			p.Animation.OutroHold = 0.2
			p.Animation.PerCharacterDelay = 0.02
			p.Animation.RandomSwitchRate = 28
			p.Animation.LockOrder = "Left-to-right"
			p.Animation.LockMode = "Hard lock"
		},
	},
	{
		Name: "Dramatic lock",
		Apply: func(p *Project) {
			p.Animation.Type = "Scramble with lock"
			p.Animation.TotalDuration = 3.6
			p.Animation.IntroDelay = 0.2
			p.Animation.OutroHold = 0.8
			p.Animation.PerCharacterDelay = 0.05
			p.Animation.RandomSwitchRate = 18
			p.Animation.LockOrder = "Center-out"
			p.Animation.LockMode = "Probabilistic lock"
		},
	},
}

var exportPresets = []namedPreset{
	{
		Name: "YouTube 1080p",
		Apply: func(p *Project) {
			p.Export.Width = 1920
			p.Export.Height = 1080
			p.Export.FPS = 30
			p.Export.Supersampling = 1
		},
	},
	{
		Name: "Shorts Vertical",
		Apply: func(p *Project) {
			p.Export.Width = 1080
			p.Export.Height = 1920
			p.Export.FPS = 30
			p.Export.Supersampling = 1
		},
	},
	{
		Name: "Transparent Master",
		Apply: func(p *Project) {
			p.Background.Mode = backgroundModeTransparent
			p.Export.Width = 1920
			p.Export.Height = 1080
			p.Export.FPS = 30
			p.Export.Supersampling = 2
		},
	},
}

func presetNames(items []namedPreset) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names
}

func applyPresetByName(project *Project, name string, items []namedPreset) bool {
	for _, item := range items {
		if item.Name == name {
			item.Apply(project)
			return true
		}
	}
	return false
}
