package main

import (
	"net"
	"os"
	"os/signal"

	//"github.com/stianeikeland/go-rpio"
	//"github.com/mcuadros/go-rpi-rgb-led-matrix"
	rgbmatrix "github.com/brendanporter/go-rpi-rgb-led-matrix"
	//"image"
	"image/color"
	//"image/draw"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Matrix color pins to GPIO pins
// Top half
const mr1 = 5
const mg1 = 13
const mb1 = 6

// Bottom half
const mr2 = 12
const mg2 = 16
const mb2 = 23

// Matrix control pins to GPIO pins
const moe = 4   // This pin controls whether the LEDs are lit at all
const mclk = 17 // High speed clock pin for clocking RGB data to the matrix
const mlat = 21 // Data latching pin for clocking RGB data to the matrix

// Matrix address pins to GPIO pins
const ma = 22 // This pin is part of the 1->32, 1->16 or 1->8 multiplexing circuitry.
const mb = 26 // This pin is part of the 1->32, 1->16 or 1->8 multiplexing circuitry.
const mc = 27 // This pin is part of the 1->32, 1->16 or 1->8 multiplexing circuitry.
const md = 20 // This pin is part of the 1->32, 1->16 multiplexing circuitry. Used for 32-pixel and 64-pixel tall displays only
const me = 24 // This pin is part of the 1->32 multiplexing circuitry. Used for 64-pixel tall displays only

var c *rgbmatrix.Canvas

func init() {
	elog = log.New(os.Stdout, "Error: ", log.LstdFlags|log.Lshortfile)
}

// Web UI gets slow sometimes waiting for updates.

func main() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for sig := range sigChan {
			// sig is a ^C, handle it

			log.Printf("GOLED shutting down for SIGINT: %v", sig)
			os.Exit(0)
		}
	}()

	loadImagesFromDisk()
	loadAnimationsFromDisk()

	var cols int
	var rows int
	var gpioSlowdown int
	var addrType int
	var multiplexing int

	flag.IntVar(&cols, "cols", 32, "LED Columns in matrix")
	flag.IntVar(&rows, "rows", 32, "LED Rows in matrix")
	flag.IntVar(&gpioSlowdown, "led-slowdown-gpio", 1, "LED GPIO Slowdown")
	flag.IntVar(&addrType, "led-row-addr-type", 0, "LED Address Type")
	flag.IntVar(&multiplexing, "led-multiplexing", 0, "LED Multiplexing")

	flag.Parse()

	if cols < 16 || rows < 16 {
		cols = 16
		rows = 16
		log.Print("Values for cols or rows was too small")
	}

	chanSig := make(chan os.Signal)
	signal.Notify(chanSig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-chanSig
		c.Close()
		os.Exit(1)
	}()

	config := &rgbmatrix.DefaultConfig
	config.Rows = rows
	config.Cols = cols
	config.HardwareMapping = "adafruit-hat-pwm"
	config.DisableHardwarePulsing = false
	//config.ShowRefreshRate = true
	m, err := rgbmatrix.NewRGBLedMatrix(config)
	if err != nil {
		log.Print(err)
	}

	c = rgbmatrix.NewCanvas(m)
	defer c.Close()

	bounds := c.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		var ySlice []color.RGBA
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			ySlice = append(ySlice, color.RGBA{0, 0, 0, 255})
		}
		pixels = append(pixels, ySlice)
	}

	//draw.Draw(c, c.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)

	//c.Render()

	//time.Sleep(time.Second * 5)
	//c.Clear()
	/*
		go cylon(color.RGBA{255, 0, 0, 255})
		time.Sleep(time.Millisecond * (20 * 30))
		go cylon(color.RGBA{0, 255, 0, 255})
		time.Sleep(time.Millisecond * (20 * 60))
		go cylon(color.RGBA{0, 0, 255, 125})
	*/
	go square()

	http.HandleFunc("/", baseHandler)
	http.HandleFunc("/api", apiHandler)

	server := http.Server{
		Handler:      nil,
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
	}

	log.Print("Starting web server")

	l, err := net.Listen("tcp4", "0.0.0.0:80")
	if err != nil {
		log.Panic(err)
	}

	err = server.Serve(l)
	if err != nil {
		log.Print(err)
	}
}

type Pixel struct {
	X, Y       int
	R, G, B, A uint8
}

var canvasSerial int

var elog *log.Logger

var pixels [][]color.RGBA // [X][Y]Pixel
var pLock sync.Mutex

func setPixel(w http.ResponseWriter, req *http.Request) {
	pxJSON := req.Form.Get("px")

	var px []Pixel
	err := json.Unmarshal([]byte(pxJSON), &px)
	if err != nil {
		log.Print(err)
		return
	}

	pLock.Lock()
	for _, p := range px {
		pixels[p.X][p.Y] = color.RGBA{p.R, p.G, p.B, p.A}
	}
	pLock.Unlock()
	drawCanvas()
}

func setPixels(w http.ResponseWriter, req *http.Request) {
	pxJSON := req.Form.Get("px")

	var px []Pixel
	err := json.Unmarshal([]byte(pxJSON), &px)
	if err != nil {
		log.Print(err)
		return
	}

	pLock.Lock()
	for _, p := range px {
		pixels[p.X][p.Y] = color.RGBA{p.R, p.G, p.B, p.A}
	}
	pLock.Unlock()
	drawCanvas()
}

func drawCanvas() {
	pLock.Lock()
	bounds := c.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c.Set(x, y, pixels[x][y])
		}
	}
	pLock.Unlock()
	c.Render()
	canvasSerial++
}

func apiHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	action := req.Form.Get("action")

	switch action {
	case "test":
		log.Printf("Running %s", action)
		go cylon(color.RGBA{255, 255, 255, 255}, time.Now().Add(time.Second*10))
		break
	case "setPixel":
		setPixel(w, req)
		w.WriteHeader(http.StatusNoContent)
		break
	case "setPixels":
		setPixels(w, req)
		w.WriteHeader(http.StatusNoContent)
		break
	case "clearDisplay":
		clearDisplay()
		getDisplay(w, req)
		break
	case "getDisplay":
		getDisplay(w, req)
		break
	case "deleteAnimationFrames":
		deleteAnimationFrames(w, req)
		break
	case "deleteAnimation":
		name := req.Form.Get("name")
		deleteAnimation(name)
		w.WriteHeader(http.StatusNoContent)
		break
	case "rearrangedAnimationFrames":
		rearrangedAnimationFrames(w, req)
		break
	case "saveNewAnimation":
		name := req.Form.Get("name")
		saveNewAnimation(name)
		w.WriteHeader(http.StatusNoContent)
		break
	case "saveFrameToAnimation":
		name := req.Form.Get("name")
		saveFrameToAnimation(name)
		w.WriteHeader(http.StatusNoContent)
		break
	case "getAnimations":
		imageHTMLSlice := getAnimations()

		w.Write([]byte(strings.Join(imageHTMLSlice, "")))
		break
	case "playAnimation":
		loopsStr := req.Form.Get("loops")
		loops, err := strconv.Atoi(loopsStr)
		if err != nil {
			loops = 3
		}

		if loops < 1 {
			loops = 3
		}

		name := req.Form.Get("name")
		if name == "" {
			w.WriteHeader(http.StatusNoContent)
			break
		}
		playAnimationToCanvas(name, loops)
	case "deleteImage":
		name := req.Form.Get("name")
		if name == "" {
			w.WriteHeader(http.StatusNoContent)
			break
		}
		deleteImage(name)
		w.WriteHeader(http.StatusNoContent)
		break
	case "getImages":
		imageHTMLSlice := getImages()

		w.Write([]byte(strings.Join(imageHTMLSlice, "")))
		break
	case "saveCanvasAsImage":
		name := req.Form.Get("name")
		saveCanvasAsImage(name)
		w.WriteHeader(http.StatusNoContent)
		break
	case "loadAnimationFrameToCanvas":
		frameStr := req.Form.Get("frame")
		frame, err := strconv.Atoi(frameStr)
		if err != nil {
			elog.Print(err)
		}
		name := req.Form.Get("name")
		loadAnimationFrameToCanvas(name, frame)
		w.WriteHeader(http.StatusNoContent)
		break
	case "loadImageToCanvas":

		name := req.Form.Get("name")
		loadImageToCanvas(name)
		getDisplay(w, req)
	default:
		log.Printf("Unknown API requested: %s", action)
		break
	}
}

type PXResponse struct {
	CanvasSerial int
	Canvas       [][]color.RGBA
}

func getDisplay(w http.ResponseWriter, req *http.Request) {
	clientCanvasSerialStr := req.Form.Get("canvasSerial")
	clientCanvasSerial, err := strconv.Atoi(clientCanvasSerialStr)
	if err != nil {
		elog.Print(err)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	pLock.Lock()

	if canvasSerial != clientCanvasSerial {

		pxResponse := PXResponse{
			Canvas:       pixels,
			CanvasSerial: canvasSerial,
		}

		p, err := json.Marshal(pxResponse)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(p)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

	pLock.Unlock()
}

func clearDisplay() {
	pLock.Lock()
	bounds := c.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pixels[x][y] = color.RGBA{0, 0, 0, 255}
			c.Set(x, y, pixels[x][y])
		}
	}
	pLock.Unlock()
	c.Render()
	canvasSerial++
}

func displayLTCPrice() {

	hc := http.Client{}

	resp, err := hc.Get("https://api.coinbase.com/v2/prices/LTC-USD/spot")
	if err != nil {
		log.Print(err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	type coinbaseSpotPrice struct {
		Data struct {
			Base     string
			Currency string
			Amount   string
		}
	}

	cbsp := coinbaseSpotPrice{}

	err = json.Unmarshal(bodyBytes, &cbsp)
	if err != nil {
		log.Print(err)
	}

	log.Print("Litecoin Price: %s", cbsp.Data.Amount)

}

func baseHandler(w http.ResponseWriter, req *http.Request) {

	req.ParseForm()

	filePath := req.URL.Path

	if filePath != "/" && filePath != "" {

		if strings.HasPrefix(filePath, "/") {
			filePath = filePath[1:]
		}

		log.Printf("Someone requested file: %#v", filePath)

		//absPath, _ := filepath.Abs("contactDNA.html")
		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write(fileBytes)
		return
	}

	buttons := &strings.Builder{}
	bounds := c.Bounds()
	buttons.WriteString("<tr><td></td>")
	for x := bounds.Min.X; x < bounds.Max.X; x++ {

		if x%5 == 0 {
			buttons.WriteString("<td class='marker'> </td>")
		} else {
			buttons.WriteString("<td> </td>")
		}
	}
	buttons.WriteString("</tr>")

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		if y%5 == 0 {
			buttons.WriteString("<tr><td class='marker'> </td>")
		} else {
			buttons.WriteString("<tr><td> </td>")
		}
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			buttons.WriteString(fmt.Sprintf("<td id='px%02d%02d' class='pixel' data-x='%d' data-y='%d' onmouseover='hoverPixel(%d,%d)' onclick='setPixel(%d,%d)'></td>", x, y, x, y, x, y, x, y))
		}
		buttons.WriteString(fmt.Sprintf("</tr>"))
	}

	index, err := filepath.Abs("index.html")
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fileBytes, err := ioutil.ReadFile(index)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h := fmt.Sprintf(string(fileBytes), buttons.String())

	w.Write([]byte(h))
}

func square() {
	c.Set(7, 14, color.RGBA{0, 0, 255, 255})
	c.Set(7, 15, color.RGBA{0, 255, 0, 255})
	c.Render()
}

func cylon(clr color.RGBA, timeout time.Time) {
	bounds := c.Bounds()
	frame := time.Millisecond * 20
	for time.Now().Before(timeout) {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.Set(x, y, clr)
			}
			c.Render()
			time.Sleep(frame)
		}

		for x := bounds.Max.X - 1; x > bounds.Min.X; x-- {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.Set(x, y, clr)
			}
			c.Render()
			time.Sleep(frame)
		}
	}
	c.Clear()
}
