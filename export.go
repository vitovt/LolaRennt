package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
)

func renderPNGSequence(project Project) (string, int, error) {
	stats := analyzeText(project.Text.Content, project.Charset.Languages, project.Text.UppercaseOnly, project.Text.AutoReplaceUnsupported)
	outputFolder := project.Export.OutputFolder
	if outputFolder == "" {
		outputFolder = "."
	}
	if err := os.MkdirAll(outputFolder, 0o755); err != nil {
		return "", 0, err
	}

	prefix := project.Export.FilePrefix
	if prefix == "" {
		prefix = "frame"
	}

	count := 0
	for frame := project.Export.StartFrame; frame <= project.Export.EndFrame; frame++ {
		img, err := renderImage(project, stats, frame, project.Export.Width, project.Export.Height)
		if err != nil {
			return "", count, err
		}

		filename := filepath.Join(outputFolder, fmt.Sprintf("%s_%05d.png", prefix, frame))
		file, err := os.Create(filename)
		if err != nil {
			return "", count, err
		}

		if err := png.Encode(file, img); err != nil {
			_ = file.Close()
			return "", count, err
		}
		if err := file.Close(); err != nil {
			return "", count, err
		}
		count++
	}

	return outputFolder, count, nil
}
