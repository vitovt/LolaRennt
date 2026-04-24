package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type presetKind string

const (
	presetKindStyle     presetKind = "style"
	presetKindAnimation presetKind = "animation"
	presetKindExport    presetKind = "export"
)

type projectPreset struct {
	Name    string  `json:"name"`
	Project Project `json:"project"`
}

type presetLibrary struct {
	Style     []projectPreset `json:"style"`
	Animation []projectPreset `json:"animation"`
	Export    []projectPreset `json:"export"`
}

func defaultPresetLibrary() presetLibrary {
	return presetLibrary{
		Style: []projectPreset{
			makeProjectPreset("Red LED Classic", func(p *Project) {
				p.Style.MainColor = "#FF6040"
				p.Style.GlowColor = "#FF8C66"
				p.Style.InactiveColor = "#31140E"
				p.Style.GlowIntensity = 58
				p.Style.InactiveVisibility = 18
				p.Display.Mode = displayModeSegment
			}),
			makeProjectPreset("Orange VFD", func(p *Project) {
				p.Style.MainColor = "#FFB347"
				p.Style.GlowColor = "#FFD28B"
				p.Style.InactiveColor = "#3C260B"
				p.Style.GlowIntensity = 42
				p.Style.InactiveVisibility = 12
				p.Display.Mode = displayModeSegment
			}),
			makeProjectPreset("Neon Cyan", func(p *Project) {
				p.Style.MainColor = "#4FF2FF"
				p.Style.GlowColor = "#A5FBFF"
				p.Style.InactiveColor = "#12333A"
				p.Style.GlowIntensity = 76
				p.Style.InactiveVisibility = 16
				p.Display.Mode = displayModeDotMatrix
			}),
			makeProjectPreset("Cyber Green", func(p *Project) {
				p.Style.MainColor = "#7DFF72"
				p.Style.GlowColor = "#C3FFBC"
				p.Style.InactiveColor = "#163117"
				p.Style.GlowIntensity = 68
				p.Style.InactiveVisibility = 14
				p.Display.Mode = displayModeDotMatrix
			}),
		},
		Animation: []projectPreset{
			makeProjectPreset("Fast scramble", func(p *Project) {
				p.Animation.Type = "Scramble basic"
				p.Animation.TotalDuration = 1.8
				p.Animation.IntroDelay = 0
				p.Animation.OutroHold = 0.2
				p.Animation.PerCharacterDelay = 0.02
				p.Animation.RandomSwitchRate = 28
				p.Animation.LockOrder = "Left-to-right"
				p.Animation.LockMode = "Hard lock"
			}),
			makeProjectPreset("Dramatic lock", func(p *Project) {
				p.Animation.Type = "Scramble with lock"
				p.Animation.TotalDuration = 3.6
				p.Animation.IntroDelay = 0.2
				p.Animation.OutroHold = 0.8
				p.Animation.PerCharacterDelay = 0.05
				p.Animation.RandomSwitchRate = 18
				p.Animation.LockOrder = "Center-out"
				p.Animation.LockMode = "Probabilistic lock"
			}),
		},
		Export: []projectPreset{
			makeProjectPreset("YouTube 1080p", func(p *Project) {
				p.Export.Width = 1920
				p.Export.Height = 1080
				p.Export.FPS = 30
				p.Export.Supersampling = 1
			}),
			makeProjectPreset("Shorts Vertical", func(p *Project) {
				p.Export.Width = 1080
				p.Export.Height = 1920
				p.Export.FPS = 30
				p.Export.Supersampling = 1
			}),
			makeProjectPreset("Transparent Master", func(p *Project) {
				p.Background.Mode = backgroundModeTransparent
				p.Export.Width = 1920
				p.Export.Height = 1080
				p.Export.FPS = 30
				p.Export.Supersampling = 2
			}),
		},
	}
}

func makeProjectPreset(name string, apply func(*Project)) projectPreset {
	project := normalizeProject(defaultProject())
	apply(&project)
	return projectPreset{
		Name:    name,
		Project: normalizeProject(project),
	}
}

func loadPresetLibrary() (presetLibrary, error) {
	path, err := presetLibraryPath()
	if err != nil {
		return defaultPresetLibrary(), err
	}

	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		library := defaultPresetLibrary()
		if saveErr := savePresetLibrary(library); saveErr != nil {
			return library, saveErr
		}
		return library, nil
	}
	if err != nil {
		return defaultPresetLibrary(), err
	}

	var library presetLibrary
	if err := json.Unmarshal(raw, &library); err != nil {
		return defaultPresetLibrary(), err
	}
	return normalizePresetLibrary(library), nil
}

func savePresetLibrary(library presetLibrary) error {
	path, err := presetLibraryPath()
	if err != nil {
		return err
	}

	library = normalizePresetLibrary(library)
	raw, err := json.MarshalIndent(library, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func presetLibraryPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "LolaRennt", "presets.json"), nil
}

func normalizePresetLibrary(library presetLibrary) presetLibrary {
	defaults := defaultPresetLibrary()
	library.Style = normalizePresetBucket(library.Style, defaults.Style)
	library.Animation = normalizePresetBucket(library.Animation, defaults.Animation)
	library.Export = normalizePresetBucket(library.Export, defaults.Export)
	return library
}

func normalizePresetBucket(items []projectPreset, fallback []projectPreset) []projectPreset {
	if items == nil {
		return clonePresetBucket(fallback)
	}

	normalized := make([]projectPreset, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		item.Name = name
		item.Project = normalizeProject(item.Project)
		normalized = append(normalized, item)
		seen[key] = true
	}
	return normalized
}

func clonePresetBucket(items []projectPreset) []projectPreset {
	cloned := make([]projectPreset, len(items))
	copy(cloned, items)
	return cloned
}

func presetNames(items []projectPreset) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names
}

func firstPresetName(items []projectPreset) string {
	if len(items) == 0 {
		return ""
	}
	return items[0].Name
}

func findPresetByName(name string, items []projectPreset) (projectPreset, bool) {
	for _, item := range items {
		if strings.EqualFold(item.Name, name) {
			return item, true
		}
	}
	return projectPreset{}, false
}

func presetNameExists(name string, items []projectPreset) bool {
	_, ok := findPresetByName(name, items)
	return ok
}

func uniquePresetName(base string, items []projectPreset) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "Preset"
	}
	if !presetNameExists(base, items) {
		return base
	}
	for suffix := 2; suffix < 1000; suffix++ {
		candidate := fmt.Sprintf("%s %d", base, suffix)
		if !presetNameExists(candidate, items) {
			return candidate
		}
	}
	return base + " Copy"
}

func upsertPreset(items []projectPreset, preset projectPreset) []projectPreset {
	preset.Name = strings.TrimSpace(preset.Name)
	if preset.Name == "" {
		return items
	}
	updated := clonePresetBucket(items)
	for index, item := range updated {
		if strings.EqualFold(item.Name, preset.Name) {
			updated[index] = preset
			return updated
		}
	}
	return append(updated, preset)
}

func removePresetByName(items []projectPreset, name string) ([]projectPreset, bool) {
	updated := make([]projectPreset, 0, len(items))
	removed := false
	for _, item := range items {
		if strings.EqualFold(item.Name, name) {
			removed = true
			continue
		}
		updated = append(updated, item)
	}
	return updated, removed
}

func capturePresetFromProject(kind presetKind, name string, project Project) projectPreset {
	snapshot := normalizeProject(defaultProject())
	snapshot.Text = project.Text
	snapshot.Charset = project.Charset

	switch kind {
	case presetKindStyle:
		snapshot.Display = project.Display
		snapshot.Style = project.Style
		snapshot.Layout = project.Layout
	case presetKindAnimation:
		snapshot.Animation = project.Animation
	case presetKindExport:
		snapshot.Background = project.Background
		snapshot.Export.Width = project.Export.Width
		snapshot.Export.Height = project.Export.Height
		snapshot.Export.FPS = project.Export.FPS
		snapshot.Export.StartFrame = project.Export.StartFrame
		snapshot.Export.EndFrame = project.Export.EndFrame
		snapshot.Export.FrameScope = project.Export.FrameScope
		snapshot.Export.Supersampling = project.Export.Supersampling
		snapshot.Export.PreviewRegion = fullExportRegion()
	}

	return projectPreset{
		Name:    strings.TrimSpace(name),
		Project: normalizeProject(snapshot),
	}
}

func applyStylePresetByName(project *Project, name string, items []projectPreset) bool {
	preset, ok := findPresetByName(name, items)
	if !ok {
		return false
	}
	project.Style = preset.Project.Style
	project.Display = preset.Project.Display
	return true
}

func applyAnimationPresetByName(project *Project, name string, items []projectPreset) bool {
	preset, ok := findPresetByName(name, items)
	if !ok {
		return false
	}
	project.Animation = preset.Project.Animation
	return true
}

func applyExportPresetByName(project *Project, name string, items []projectPreset) bool {
	preset, ok := findPresetByName(name, items)
	if !ok {
		return false
	}
	project.Background = preset.Project.Background
	project.Export.Width = preset.Project.Export.Width
	project.Export.Height = preset.Project.Export.Height
	project.Export.FPS = preset.Project.Export.FPS
	project.Export.StartFrame = preset.Project.Export.StartFrame
	project.Export.EndFrame = preset.Project.Export.EndFrame
	project.Export.FrameScope = preset.Project.Export.FrameScope
	project.Export.Supersampling = preset.Project.Export.Supersampling
	project.Export.PreviewRegion = fullExportRegion()
	return true
}
