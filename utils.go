package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func parseHexColor(value string, fallback color.NRGBA) color.NRGBA {
	trimmed := strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(trimmed) != 6 {
		return fallback
	}

	raw, err := strconv.ParseUint(trimmed, 16, 32)
	if err != nil {
		return fallback
	}

	return color.NRGBA{
		R: uint8(raw >> 16),
		G: uint8((raw >> 8) & 0xFF),
		B: uint8(raw & 0xFF),
		A: 255,
	}
}
