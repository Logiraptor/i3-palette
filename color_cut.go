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

type axis int

const (
	RedAxis = iota
	GreenAxis
	BlueAxis
)

type box struct {
	min, max, rng color.RGBA
	colors        []color.RGBA
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

	firstBox := &box{colors: colors}
	firstBox.fit()
	var boxes = []*box{
		firstBox,
	}
	l := cap(p) - len(p)
	for i := 0; i < l; i++ {
		sort.Slice(boxes, func(i, j int) bool {
			return (boxes[i].volume() * len(boxes[i].colors)) < (boxes[j].volume() * len(boxes[j].colors))
		})

		selectedBox := boxes[len(boxes)-1]
		boxes = append(boxes, selectedBox.split())
	}

	for _, b := range boxes {
		p = append(p, b.avg())
	}
	return p
}

func (b *box) longestEdgeAccessor() func(color.RGBA) uint8 {
	if b.rng.R >= b.rng.G && b.rng.R >= b.rng.B {
		return func(c color.RGBA) uint8 {
			return c.R
		}
	} else if b.rng.G >= b.rng.R && b.rng.G >= b.rng.B {
		return func(c color.RGBA) uint8 {
			return c.G
		}
	} else {
		return func(c color.RGBA) uint8 {
			return c.B
		}
	}
}

func (b *box) volume() int {
	return int(b.rng.R) * int(b.rng.G) * int(b.rng.B)
}

func (b *box) fit() {
	b.min = color.RGBA{math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8}
	b.max = color.RGBA{0, 0, 0, 0}
	b.rng = color.RGBA{0, 0, 0, 0}

	// fit bounding box
	for _, c := range b.colors {
		b.min.R = uint8min(b.min.R, c.R)
		b.min.G = uint8min(b.min.G, c.G)
		b.min.B = uint8min(b.min.B, c.B)

		b.max.R = uint8max(b.max.R, c.R)
		b.max.G = uint8max(b.max.G, c.G)
		b.max.B = uint8max(b.max.B, c.B)
	}

	// find longest axis
	b.rng.R = b.max.R - b.min.R
	b.rng.G = b.max.G - b.min.G
	b.rng.B = b.max.B - b.min.B
}

func (b *box) split() *box {
	b.fit()

	var component = b.longestEdgeAccessor()
	var midColor uint8

	midColor = uint8((uint16(component(b.max)) + uint16(component(b.min))) >> 1)

	sort.Slice(b.colors, func(i, j int) bool {
		return component(b.colors[i]) < component(b.colors[j])
	})

	midPoint := sort.Search(len(b.colors), func(i int) bool {
		return component(b.colors[i]) > midColor
	})

	newBox := &box{
		colors: b.colors[:midPoint],
	}
	b.colors = b.colors[midPoint:]

	b.fit()
	newBox.fit()

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
