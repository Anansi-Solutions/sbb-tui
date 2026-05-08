package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/lucasb-eyer/go-colorful"
)

const (
	animTaglineBuild = "taglineBuild"

	// taglineBuildFadeWindow is the fraction of the total animation
	// each character spends fading in. Smaller than the logo build's
	// because the tagline has many characters and a snappier feel
	// reads as "typing" more than "fading".
	taglineBuildFadeWindow = 0.20
)

// renderTaglineBuild paints `text` left-to-right, with each rune
// fading from invisible to the base color over a short window. The
// total animation completes at progress=1.
func renderTaglineBuild(text string, base colorful.Color, progress float64) string {
	n := utf8.RuneCountInString(text)
	if n == 0 {
		return text
	}

	palette := buildLogoBuildPalette(base)

	var b strings.Builder
	b.Grow(len(text) * 6)
	i := 0
	for _, r := range text {
		if r == ' ' {
			b.WriteRune(r)
			i++
			continue
		}
		norm := float64(i) / float64(n)
		factor := taglineBuildFactor(progress, norm)
		b.WriteString(palette.render(factor, r))
		i++
	}
	return b.String()
}

func taglineBuildFactor(progress, norm float64) float64 {
	start := norm * (1 - taglineBuildFadeWindow)
	end := start + taglineBuildFadeWindow
	if progress <= start {
		return 0
	}
	if progress >= end {
		return 1
	}
	t := (progress - start) / taglineBuildFadeWindow
	return t * t * (3 - 2*t)
}
