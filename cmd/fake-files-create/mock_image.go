package main

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
)

// http://tech.nitoyon.com/en/blog/2015/12/31/go-image-gen/

type circle struct {
	X, Y, R float64
}

func (c *circle) Brightness(x, y float64) uint8 {
	var dx, dy float64 = c.X - x, c.Y - y
	d := math.Sqrt(dx*dx+dy*dy) / c.R
	if d > 1 {
		return 0
	}
	return 255
	// return uint8((1 - math.Pow(d, 5)) * 255)
}

var img *image.RGBA

func makeRawImage() *image.RGBA {

	var w, h int = 280, 240
	var hw, hh float64 = float64(w / 2), float64(h / 2)
	r := 40.0
	θ := 2 * math.Pi / 3
	cr := &circle{hw - r*math.Sin(0), hh - r*math.Cos(0), 60}
	cg := &circle{hw - r*math.Sin(θ), hh - r*math.Cos(θ), 60}
	cb := &circle{hw - r*math.Sin(-θ), hh - r*math.Cos(-θ), 60}

	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c := color.RGBA{
				cr.Brightness(float64(x), float64(y)),
				cg.Brightness(float64(x), float64(y)),
				cb.Brightness(float64(x), float64(y)),
				255,
			}
			m.Set(x, y, c)
		}
	}
	return m
}

func makePNG() []byte {
	if img == nil {
		img = makeRawImage()
	}
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)
	png.Encode(bufw, img)
	bufw.Flush()
	return buf.Bytes()
}

func makeJpeg() []byte {
	if img == nil {
		img = makeRawImage()
	}
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)
	jpeg.Encode(bufw, img, nil)
	bufw.Flush()
	return buf.Bytes()
}

// The GIF encoder is not working. I don't know why right now. -- But using an embedded gif now anyway.
/*
func makeGIF() []byte {
		if img == nil {
			img = makeRawImage()
		}
		var buf bytes.Buffer
		bufw := bufio.NewWriter(&buf)
		gif.Encode(bufw, img, nil)
		bufw.Flush()
		return buf.Bytes()
}
*/
