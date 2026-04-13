package main

import (
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
)

type renderProgress struct {
	CurrentFrame int
	TotalFrames  int
	Filename     string
}

var errRenderCancelled = errors.New("render cancelled")

type overwriteConflictError struct {
	Path string
}

func (e overwriteConflictError) Error() string {
	return "render output already exists: " + e.Path
}

func prepareRenderProject(project Project, allowOverwrite bool) (Project, error) {
	project = normalizeProject(project)
	outputFolder := project.Export.OutputFolder
	if outputFolder == "" {
		outputFolder = "."
	}
	prefix := project.Export.FilePrefix
	if prefix == "" {
		prefix = "frame"
	}

	switch project.Export.OverwritePolicy {
	case "Create new suffix":
		project.Export.FilePrefix = nextAvailablePrefix(outputFolder, prefix, project)
		return project, nil
	case "Overwrite":
		return project, nil
	default:
		if conflict := firstExistingFrame(outputFolder, prefix, project); conflict != "" && !allowOverwrite {
			return project, overwriteConflictError{Path: conflict}
		}
		return project, nil
	}
}

func renderPNGSequence(project Project, cancel <-chan struct{}, progress func(renderProgress)) (string, int, error) {
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

	totalFrames := maxInt(project.Export.EndFrame-project.Export.StartFrame+1, 1)
	count := 0
	for frame := project.Export.StartFrame; frame <= project.Export.EndFrame; frame++ {
		select {
		case <-cancel:
			return outputFolder, count, errRenderCancelled
		default:
		}

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
		if progress != nil {
			progress(renderProgress{
				CurrentFrame: count,
				TotalFrames:  totalFrames,
				Filename:     filename,
			})
		}
	}

	return outputFolder, count, nil
}

func nextAvailablePrefix(outputFolder, prefix string, project Project) string {
	if firstExistingFrame(outputFolder, prefix, project) == "" {
		return prefix
	}
	for suffix := 1; suffix < 1000; suffix++ {
		candidate := fmt.Sprintf("%s_%02d", prefix, suffix)
		if firstExistingFrame(outputFolder, candidate, project) == "" {
			return candidate
		}
	}
	return prefix + "_new"
}

func firstExistingFrame(outputFolder, prefix string, project Project) string {
	for frame := project.Export.StartFrame; frame <= project.Export.EndFrame; frame++ {
		filename := filepath.Join(outputFolder, fmt.Sprintf("%s_%05d.png", prefix, frame))
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}
	return ""
}
