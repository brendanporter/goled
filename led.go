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

func init() {
	elog = log.New(os.Stdout, "Error: ", log.LstdFlags|log.Lshortfile)
}

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

			buttons.WriteString(fmt.Sprintf("<td class='pixel' data-x='%d' data-y='%d' onmouseover='hoverPixel(%d,%d)' onclick='setPixel(%d,%d)'></td>", x, y, x, y, x, y))
		}
		buttons.WriteString(fmt.Sprintf("</tr>"))
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
	<link rel="stylesheet" href="https://use.fontawesome.com/releases/v5.2.0/css/all.css" integrity="sha384-hWVjflwFxL6sNzntih27bfxkr27PmbbK/iSvJ+a4+0owXq79v+lsFkW54bOGbiDQ" crossorigin="anonymous">

	<script src="https://code.jquery.com/ui/1.12.1/jquery-ui.min.js"></script>

	<link rel="stylesheet" type="text/css" href="https://code.jquery.com/ui/1.12.1/themes/base/jquery-ui.css">
	

	<script type='text/javascript'>

	color = {}
	color.R = 0;
	color.G = 0;
	color.B = 0;
	color.A = 0;

	fill = false;

	drawmode = false;
	canvasSerial = 0;

	$(document).ready(function(){
		init();
	});


	function init() {


		$('.pixel').on('mousedown', function(event) {
			if(event.button == 0){
				event.target.onclick.apply(event);
			}
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

		getImages()
		getAnimations()

		

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

		var name = prompt('Please name the image:')
		if (name === "") {
		    // user pressed OK, but the input field was empty
		    return false;
		} else if (name) {
		    // user typed something and hit OK
		} else {
		    // user hit cancel
		    return false;
		}

		if(name === ""){
			return false;
		}

		$.ajax({
		url: "/api?action=saveCanvasAsImage",
		type: 'post',
		dataType: 'json',
		data: {name: name},
		beforeSend: function(){
		},
		success: function(json){
			getImages()
		}
		});
	}

	function playAnimation(name) {
		
		var loops = prompt('How many loops?')
		if (loops === "") {
		    // user pressed OK, but the input field was empty
		    loops = 3;
		} else if (loops) {
		    // user typed something and hit OK
		} else {
		    // user hit cancel
		    return false;
		}


		$.ajax({
		url: "/api?action=playAnimation",
		type: 'post',
		dataType: 'json',
		data: {name: name, loops: loops},
		beforeSend: function(){
		},
		success: function(json){
			clearDisplay();
		}
		});
	}

	function newAnimation() {
		var name = prompt('Please name the animation:')
		if (name === "") {
		    // user pressed OK, but the input field was empty
		    return false;
		} else if (name) {
		    // user typed something and hit OK
		} else {
		    // user hit cancel
		    return false;
		}

		if(name === ""){
			return false;
		}

		$.ajax({
		url: "/api?action=saveNewAnimation",
		type: 'post',
		dataType: 'json',
		data: {name: name},
		beforeSend: function(){
		},
		success: function(json){
			getAnimations();
		}
		});
	}

	function deleteAnimationFrames(name) {
		frames = [];
		$('#animations .card-header b:contains("'+name+'")').filter(function(){return $(this).text() === name;}).parent().parent().find('input:checked').each(function(i,e){
			frames.push($(e).data('frame'));
		});

		$.ajax({
		url: "/api?action=deleteAnimationFrames",
		type: 'post',
		data: {name: name, frames: frames},
		beforeSend: function(){
		},
		success: function(){
			getAnimations();
		}
		});
	}

	function saveFrameToAnimation(name){
		$.ajax({
		url: "/api?action=saveFrameToAnimation",
		type: 'post',
		dataType: 'json',
		data: {name: name},
		beforeSend: function(){
		},
		success: function(json){
			getAnimations();
		}
		});
	}


	function getFramesForAnimation(name) {
		$.ajax({
			url: "/api?action=getAnimations",
			type: 'post',
			dataType: 'html',
			data: {canvasSerial: canvasSerial},
			beforeSend: function(){
				$('#animations').html('')
			},
			success: function(html){

			}
		});
	}


	function getAnimations() {
		$.ajax({
		url: "/api?action=getAnimations",
		type: 'post',
		dataType: 'html',
		data: {canvasSerial: canvasSerial},
		beforeSend: function(){
			$('#animations').html('')
		},
		success: function(html){
			$('#animations').html(html)
			$('.sortable').sortable();
			//$('.sortable').disableSelection();
			$('.sortable').on('sortstop', function(event,ui){
				name = $(event.target).data('animation');
				frames = $(event.target).sortable("serialize");
				$.ajax({
					url: "/api?action=rearrangedAnimationFrames&" + frames,
					type: 'post',
					dataType: 'html',
					data: {name: name},
					beforeSend: function(){
					},
					success: function(html){
						console.log('Retrieved new frames');
						$(event.target).html(html);
					}
				});
			});
		}
		});
	}

	function deleteAnimation(name) {

		if(!confirm("Delete image '" + name + "'?")) {
			return false;	
		}

		$.ajax({
		url: "/api?action=deleteAnimation",
		type: 'post',
		dataType: 'json',
		data: {name: name, canvasSerial: canvasSerial},
		beforeSend: function(){
		},
		success: function(){
			getAnimations()
		}
		});
	}

	function deleteImage(name) {

		if(!confirm("Delete image '" + name + "'?")) {
			return false;	
		}

		$.ajax({
		url: "/api?action=deleteImage",
		type: 'post',
		dataType: 'json',
		data: {name: name, canvasSerial: canvasSerial},
		beforeSend: function(){
		},
		success: function(){
			getImages()
		}
		});
	}

	function loadImageToCanvas(name) {
		$.ajax({
		url: "/api?action=loadImageToCanvas",
		type: 'post',
		dataType: 'json',
		data: {name: name, canvasSerial: canvasSerial},
		beforeSend: function(){
		},
		success: function(json){
		}
		});
	}

	function loadAnimationFrameToCanvas(name, frame) {
		$.ajax({
		url: "/api?action=loadAnimationFrameToCanvas",
		type: 'post',
		dataType: 'json',
		data: {name: name, frame: frame},
		beforeSend: function(){
		},
		success: function(json){
		}
		});
	}

	function getImages() {
		$.ajax({
		url: "/api?action=getImages",
		type: 'post',
		dataType: 'html',
		data: {canvasSerial: canvasSerial},
		beforeSend: function(){
			$('#images').html('')
		},
		success: function(html){
			$('#images').html(html)
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

	function toggleDrawer(e){
		target = $(e).data('target');
		$('#' + target).toggle();
		$(e).find('.fas').each(function(i,el){
				$(el).toggle();
			})
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

	function isNeighbor(pixelArray, target) {
		$.each(pixelArray, function(j,paPx){
				console.log("Testing if " + target.X + " is neighbor of " + paPx.X);
				if(Math.abs(target.X - paPx.X) == 1 || Math.abs(target.Y - paPx.Y) == 1){
					return true;
				}
		});
		return false;
	}

	function setPixel(x,y){

		tr = y +2;
		td = x +2;
		target = $('#pixelTable tr:nth-child('+tr+') td:nth-child('+td+')');
		targetColor = $(target).css('background-color');


		pixels = []
		px = {}
		px.X = x;
		px.Y = y;
		px.R = color.R;
		px.G = color.G;
		px.B = color.B;
		px.A = 255;
		pixels.push(px);

		

		if(fill === true){
			// Find all other pixels currently matching the target pixel color and add to array to be changed

			// Radius is unlimited, but only seeking pixels that share borders, not diagonals

			// Find all TDs on the same row that match color, then look up and down

			/*
			var added = 1;
			while(added > 0){
				added = 0;
			*/

				for(var yy = tr; yy >= 0; y--){
					rowCandidates = $(target).siblings().filter('td[style="background-color:'+targetColor+';"]');
					$.each(rowCandidates, function(i,v){
						if(isNeighbor(pixels, {X: $(v).data('x'), Y: $(v).data('y')})){
							pixels.push({X: $(v).data('x'), Y: $(v).data('y'), R: color.R, G: color.G, B: color.B, A: 255});
							added++;
							console.log("Added " + $(v).data('x') + ", " + $(v).data('y'));
						}
					});
				}

				for(var yy = tr; yy < $('#pixelTable tr').length - 1; y++){
					rowCandidates = $(target).siblings().filter('td[style="background-color:'+targetColor+';"]');
					$.each(rowCandidates, function(i,v){
						if(isNeighbor(pixels, {X: $(v).data('x'), Y: $(v).data('y')})){
							pixels.push({X: $(v).data('x'), Y: $(v).data('y'), R: color.R, G: color.G, B: color.B, A: 255});
							added++;
							console.log("Added " + $(v).data('x') + ", " + $(v).data('y'));
						}
					});
				}

			//}
		
		}

		pxJSON = JSON.stringify(pixels);

		// Set the target to the specified color
		$('#pixelTable tr:nth-child('+tr+') td:nth-child('+td+')').css('background-color','rgba('+color.R+','+color.G+','+color.B+','+color.A+')');




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
	td {padding: 0px !important;}
	.pixel {background-color: black;}
	#pixelTable {position: absolute; top:40px; width:initial;}
	#pixelTable tr td {width:9px; height:9px;}
	#clear {position: absolute; right:0px;}
	.marker {background-color: black;}
	#storage {position:absolute; bottom:0px; left:0px; right:0px;}
	.imageCarousel {width:100%%; white-space: nowrap; overflow-x: scroll; overflow-y: hidden; background-color:lightgrey; margin-bottom: 10px;}
	.imgContainer {margin:10px; cursor:pointer; display:inline-block;}
	.animationCollection {width:100%%; overflow-x: scroll; overflow-y: scroll; background-color:lightgrey; margin-bottom: 10px; max-height: 200px;}
	.animContainer {margin:10px; cursor:pointer; display:inline-block;}
	.close-btn {right:25px; top: 15px; position: absolute;}
	.carouselTitle {font-weight:bold; padding-left:5px;}
	.animationFrame {border: white 1px dashed;}

	.sortable {list-style-type: none; margin: 0px; padding: 0; width:100%%; display:inline-block;}
	.sortable li {margin:3px 3px 3px 0; padding: 1px; float:left;}

	.card-header {padding: 0.25rem 0.5rem}
	.card-body {padding: 0.5rem}
	.card-footer {padding: 0.25rem 0.5rem}

	.default-hidden {display: none;}

	img {width:50%%;}

	.imageSelector {}
	.drawerToggle {float:left; margin-left:10px; margin-right:5px;}
	</style>
	</head>
	<body>
	<div class='container-fluid'>
		<input id='color' type='color' />
		<button class='pallette btn btn-sm btn-danger' onclick="$('#color').spectrum('set', 'rgb(255,0,0)');setColor();">Red</button>
		<button class='pallette btn btn-sm btn-success' onclick="$('#color').spectrum('set', 'rgb(0,255,0)');setColor();">Green</button>
		<button class='pallette btn btn-sm btn-primary' onclick="$('#color').spectrum('set', 'rgb(0,0,255)');setColor();">Blue</button>
		
		<button id='clear' class='btn btn-sm btn-danger' onclick='clearDisplay()'>Clear</button>

		<table id='pixelTable' class='table table-striped table-bordered table-condensed'>%s</table>
		<div id='storage'>
			<div class='imageCarousel'>
				<div class='drawerToggle' data-target='images' onclick="toggleDrawer(this)"><i class='fas fa-window-minimize'></i><i class='fas fa-window-maximize default-hidden'></i></div>
				<span class='carouselTitle'>
					<b style='font-size:18px;'>Images</b>
					<button class='pallette btn btn-sm btn-info' onclick="saveCanvasAsImage()">Save Image <i class='fas fa-save'></i></button>
				<span>
				<div id='images'></div>
			</div>
			<div class='animationCollection'>
				<div class='drawerToggle' data-target='animations' onclick="toggleDrawer(this)"><i class='fas fa-window-minimize'></i><i class='fas fa-window-maximize default-hidden'></i></div>
				<span class='carouselTitle'>
					<b style='font-size:18px;'>Animations</b>
					<button class='pallette btn btn-sm btn-success' onclick="newAnimation()">New Animation <i class='fas fa-plus'></i></button> 
				</span>
				<div id='animations'></div>
			</div>
		</div>

	</div>

	</body>
	</html>`, buttons.String())

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
