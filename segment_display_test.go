package main

import "testing"

func TestSegmentGlyphCoverage(t *testing.T) {
	for r := range supportedRunes(languageOptions) {
		if r == ' ' || r == '\n' {
			continue
		}
		glyph, ok := segmentGlyphs[r]
		if !ok {
			t.Fatalf("missing segment glyph for %q", r)
		}
		if len(glyph.segments) == 0 && len(glyph.strokes) == 0 {
			t.Fatalf("empty segment glyph for %q", r)
		}
	}
}

func TestSegmentStrokesAreLines(t *testing.T) {
	for _, id := range allSegmentIDs {
		stroke, ok := segmentStrokeByID[id]
		if !ok {
			t.Fatalf("missing stroke geometry for segment %d", id)
		}
		assertNonZeroStroke(t, stroke)
	}

	for r, glyph := range segmentGlyphs {
		for _, stroke := range glyph.strokes {
			if stroke.widthScale <= 0 {
				t.Fatalf("invalid width scale for %q", r)
			}
			assertNonZeroStroke(t, stroke)
		}
	}
}

func assertNonZeroStroke(t *testing.T, stroke segmentStrokeUnit) {
	t.Helper()
	if stroke.x1 == stroke.x2 && stroke.y1 == stroke.y2 {
		t.Fatalf("zero-length segment stroke: %+v", stroke)
	}
}
