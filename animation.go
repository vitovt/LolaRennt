package main

import (
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

type animatedFrame struct {
	Text  string
	Frame int
	Time  float64
}

func buildAnimatedFrame(project Project, stats textStats, frame int) animatedFrame {
	fps := maxInt(project.Export.FPS, 1)
	normalizedFrame := maxInt(frame-project.Export.StartFrame, 0)
	return animatedFrame{
		Text:  animateText(project, stats.DisplayText, normalizedFrame, fps),
		Frame: frame,
		Time:  float64(normalizedFrame) / float64(fps),
	}
}

func animateText(project Project, finalText string, frame, fps int) string {
	runes := []rune(finalText)
	animatable := animatablePositions(runes)
	if len(animatable) == 0 {
		return finalText
	}

	introFrames := int(project.Animation.IntroDelay * float64(fps))
	outroFrames := int(project.Animation.OutroHold * float64(fps))
	totalFrames := maxInt(int(project.Animation.TotalDuration*float64(fps)), introFrames+outroFrames+1)
	finalLockFrame := maxInt(totalFrames-outroFrames, 1)

	if frame >= finalLockFrame {
		return finalText
	}

	perCharFrames := maxInt(int(project.Animation.PerCharacterDelay*float64(fps)), 1)
	scrambleWindow := maxInt(finalLockFrame/maxInt(len(animatable), 1), 3)
	scrambleWindow = maxInt(scrambleWindow, perCharFrames*2)
	order := orderIndices(animatable, runes, project.Animation.LockOrder, project.Animation.Seed)
	pool := buildRandomPool(project)
	seedBase := hashSeed(project.Animation.Seed)

	out := make([]rune, len(runes))
	copy(out, runes)

	for rank, idx := range order {
		r := runes[idx]
		if r == '\n' || r == ' ' || r == '\t' {
			continue
		}

		start := introFrames + rank*perCharFrames
		lockAt := minInt(start+scrambleWindow, finalLockFrame)
		if frame >= lockAt {
			continue
		}

		if frame < start {
			out[idx] = scrambleRune(pool, seedBase, frame, idx, r, project.Animation.AllowEmptyCell)
			continue
		}

		progress := float64(frame-start) / float64(maxInt(lockAt-start, 1))
		switch project.Animation.LockMode {
		case "Hard lock":
			out[idx] = scrambleRune(pool, seedBase, frame, idx, r, project.Animation.AllowEmptyCell)
		default:
			if deterministicChance(seedBase, frame, idx) < progress {
				out[idx] = r
			} else {
				out[idx] = scrambleRune(pool, seedBase, frame, idx, r, project.Animation.AllowEmptyCell)
			}
		}
	}

	return string(out)
}

func buildRandomPool(project Project) []rune {
	switch project.Animation.RandomSource {
	case "Digits only":
		return []rune("0123456789")
	case "Letters only":
		return []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	case "Alphanumeric":
		return []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	default:
		pool := make([]rune, 0, 128)
		for r := range supportedRunes(project.Charset.Languages) {
			if r == '\n' || r == ' ' || r == '\t' {
				continue
			}
			if !project.Animation.AllowInvalidRandomChars && !isLetterOrDigit(r) {
				continue
			}
			pool = append(pool, r)
		}
		sort.Slice(pool, func(i, j int) bool { return pool[i] < pool[j] })
		if len(pool) == 0 {
			return []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
		}
		return pool
	}
}

func scrambleRune(pool []rune, seedBase uint64, frame, index int, fallback rune, allowEmpty bool) rune {
	if allowEmpty && deterministicChance(seedBase, frame, index+991) < 0.08 {
		return ' '
	}
	if len(pool) == 0 {
		return fallback
	}
	rng := rand.New(rand.NewSource(int64(seedBase) + int64(frame*131) + int64(index*17)))
	return pool[rng.Intn(len(pool))]
}

func deterministicChance(seedBase uint64, frame, index int) float64 {
	value := hashSeed(strconv.FormatUint(seedBase, 10) + ":" + strconv.Itoa(frame) + ":" + strconv.Itoa(index))
	return float64(value%1000) / 1000.0
}

func hashSeed(seed string) uint64 {
	h := fnv.New64a()
	if strings.TrimSpace(seed) == "" {
		seed = "default-seed"
	}
	_, _ = h.Write([]byte(seed))
	return h.Sum64()
}

func animatablePositions(runes []rune) []int {
	positions := make([]int, 0, len(runes))
	for i, r := range runes {
		if r == '\n' {
			continue
		}
		positions = append(positions, i)
	}
	return positions
}

func orderIndices(positions []int, runes []rune, order, seed string) []int {
	ordered := append([]int(nil), positions...)
	switch order {
	case "Right-to-left":
		sort.SliceStable(ordered, func(i, j int) bool { return ordered[i] > ordered[j] })
	case "Center-out":
		center := float64(len(runes)-1) / 2.0
		sort.SliceStable(ordered, func(i, j int) bool {
			di := math.Abs(float64(ordered[i]) - center)
			dj := math.Abs(float64(ordered[j]) - center)
			if di == dj {
				return ordered[i] < ordered[j]
			}
			return di < dj
		})
	case "Random":
		rng := rand.New(rand.NewSource(int64(hashSeed(seed))))
		rng.Shuffle(len(ordered), func(i, j int) {
			ordered[i], ordered[j] = ordered[j], ordered[i]
		})
	case "By lines":
		type linePos struct {
			index    int
			line     int
			position int
		}
		seq := make([]linePos, 0, len(ordered))
		line := 0
		pos := 0
		for _, idx := range ordered {
			for cursor := 0; cursor < idx; cursor++ {
				if runes[cursor] == '\n' {
					line++
					pos = 0
				} else {
					pos++
				}
			}
			seq = append(seq, linePos{index: idx, line: line, position: pos})
			line = 0
			pos = 0
		}
		sort.SliceStable(seq, func(i, j int) bool {
			if seq[i].line == seq[j].line {
				return seq[i].position < seq[j].position
			}
			return seq[i].line < seq[j].line
		})
		for i, item := range seq {
			ordered[i] = item.index
		}
	default:
		sort.SliceStable(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	}
	return ordered
}

func isLetterOrDigit(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || (r >= 'А' && r <= 'я') || r == 'Ё' || r == 'Є' || r == 'Ї' || r == 'І' || r == 'Ґ' || r == 'Ä' || r == 'Ö' || r == 'Ü' || r == 'ẞ'
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
