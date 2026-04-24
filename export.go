package main

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
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

		img, err := renderExportImage(project, stats, frame)
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

func renderExportImage(project Project, stats textStats, frame int) (image.Image, error) {
	width := project.Export.Width
	height := project.Export.Height
	supersampling := project.Export.Supersampling
	if supersampling <= 1 {
		rendered, err := renderImage(project, stats, frame, width, height)
		if err != nil {
			return nil, err
		}
		return renderExportScope(rendered, project.Export, width, height), nil
	}

	renderWidth := maxInt(int(float64(width)*supersampling+0.5), width)
	renderHeight := maxInt(int(float64(height)*supersampling+0.5), height)
	highRes, err := renderImage(project, stats, frame, renderWidth, renderHeight)
	if err != nil {
		return nil, err
	}

	return renderExportScope(highRes, project.Export, width, height), nil
}

func renderExportScope(src image.Image, export ExportSettings, width, height int) image.Image {
	target := image.NewNRGBA(image.Rect(0, 0, width, height))
	sourceBounds := src.Bounds()
	if export.FrameScope == exportScopePreviewRegion {
		sourceBounds = exportRegionBounds(src.Bounds(), export.PreviewRegion)
	}
	xdraw.CatmullRom.Scale(target, target.Bounds(), src, sourceBounds, draw.Src, nil)
	return target
}

func exportRegionBounds(bounds image.Rectangle, region ExportRegion) image.Rectangle {
	region = normalizeExportRegion(region)
	if region.Width >= 1 && region.Height >= 1 {
		return bounds
	}

	x1 := bounds.Min.X + int(float64(bounds.Dx())*region.X)
	y1 := bounds.Min.Y + int(float64(bounds.Dy())*region.Y)
	x2 := bounds.Min.X + int(float64(bounds.Dx())*(region.X+region.Width))
	y2 := bounds.Min.Y + int(float64(bounds.Dy())*(region.Y+region.Height))

	if x2 <= x1 {
		x2 = x1 + 1
	}
	if y2 <= y1 {
		y2 = y1 + 1
	}
	if x2 > bounds.Max.X {
		x2 = bounds.Max.X
	}
	if y2 > bounds.Max.Y {
		y2 = bounds.Max.Y
	}
	return image.Rect(x1, y1, x2, y2)
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
