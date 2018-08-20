package main

import (
	"image/color"
	"log"
	"time"
)

var animations map[string][][][]color.RGBA // [Many][Frames][X][Y]color

var images [][][]color.RGBA // [Many][X][Y]color

func saveCanvasAsImage() {
	var newPixels [][]color.RGBA
	/*
		for _, pixcol := range pixels {
			var newCol []color.RGBA
			for _, pixel := range pixcol {
				newCol = append(newCol, pixel)
			}
			newPixels = append(newPixels, newCol)
		}
	*/
	copy(newPixels, pixels)
	images = append(images, newPixels)
	for i, image := range images {
		log.Printf("Image %d pixel 0,0 is: %v", i, image[0][0])
	}
}

func loadImageToCanvas(index int) {
	for i, p := range images {
		if i == index {
			pLock.Lock()
			bounds := c.Bounds()
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					pixels[x][y] = p[x][y]
				}
			}
			log.Printf("Pixel 0,0 color: %#v", pixels[0][0])
			pLock.Unlock()
			drawCanvas()
			return
		}
	}
}

func saveCanvasAsAnimationFrame(name string, frameIndex int) {
	if frameIndex == 0 {
		animations[name] = append(animations[name], pixels)
	} else {
		animations[name][frameIndex] = pixels
	}
}

func playAnimationToCanvas(name string) {
	a := animations[name]

	for _, frame := range a {

		pLock.Lock()
		pixels = frame
		pLock.Unlock()

		drawCanvas()
		//time.Sleep(time.Millisecond * 16)
		time.Sleep(time.Millisecond * 500)
	}
}
