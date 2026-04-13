package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os/exec"
	"sync"
)

var (
	videoFrameCacheMu sync.Mutex
	videoFrameCache   = map[string]image.Image{}
)

func loadVideoBackgroundFrame(path string, timeSec float64) (image.Image, error) {
	key := fmt.Sprintf("%s@%.3f", path, timeSec)
	videoFrameCacheMu.Lock()
	if cached, ok := videoFrameCache[key]; ok {
		videoFrameCacheMu.Unlock()
		return cached, nil
	}
	videoFrameCacheMu.Unlock()

	cmd := exec.Command(
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-ss", fmt.Sprintf("%.3f", timeSec),
		"-i", path,
		"-frames:v", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"pipe:1",
	)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg video frame extraction failed: %w: %s", err, stderr.String())
	}

	img, _, err := image.Decode(bytes.NewReader(out.Bytes()))
	if err != nil {
		return nil, err
	}

	videoFrameCacheMu.Lock()
	videoFrameCache[key] = img
	videoFrameCacheMu.Unlock()
	return img, nil
}
