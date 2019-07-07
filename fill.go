package main

import (
	"errors"
	"fmt"
	"image/color"
	"time"
)

func (p Pixel) matches(otherP Pixel) bool {
	if p.R != otherP.R {
		return false
	}
	if p.G != otherP.G {
		return false
	}
	if p.B != otherP.B {
		return false
	}

	return true
}

func pixelFromLocation(x, y int) (Pixel, error) {
	bounds := c.Bounds()
	if x >= bounds.Max.X {
		return Pixel{}, errors.New("Pixel out of bounds")
	}

	if y >= bounds.Max.Y {
		return Pixel{}, errors.New("Pixel out of bounds")
	}

	if x < 0 || y < 0 {
		return Pixel{}, errors.New("Pixel out of bounds")
	}

	return Pixel{
		X: x,
		Y: y,
		R: pixels[x][y].R,
		G: pixels[x][y].G,
		B: pixels[x][y].B,
		A: uint8(255),
	}, nil
}

func (p Pixel) neighbors() []Pixel {

	var neighbors []Pixel

	if p.X > 0 {
		pp, err := pixelFromLocation(p.X-1, p.Y)
		if err == nil {
			if pp.matches(p) {
				neighbors = append(neighbors, pp)
			}
		}
	}

	if p.Y > 0 {
		pp, err := pixelFromLocation(p.X, p.Y-1)
		if err == nil {
			if pp.matches(p) {
				neighbors = append(neighbors, pp)
			}
		}
	}

	if p.X <= len(pixels)-1 {
		pp, err := pixelFromLocation(p.X+1, p.Y)
		if err == nil {
			if pp.matches(p) {
				neighbors = append(neighbors, pp)
			}
		}
	}

	if p.X <= len(pixels[len(pixels)-1])-1 {
		pp, err := pixelFromLocation(p.X, p.Y+1)
		if err == nil {
			if pp.matches(p) {
				neighbors = append(neighbors, pp)
			}
		}
	}

	return neighbors
}

func (p Pixel) in(pix []Pixel) bool {
	for _, aPixel := range pix {
		if aPixel.X == p.X && aPixel.Y == p.Y {
			return true
		}
	}
	return false
}

func fill(p Pixel, speed int) {

	//log.Printf("Filling pixel %v", p)

	var neighbors []Pixel
	fillable := make(map[string]Pixel)

	origPix, _ := pixelFromLocation(p.X, p.Y)

	neighbors = origPix.neighbors()

	//log.Printf("Original neighbors: %#v", neighbors)

	for {

		if len(neighbors) == 0 {
			break
		}

		var positives []Pixel

		for _, neighborPx := range neighbors {
			key := fmt.Sprintf("%02d%02d", neighborPx.X, neighborPx.Y)
			if _, ok := fillable[key]; !ok {
				positives = append(positives, neighborPx)
				fillable[key] = neighborPx
			}
		}

		neighbors = []Pixel{}

		//log.Printf("Positives: %d", len(positives))

		for _, posPX := range positives {
			neighbors = append(neighbors, posPX.neighbors()...)
		}

		//log.Printf("Fillable: %d, Neighbors: %d", len(fillable), len(neighbors))

		if speed > 0 {
			pLock.Lock()
			for _, px := range fillable {
				pixels[px.X][px.Y] = color.RGBA{p.R, p.G, p.B, p.A}
			}
			pLock.Unlock()
			drawCanvas()

			time.Sleep(time.Millisecond * time.Duration(speed))
		}

	}

	pLock.Lock()
	for _, px := range fillable {
		pixels[px.X][px.Y] = color.RGBA{p.R, p.G, p.B, p.A}
	}
	pLock.Unlock()
	drawCanvas()

}
