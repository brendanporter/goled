package main

import (
	"os"
	"os/signal"
	//"github.com/stianeikeland/go-rpio"
	//"github.com/mcuadros/go-rpi-rgb-led-matrix"
	"github.com/brendanporter/go-rpi-rgb-led-matrix"
	//"image"
	"image/color"
	//"image/draw"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//"path/filepath"
	"flag"
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

func main() {

	var cols int
	var rows int
	var gpioSlowdown int

	flag.IntVar(&cols, "cols", 32, "LED Columns in matrix")
	flag.IntVar(&rows, "rows", 32, "LED Rows in matrix")
	flag.IntVar(&gpioSlowdown, "slowdown-gpio", 1, "LED GPIO Slowdown")

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
	argC := len(os.Args)
	argV := os.Args
	m, err := rgbmatrix.NewRGBLedMatrix(config, &argC, &argV)
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
		Addr:         ":80",
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Print(err)
	}
}

type Pixel struct {
	X, Y       int
	R, G, B, A uint8
}

var canvasSerial int

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
	case "clearDisplay":
		clearDisplay()
		getDisplay(w, req)
		break
	case "getDisplay":
		getDisplay(w, req)
		break
	case "saveCanvasAsImage":
		saveCanvasAsImage()
		w.WriteHeader(http.StatusNoContent)
		break
	case "loadImageToCanvas":

		index, err := strconv.Atoi(req.Form.Get("index"))
		if err != nil {
			log.Print(err)
			break
		}
		loadImageToCanvas(index)
		w.WriteHeader(http.StatusNoContent)
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
		log.Print(err)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if canvasSerial != clientCanvasSerial {

		pxResponse := PXResponse{
			Canvas:       pixels,
			CanvasSerial: canvasSerial,
		}

		pLock.Lock()
		p, err := json.Marshal(pxResponse)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		pLock.Unlock()
		w.Write(p)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func clearDisplay() {
	pLock.Lock()
	bounds := c.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pixels[x][y] = color.RGBA{0, 0, 0, 0}
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
			return
		}

		w.Write(fileBytes)
		return
	}

	var buttons string
	bounds := c.Bounds()
	buttons += "<tr><td></td>"
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		buttons += fmt.Sprintf("<td>%d</td>", x)
	}
	buttons += fmt.Sprintf("</tr>")

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		buttons += fmt.Sprintf("<tr><td>%d</td>", y)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			buttons += fmt.Sprintf("<td class='pixel' onmouseover='hoverPixel(%d,%d)' onclick='setPixel(%d,%d)'> </td>", x, y, x, y)
		}
		buttons += fmt.Sprintf("</tr>")
	}

	h := fmt.Sprintf(`<!DOCTYPE html>
	<html>
	<head>
	<title>Sign</title>
	<script src="https://code.jquery.com/jquery-2.2.4.min.js"
			  integrity="sha256-BbhdlvQf/xTY9gja0Dq3HiwQF8LaCRTXxZKRutelT44="
			  crossorigin="anonymous"></script>
	<link href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-MCw98/SFnGE8fJT3GXwEOngsV7Zt27NXFoaoApmYm81iuXoPkFOJwJ8ERdknLPMO" crossorigin="anonymous">
	<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/js/bootstrap.min.js" integrity="sha384-ChfqqxuZUCnJSK3+MXmPNIyE6ZbWh2IMqE241rYiqJxyMiZ6OW/JmZQ5stwEULTy" crossorigin="anonymous"></script>
	
	<script src="https://cdnjs.cloudflare.com/ajax/libs/spectrum/1.8.0/spectrum.min.js"></script>
	<link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/spectrum/1.8.0/spectrum.min.css">

	<script type='text/javascript'>

	color = {}
	color.R = 0;
	color.G = 0;
	color.B = 0;
	color.A = 0;

	drawmode = false;
	canvasSerial = 0;

	$(document).ready(function(){
		init();
	});


	function init() {
		$(document).keyup(function(event) {
  			  if ( event.which == 68 ) {
			    event.preventDefault();
			    drawmode = !drawmode;
			  }
		});

		$('.pixel').on('mousedown', function(event) {
			event.preventDefault();
		    drawmode = true;
		});

		$('.pixel').on('contextmenu', function(event){
			event.preventDefault();
			dropper = $(event.target).css('background-color');
			$("#color").spectrum('set', dropper);
			setColor();


		});

		$(document).on('mouseup', function(event) {
		    drawmode = false;
		});

		$.fn.spectrum.load = false;

		$("#color").spectrum({
		    change: function(color){
		    	console.log(color);
		    	setColor();
		    },
		});

		refreshDisplayFromServer();

		setInterval(refreshDisplayFromServer, 2000);
	}

	function setColor(){
		p = $('#color').spectrum('get').toRgb();
		color.R = p.r;
		color.G = p.g;
		color.B = p.b;
		color.A = 255;
	}

	function hoverPixel(x,y){
		if(drawmode) {
			setPixel(x,y);
		}
	}

	function saveCanvasAsImage() {
		$.ajax({
		url: "/api?action=saveCanvasAsImage",
		type: 'post',
		dataType: 'json',
		beforeSend: function(){
		},
		success: function(json){
		}
		});
	}

	function loadImageToCanvas(index) {
		$.ajax({
		url: "/api?action=loadImageToCanvas&index=" + index,
		type: 'post',
		dataType: 'json',
		beforeSend: function(){
		},
		success: function(json){
		}
		});
	}

	function clearDisplay(){
		$.ajax({
		url: "/api?action=clearDisplay&canvasSerial=" + canvasSerial,
		type: 'post',
		dataType: 'json',
		beforeSend: function(){
		},
		success: function(json){
			$.each(json, function(i,col){
				$.each(col, function(j, px){
					td = i +2;
					tr = j +2;
					$('#pixelTable tr:nth-child('+tr+') td:nth-child('+td+')').css('background-color','rgba('+px.R+','+px.G+','+px.B+',255)');
				});
			});
		}
		});
	}

	function refreshDisplayFromServer(){


		if(!drawmode){

		$.ajax({
		url: "/api?action=getDisplay&canvasSerial=" + canvasSerial,
		type: 'post',
		dataType: 'json',
		beforeSend: function(){
		},
		success: function(json){

			if(typeof json == "undefined"){
				return
			}

			canvasSerial = json.CanvasSerial;

			$.each(json.Canvas, function(i,col){
				$.each(col, function(j, px){
					td = i +2;
					tr = j +2;
					$('#pixelTable tr:nth-child('+tr+') td:nth-child('+td+')').css('background-color','rgba('+px.R+','+px.G+','+px.B+',255)');
				});
			});
		}
		});

		}
	}

	function setPixel(x,y){

		tr = y +2;
		td = x +2;
		$('#pixelTable tr:nth-child('+tr+') td:nth-child('+td+')').css('background-color','rgba('+color.R+','+color.G+','+color.B+','+color.A+')');

		pixels = []
		px = {}
		px.X = x;
		px.Y = y;
		px.R = color.R;
		px.G = color.G;
		px.B = color.B;
		px.A = 255;
		pixels.push(px);

		pxJSON = JSON.stringify(pixels);

		console.log(pxJSON);

		$.ajax({
		url: "/api?action=setPixel",
		type: 'post',
		data: {px: pxJSON},
		dataType: 'json',
		beforeSend: function(){
		},
		success: function(json){

		}
		});
	}
	</script>
	<style>
	td {padding: 0px !important; min-width: 25px;}
	.pixel {background-color: black;}
	#pixelTable {position: absolute; top:40px;}
	#clear {position: absolute; right:0px;}
	</style>
	</head>
	<body>
	<input id='color' type='color' />
	<button class='pallette btn btn-danger' onclick="$('#color').spectrum('set', 'rgb(255,0,0)');setColor();">Red</button>
	<button class='pallette btn btn-success' onclick="$('#color').spectrum('set', 'rgb(0,255,0)');setColor();">Green</button>
	<button class='pallette btn btn-primary' onclick="$('#color').spectrum('set', 'rgb(0,0,255)');setColor();"">Blue</button>
	<button id='clear' class='btn btn-danger' onclick='clearDisplay()'>Clear</button>

	<table id='pixelTable' class='table table-striped table-bordered table-condensed'>%s</table>
	</body>
	</html>`, buttons)

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
