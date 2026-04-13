package main

import (
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	languageOptions = []string{"English", "German", "Ukrainian", "Russian"}

	punctuationRunes = []rune{
		' ', '\n', '.', ',', ':', ';', '!', '?', '-', '(', ')', '/', '\'', '"',
	}

	languageRuneSets = map[string]string{
		"English":   "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		"German":    "ABCDEFGHIJKLMNOPQRSTUVWXYZÄÖÜẞ0123456789",
		"Ukrainian": "АБВГҐДЕЄЖЗИІЇЙКЛМНОПРСТУФХЦЧШЩЬЮЯ0123456789",
		"Russian":   "АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ0123456789",
	}
)

type textStats struct {
	CharacterCount    int
	LineCount         int
	UnsupportedCount  int
	UnsupportedUnique []rune
	DisplayText       string
}

func sortedLanguages(languages []string) []string {
	unique := make([]string, 0, len(languages))
	seen := map[string]bool{}
	for _, lang := range languages {
		if !seen[lang] {
			seen[lang] = true
			unique = append(unique, lang)
		}
	}
	slices.Sort(unique)
	return unique
}

func supportedRunes(languages []string) map[rune]bool {
	supported := map[rune]bool{}
	for _, r := range punctuationRunes {
		supported[r] = true
	}
	for _, lang := range languages {
		for _, r := range languageRuneSets[lang] {
			supported[r] = true
		}
	}
	return supported
}

func analyzeText(text string, languages []string, uppercaseOnly, autoReplace bool) textStats {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	if uppercaseOnly {
		normalized = strings.ToUpper(normalized)
	}

	if normalized == "" {
		return textStats{
			LineCount:   1,
			DisplayText: "",
		}
	}

	supported := supportedRunes(languages)
	unsupportedSeen := map[rune]bool{}
	unsupportedUnique := make([]rune, 0, 8)
	out := strings.Builder{}
	charCount := 0

	for _, r := range normalized {
		charCount++
		if supported[r] {
			out.WriteRune(r)
			continue
		}

		if !unsupportedSeen[r] {
			unsupportedSeen[r] = true
			unsupportedUnique = append(unsupportedUnique, r)
		}

		if autoReplace {
			out.WriteRune('?')
		} else {
			out.WriteRune(r)
		}
	}

	lineCount := strings.Count(normalized, "\n") + 1
	return textStats{
		CharacterCount:    charCount,
		LineCount:         lineCount,
		UnsupportedCount:  len(unsupportedUnique),
		UnsupportedUnique: unsupportedUnique,
		DisplayText:       out.String(),
	}
}

func formatUnsupportedRunes(runes []rune) string {
	if len(runes) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(runes))
	for _, r := range runes {
		parts = append(parts, string(r))
	}
	return strings.Join(parts, " ")
}

func ensureSeed(seed string) string {
	if strings.TrimSpace(seed) != "" {
		return strings.TrimSpace(seed)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strconv.FormatInt(rng.Int63(), 10)
}
