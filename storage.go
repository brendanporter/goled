package main

import (
	"image/color"
	"image/png"
	//"log"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var animations map[string][][][]color.RGBA // [Many][Frames][X][Y]color

var images map[string][][]color.RGBA // [Many][X][Y]color

func init() {
	images = make(map[string][][]color.RGBA)
	animations = make(map[string][][][]color.RGBA)
}

func saveAnimationsToDisk() {
	absPath, err := filepath.Abs("animations.json")
	if err != nil {
		elog.Print(err)
	}

	animationsJSON, err := json.Marshal(animations)
	if err != nil {
		elog.Print(err)
		return
	}

	err = ioutil.WriteFile(absPath, animationsJSON, 0755)
	if err != nil {
		elog.Print(err)
	}

	log.Printf("Saved %d animations", len(animations))
}

func loadAnimationsFromDisk() {
	absPath, err := filepath.Abs("animations.json")
	if err != nil {
		elog.Print(err)
	}

	fileBytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		elog.Print(err)
		return
	}

	err = json.Unmarshal(fileBytes, &animations)
	if err != nil {
		elog.Print(err)
	}

	log.Printf("Loaded %d animations from disk", len(animations))

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

	log.Printf("Saved %d images", len(images))
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
	clearDisplay()
	if _, ok := images[name]; ok {
		pLock.Lock()
		bounds := c.Bounds()
		for x := bounds.Min.X; x < bounds.Max.X && x < len(images[name]); x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y && y < len(images[name][0]); y++ {
				pixels[x][y] = images[name][x][y]
			}
		}
		//log.Printf("Pixel 0,0 color: %#v", pixels[0][0])
		pLock.Unlock()
		drawCanvas()
		return
	}
}

func loadAnimationFrameToCanvas(name string, frame int) {
	clearDisplay()
	if _, ok := animations[name]; ok {
		for i, aFrame := range animations[name] {
			if i == frame {
				pLock.Lock()
				bounds := c.Bounds()
				for x := bounds.Min.X; x < bounds.Max.X && x < len(aFrame); x++ {
					for y := bounds.Min.Y; y < bounds.Max.Y && y < len(aFrame[0]); y++ {
						pixels[x][y] = aFrame[x][y]
					}
				}
				pLock.Unlock()
				drawCanvas()
				return
			}
		}
	}
}

func deleteImage(name string) {
	delete(images, name)
	saveImagesToDisk()
}

func getImages() []string {
	buf := &bytes.Buffer{}
	m := 1
	bounds := c.Bounds()
	var imageCollection []string
	for name, p := range images {
		img := image.NewRGBA(image.Rect(0, 0, (bounds.Max.X*m)-1, (bounds.Max.Y*m)-1))

		imgMaxX := len(p)
		imgMaxY := len(p[0])

		for x := bounds.Min.X; x < bounds.Max.X && x < imgMaxX; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y && y < imgMaxY; y++ {
				for xx := x * m; xx < (x*m)+m; xx++ {
					for yy := y * m; yy < (y*m)+m; yy++ {
						img.Set(xx, yy, p[x][y])
					}
				}
			}
		}

		png.Encode(buf, img)
		imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
		img2html := "<div class='imgContainer card text-white bg-dark mb-3'><div class='card-header'><b>" + name + "</b><i class='fas fa-times close-btn' onclick=\"deleteImage('" + name + "')\"></i></div><div class='card-body'><img src=\"data:image/png;base64," + imgBase64Str + "\" onclick=\"loadImageToCanvas('" + name + "')\" /></div></div>"
		imageCollection = append(imageCollection, img2html)
		buf.Reset()
	}
	return imageCollection
}

func saveNewAnimation(name string) {
	animations[name] = [][][]color.RGBA{}
	saveAnimationsToDisk()
}

func saveFrameToAnimation(name string) {
	var newPixels [][]color.RGBA
	for _, pixcol := range pixels {
		var newCol []color.RGBA
		for _, pixel := range pixcol {
			newCol = append(newCol, pixel)
		}
		newPixels = append(newPixels, newCol)
	}
	animations[name] = append(animations[name], newPixels)
	saveAnimationsToDisk()
}

func deleteAnimation(name string) {
	delete(animations, name)
	saveAnimationsToDisk()
}

func deleteAnimationFrames(w http.ResponseWriter, req *http.Request) {
	framesStrSlice := req.Form["frames[]"]
	name := req.Form.Get("name")

	log.Printf("Deleted: %s", strings.Join(framesStrSlice, ", "))
	var deletable []int
	for _, v := range framesStrSlice {
		d, err := strconv.Atoi(v)
		if err != nil {
			elog.Print(err)
			return
		}
		deletable = append(deletable, d)
	}

	for i := len(animations[name]) - 1; i >= 0; i-- {

		for j := len(deletable) - 1; j >= 0; j-- {
			if deletable[j] == i {
				animations[name] = append(animations[name][:i], animations[name][i+1:]...)
			}
		}

	}
	saveAnimationsToDisk()
	w.WriteHeader(http.StatusOK)
}

func getAnimations() []string {
	buf := &bytes.Buffer{}
	m := 1
	bounds := c.Bounds()
	var animationCollection []string
	for name, animationFrames := range animations {

		var frames []string

		for i, animationFrame := range animationFrames {
			img := image.NewRGBA(image.Rect(0, 0, (bounds.Max.X*m)-1, (bounds.Max.Y*m)-1))

			imgMaxX := len(animationFrame)
			imgMaxY := len(animationFrame[0])

			for x := bounds.Min.X; x < bounds.Max.X && x < imgMaxX; x++ {
				for y := bounds.Min.Y; y < bounds.Max.Y && y < imgMaxY; y++ {
					for xx := x * m; xx < (x*m)+m; xx++ {
						for yy := y * m; yy < (y*m)+m; yy++ {
							img.Set(xx, yy, animationFrame[x][y])
						}
					}
				}
			}

			png.Encode(buf, img)
			imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
			frames = append(frames, fmt.Sprintf("<li id='frame-%d'><input data-frame='%d' class='imageSelector' type='checkbox' /><img class='animationFrame' src=\"data:image/png;base64,%s\" onclick=\"loadAnimationFrameToCanvas('%s',%d)\"/></li>", i, i, imgBase64Str, name, i))
			buf.Reset()
		}

		frameThumbnails := strings.Join(frames, "")

		img2html := fmt.Sprintf(`
			<div class='animContainer card text-white bg-dark mb-3'>
				<div class='card-header'>
					<b'>%s</b><i class='fas fa-times close-btn' onclick="deleteAnimation('%s')"></i>
				</div>
				<div class='card-body'>
					<ul class='sortable' data-animation='%s'>%s</ul>
				</div>
				<div class='card-footer'>
					<button class='btn btn-sm btn-danger' onclick="deleteAnimationFrames('%s')">Delete Selected</button>
					<button class='btn btn-sm btn-success' onclick="saveFrameToAnimation('%s')">Save Frame <i class='fas fa-save'></i></button> 
					<button class='btn btn-sm btn-success' onclick="playAnimation('%s')">Play <i class='fas fa-play'></i></button>
				</div>
			</div>`, name, name, name, frameThumbnails, name, name, name)
		animationCollection = append(animationCollection, img2html)
		buf.Reset()

	}
	return animationCollection
}

func rearrangedAnimationFrames(w http.ResponseWriter, req *http.Request) {

	frameOrderStrSlice := req.Form["frame[]"]
	name := req.Form.Get("name")

	log.Printf("New order for %s: %s", name, strings.Join(frameOrderStrSlice, ", "))

	var newAnimationsFrames [][][]color.RGBA
	for _, v := range frameOrderStrSlice {
		newIndex, err := strconv.Atoi(v)
		if err != nil {
			elog.Print(err)
			return
		}

		newAnimationsFrames = append(newAnimationsFrames, animations[name][newIndex])
	}

	animations[name] = newAnimationsFrames
	saveAnimationsToDisk()
	w.WriteHeader(http.StatusOK)
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

func playAnimationToCanvas(name string, loops int) {
	clearDisplay()
	bounds := c.Bounds()
	for i := 0; i < loops; i++ {
		for _, frame := range animations[name] {

			pLock.Lock()
			for x := bounds.Min.X; x < bounds.Max.X && x < len(frame); x++ {
				for y := bounds.Min.Y; y < bounds.Max.Y && y < len(frame[0]); y++ {
					pixels[x][y] = frame[x][y]
				}
			}
			pLock.Unlock()

			drawCanvas()
			time.Sleep(time.Millisecond * 16)
			//time.Sleep(time.Millisecond * 500)
		}
	}

}
