package main

import (
	"image/color"
	"math"
)

// This file translated from https://android.googlesource.com/platform/frameworks/support/+/master/v7/palette/src/main/java/android/support/v7/graphics/Palette.java

const MinContrastBodyText = 3
const MinContrastTitleText = 4.5
const MinAlphaSearchMaxIterations = 10
const MinAlphaSearchPrecision = 10

var white = color.RGBAModel.Convert(color.White).(color.RGBA)
var black = color.RGBAModel.Convert(color.Black).(color.RGBA)

func GenerateTextColors(rgb color.RGBA) (title, body color.RGBA) {

	whiteContrast := calculateContrast(white, rgb)
	blackContrast := calculateContrast(black, rgb)
	if whiteContrast > blackContrast {
		return white, white
	} else {
		return black, black
	}
}

func calculateMinimumAlpha(foreground color.RGBA, background color.RGBA, minContrastRatio float64) (uint8, bool) {
	if background.A != 255 {
		panic("background cannot be translucent")
	}
	// First lets check that a fully opaque foreground has sufficient contrast
	testForeground := setAlphaComponent(foreground, 255)
	testRatio := calculateContrast(testForeground, background)
	if testRatio < minContrastRatio {
		// Fully opaque foreground does not have sufficient contrast, return error
		return 0, false
	}
	// Binary search to find a value with the minimum value which provides sufficient contrast
	numIterations := 0
	var minAlpha uint8 = 0
	var maxAlpha uint8 = 255
	for numIterations <= MinAlphaSearchMaxIterations &&
		(maxAlpha-minAlpha) > MinAlphaSearchPrecision {
		testAlpha := (minAlpha + maxAlpha) / 2
		testForeground = setAlphaComponent(foreground, testAlpha)
		testRatio = calculateContrast(testForeground, background)
		if testRatio < minContrastRatio {
			minAlpha = testAlpha
		} else {
			maxAlpha = testAlpha
		}
		numIterations++
	}
	// Conservatively return the max of the range of possible alphas, which is known to pass.
	return maxAlpha, true
}

func setAlphaComponent(c color.RGBA, a uint8) color.RGBA {
	return color.RGBA{
		R: c.R,
		G: c.G,
		B: c.B,
		A: a,
	}
}

func compositeColors(foreground color.RGBA, background color.RGBA) color.RGBA {
	a := compositeAlpha(foreground.A, background.A)
	r := compositeComponent(foreground.R, foreground.A, background.R, background.A, a)
	g := compositeComponent(foreground.G, foreground.A, background.G, background.A, a)
	b := compositeComponent(foreground.B, foreground.A, background.B, background.A, a)
	return color.RGBA{
		R: r, G: g, B: b, A: a,
	}
}
func compositeAlpha(foregroundAlpha, backgroundAlpha uint8) uint8 {
	return uint8(0xFF - (((0xFF - uint64(backgroundAlpha)) * (0xFF - uint64(foregroundAlpha))) / 0xFF))
}
func compositeComponent(fgC, fgA, bgC, bgA, a uint8) uint8 {
	var (
		fgC64 = uint64(fgC)
		fgA64 = uint64(fgA)
		bgC64 = uint64(bgC)
		bgA64 = uint64(bgA)
		a64   = uint64(a)
	)
	if a == 0 {
		return 0
	}
	return uint8(((0xFF * fgC64 * fgA64) + (bgC64 * bgA64 * (0xFF - fgA64))) / (a64 * 0xFF))
}

func calculateContrast(foreground, background color.RGBA) float64 {
	if background.A != 255 {
		panic("background cannot be translucent")
	}
	if foreground.A < 255 {
		// If the foreground is translucent, composite the foreground over the background
		foreground = compositeColors(foreground, background)
	}
	luminance1 := calculateLuminance(foreground) + 0.05
	luminance2 := calculateLuminance(background) + 0.05
	// Now return the lighter luminance divided by the darker luminance
	return math.Max(luminance1, luminance2) / math.Min(luminance1, luminance2)
}

func calculateLuminance(c color.RGBA) float64 {
	red := f(float64(c.R))
	green := f(float64(c.G))
	blue := f(float64(c.B))
	return (0.2126 * red) + (0.7152 * green) + (0.0722 * blue)
}

func f(x float64) float64 {
	x /= 255
	if x < 0.03928 {
		return x / 12.92
	}
	return math.Pow((x+0.055)/1.055, 2.4)
}
