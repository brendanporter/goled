package main

import (
	"image/color"
	"image/png"
	//"log"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

var animations map[string][][][]color.RGBA // [Many][Frames][X][Y]color

var images map[string][][]color.RGBA // [Many][X][Y]color

func init() {
	images = make(map[string][][]color.RGBA)
	animations = make(map[string][][][]color.RGBA)
}

func saveImagesToDisk() {
	absPath, err := filepath.Abs("images.json")
	if err != nil {
		elog.Print(err)
	}

	imagesJSON, err := json.Marshal(images)
	if err != nil {
		elog.Print(err)
		return
	}

	err = ioutil.WriteFile(absPath, imagesJSON, 0755)
	if err != nil {
		elog.Print(err)
	}
}

func loadImagesFromDisk() {
	absPath, err := filepath.Abs("images.json")
	if err != nil {
		elog.Print(err)
	}

	fileBytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		elog.Print(err)
		return
	}

	err = json.Unmarshal(fileBytes, &images)
	if err != nil {
		elog.Print(err)
	}

	log.Printf("Loaded %d images from disk", len(images))

}

func saveCanvasAsImage(name string) {
	var newPixels [][]color.RGBA
	for _, pixcol := range pixels {
		var newCol []color.RGBA
		for _, pixel := range pixcol {
			newCol = append(newCol, pixel)
		}
		newPixels = append(newPixels, newCol)
	}
	images[name] = newPixels
	saveImagesToDisk()
}

func loadImageToCanvas(name string) {
	if _, ok := images[name]; ok {
		pLock.Lock()
		bounds := c.Bounds()
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				pixels[x][y] = images[name][x][y]
			}
		}
		//log.Printf("Pixel 0,0 color: %#v", pixels[0][0])
		pLock.Unlock()
		drawCanvas()
		return
	}
}

func deleteImage(name string) {
	delete(images, name)
}

func getImages() []string {
	buf := &bytes.Buffer{}
	m := 5
	bounds := c.Bounds()
	var imageCollection []string
	for name, p := range images {
		img := image.NewRGBA(image.Rect(0, 0, (bounds.Max.X*m)-1, (bounds.Max.Y*m)-1))

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for xx := x * m; xx < (x*m)+m; xx++ {
					for yy := y * m; yy < (y*m)+m; yy++ {
						img.Set(xx, yy, p[x][y])
					}
				}
			}
		}

		png.Encode(buf, img)
		imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
		img2html := "<div class='imgContainer card text-white bg-dark mb-3'><div class='card-header'><b style='font-size:28px;'>" + name + "</b><i class='fas fa-times fa-2x close-btn' onclick=\"deleteImage('" + name + "')\"></i></div><div class='card-body'><img src=\"data:image/png;base64," + imgBase64Str + "\" onclick=\"loadImageToCanvas('" + name + "')\" /></div></div>"
		imageCollection = append(imageCollection, img2html)
		buf.Reset()
	}
	return imageCollection
}

func saveCanvasAsAnimationFrame(name string, frameIndex int) {

	var newPixels [][]color.RGBA
	for _, pixcol := range pixels {
		var newCol []color.RGBA
		for _, pixel := range pixcol {
			newCol = append(newCol, pixel)
		}
		newPixels = append(newPixels, newCol)
	}
	animations[name][frameIndex] = newPixels

}

func playAnimationToCanvas(name string) {
	bounds := c.Bounds()
	for _, frame := range animations[name] {

		pLock.Lock()
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				pixels[x][y] = frame[x][y]
			}
		}
		pLock.Unlock()

		drawCanvas()
		//time.Sleep(time.Millisecond * 16)
		time.Sleep(time.Millisecond * 500)
	}
}
