package main

import (
	"os/exec"
	"path/filepath"
)

type ffmpegTools struct {
	FFmpegPath  string
	FFprobePath string
}

func detectFFmpegTools() ffmpegTools {
	tools := ffmpegTools{}
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		tools.FFmpegPath = filepath.Clean(path)
	}
	if path, err := exec.LookPath("ffprobe"); err == nil {
		tools.FFprobePath = filepath.Clean(path)
	}
	return tools
}
