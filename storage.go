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

func loadAnimationFrameToCanvas(name string, frame int) {
	if _, ok := animations[name]; ok {
		for i, aFrame := range animations[name] {
			if i == frame {
				pLock.Lock()
				bounds := c.Bounds()
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
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

func getAnimations() []string {
	buf := &bytes.Buffer{}
	m := 2
	bounds := c.Bounds()
	var animationCollection []string
	for name, animationFrames := range animations {

		var frames []string

		for i, animationFrame := range animationFrames {
			img := image.NewRGBA(image.Rect(0, 0, (bounds.Max.X*m)-1, (bounds.Max.Y*m)-1))

			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					for xx := x * m; xx < (x*m)+m; xx++ {
						for yy := y * m; yy < (y*m)+m; yy++ {
							img.Set(xx, yy, animationFrame[x][y])
						}
					}
				}
			}

			png.Encode(buf, img)
			imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
			frames = append(frames, fmt.Sprintf("<li id='frame-%d'><img class='animationFrame' src=\"data:image/png;base64,"+imgBase64Str+"\" onclick=\"loadAnimationFrameToCanvas('"+name+"',%d)\"/></li>", i, i))
			buf.Reset()
		}

		frameThumbnails := strings.Join(frames, "")

		img2html := fmt.Sprintf("<div class='animContainer card text-white bg-dark mb-3'><div class='card-header'><b style='font-size:28px;'>"+name+"</b><i class='fas fa-times fa-2x close-btn' onclick=\"deleteAnimation('"+name+"')\"></i></div><div class='card-body'><ul class='sortable' data-animation='"+name+"'>%s</ul></div><div class='card-footer'><div class='btn-group'><button class='btn btn-secondary' onclick=\"getAnimationEditor('"+name+"')\">Edit Animation <i class='fas fa-edit'></i></button><button class='btn btn-success' onclick=\"saveFrameToAnimation('"+name+"')\">Save Frame <i class='fas fa-save'></i></button></div> <button class='btn btn-success' onclick=\"playAnimation('"+name+"')\">Play <i class='fas fa-play'></i></button></div></div>", frameThumbnails)
		animationCollection = append(animationCollection, img2html)
		buf.Reset()

	}
	return animationCollection
}

func rearrangedAnimationFrames(w http.ResponseWriter, req *http.Request) {

	frameOrderStrSlice := req.Form["frame[]"]
	name := req.Form.Get("name")

	log.Printf("New order: %s", strings.Join(frameOrderStrSlice, ", "))

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

func getAnimationEditor(w http.ResponseWriter, req *http.Request) {

	output := &strings.Builder{}

	output.WriteString(`<table class='table table-striped table-condensed'><tr><th>Frame</th></tr>`)

	name := req.Form.Get("name")

	buf := &bytes.Buffer{}
	m := 5
	bounds := c.Bounds()
	for _, animationFrame := range animations[name] {

		img := image.NewRGBA(image.Rect(0, 0, (bounds.Max.X*m)-1, (bounds.Max.Y*m)-1))

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for xx := x * m; xx < (x*m)+m; xx++ {
					for yy := y * m; yy < (y*m)+m; yy++ {
						img.Set(xx, yy, animationFrame[x][y])
					}
				}
			}
		}

		png.Encode(buf, img)
		imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
		output.WriteString("<tr><td><img class='animationFrame' src=\"data:image/png;base64," + imgBase64Str + "\" /></td></tr>")
		buf.Reset()

		//img2html := fmt.Sprintf("<div class='animContainer card text-white bg-dark mb-3'><div class='card-header'><b style='font-size:28px;'>%s</b><i class='fas fa-times fa-2x close-btn' onclick=\"deleteAnimation('%s')\"></i></div><div class='card-body'>%s</div><div class='card-footer'><div class='btn-group'><button class='btn btn-secondary' onclick=\"editAnimation('%s')\">Edit Animation <i class='fas fa-edit'></i></button><button class='btn btn-success' onclick=\"saveFrameToAnimation('%s')\">Save Frame <i class='fas fa-save'></i></button></div> <button class='btn btn-success' onclick=\"playAnimation('%s')\">Play <i class='fas fa-play'></i></button></div></div>", name, name, frameThumbnails, name, name, name)
		//frames = append(frames, img2html)

	}

	output.WriteString("</table>")
	w.Write([]byte(output.String()))
}

func playAnimationToCanvas(name string, loops int) {
	bounds := c.Bounds()
	for i := 0; i < loops; i++ {
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

}
