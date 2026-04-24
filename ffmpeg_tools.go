package main

import (
	"os/exec"
	"path/filepath"
	"strings"
)

type ffmpegTools struct {
	FFmpegPath    string
	FFprobePath   string
	FFmpegManual  bool
	FFprobeManual bool
}

func detectFFmpegTools() ffmpegTools {
	tools := ffmpegTools{}
	tools.FFmpegPath = detectToolPath("ffmpeg")
	tools.FFprobePath = detectToolPath("ffprobe")
	return tools
}

func resolveFFmpegTools(project Project) ffmpegTools {
	tools := detectFFmpegTools()
	if override := cleanToolOverride(project.Export.FFmpegPath); override != "" {
		tools.FFmpegPath = override
		tools.FFmpegManual = true
	}
	if override := cleanToolOverride(project.Export.FFprobePath); override != "" {
		tools.FFprobePath = override
		tools.FFprobeManual = true
	}
	return tools
}

func detectToolPath(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return filepath.Clean(path)
}

func cleanToolOverride(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	return filepath.Clean(trimmed)
}

func defaultToolPath(path string, fallback string) string {
	cleaned := cleanToolOverride(path)
	if cleaned != "" {
		return cleaned
	}
	return fallback
}
