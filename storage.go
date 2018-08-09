package main

import (
	"image/color"
	"time"
)

var animations map[string][][][]color.RGBA // [Many][Frames][X][Y]color

var images [][][]color.RGBA // [Many][X][Y]color

func saveCanvasAsImage() {
	image = append(image, pixels)
}

func loadImageToCanvas(index int) {
	for i, p := range images {
		if i == index {
			pLock.Lock()
			pixels = p
			pLock.Unlock()
			c.Render()
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

		c.Render()
		time.Sleep(time.Millisecond * 16)
	}
}
