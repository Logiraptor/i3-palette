package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"sort"
)

type ColorCut struct {
}

type box struct {
	colors []color.RGBA
}

var _ draw.Quantizer = &ColorCut{}

func (c *ColorCut) Quantize(p color.Palette, m image.Image) color.Palette {
	bounds := m.Bounds()
	colors := make([]color.RGBA, bounds.Dx()*bounds.Dy())

	for i := 0; i < bounds.Dx(); i++ {
		for j := 0; j < bounds.Dy(); j++ {
			colors[i+j*bounds.Dx()] = color.RGBAModel.Convert(m.At(i, j)).(color.RGBA)
		}
	}

	var boxes = []*box{
		&box{colors: colors},
	}
	l := cap(p) - len(p)
	for i := 0; i < l; i++ {
		sort.Slice(boxes, func(i, j int) bool {
			return len(boxes[i].colors) < len(boxes[j].colors)
		})

		selectedBox := boxes[len(boxes)-1]
		boxes = append(boxes, selectedBox.split())
	}

	for _, b := range boxes {
		p = append(p, b.avg())
	}
	return p
}

func (b *box) split() *box {
	var (
		min = color.RGBA{math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8}
		max = color.RGBA{0, 0, 0, 0}
		rng = color.RGBA{0, 0, 0, 0}
	)

	// fit bounding box
	for _, c := range b.colors {
		min.R = uint8min(min.R, c.R)
		min.G = uint8min(min.G, c.G)
		min.B = uint8min(min.B, c.B)

		max.R = uint8max(max.R, c.R)
		max.G = uint8max(max.G, c.G)
		max.B = uint8max(max.B, c.B)
	}

	// find longest axis
	rng.R = max.R - min.R
	rng.G = max.G - min.G
	rng.B = max.B - min.B

	const (
		RED = iota
		GREEN
		BLUE
	)

	var comparator func(int, int) bool

	if rng.R > rng.G && rng.R > rng.B {
		comparator = func(i, j int) bool {
			return b.colors[i].R < b.colors[j].R
		}
	} else if rng.B > rng.G && rng.B > rng.R {
		comparator = func(i, j int) bool {
			return b.colors[i].G < b.colors[j].G
		}
	} else {
		comparator = func(i, j int) bool {
			return b.colors[i].B < b.colors[j].B
		}
	}

	sort.Slice(b.colors, comparator)

	midPoint := len(b.colors) / 2

	newBox := &box{
		colors: b.colors[:midPoint],
	}
	b.colors = b.colors[midPoint:]

	return newBox
}

func (b *box) avg() color.RGBA {
	var (
		rSum, gSum, bSum uint32
	)

	for _, c := range b.colors {
		rSum += uint32(c.R)
		gSum += uint32(c.G)
		bSum += uint32(c.B)
	}

	return color.RGBA{
		R: uint8(rSum / uint32(len(b.colors))),
		G: uint8(gSum / uint32(len(b.colors))),
		B: uint8(bSum / uint32(len(b.colors))),
		A: math.MaxUint8,
	}
}

func uint8min(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

func uint8max(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}
