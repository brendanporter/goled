package main

import (
	"os"
	"os/signal"
	//"github.com/stianeikeland/go-rpio"
	"github.com/mcuadros/go-rpi-rgb-led-matrix"
	//"image"
	"image/color"
	//"image/draw"
	"log"
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

	chanSig := make(chan os.Signal)
	signal.Notify(chanSig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-chanSig
		c.Close()
		os.Exit(1)
		// sig is a ^C, handle it
	}()

	/*
		err := rpio.Open()
		if err != nil {
			log.Panic(err)
		}
		defer rpio.Close()

		pin := rpio.Pin()

		pin.Output()
		pin.High()
		pin.Low()
		pin.Toggle()
	*/

	config := &rgbmatrix.DefaultConfig
	config.Rows = 16
	config.Cols = 64
	config.HardwareMapping = "adafruit-hat-pwm"
	config.DisableHardwarePulsing = false
	config.ShowRefreshRate = true

	m, err := rgbmatrix.NewRGBLedMatrix(config)
	if err != nil {
		log.Print(err)
	}

	c = rgbmatrix.NewCanvas(m)
	defer c.Close()

	//draw.Draw(c, c.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)

	//c.Render()

	//time.Sleep(time.Second * 5)
	//c.Clear()
	/*
		go cylonRed()
		time.Sleep(time.Millisecond * (20 * 30))
		go cylonGreen()
		time.Sleep(time.Millisecond * (20 * 60))
		go cylonBlue()
	*/
	go square()

	for {

		time.Sleep(time.Second * 5)
		//draw.Draw(c, c.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)
		//c.Render()
		//time.Sleep(time.Millisecond * 400)
	}

}

func square() {
	c.Set(7, 14, color.RGBA{0, 0, 255, 255})
	c.Set(7, 15, color.RGBA{0, 255, 0, 255})
	c.Render()
}

func cylonRed() {
	bounds := c.Bounds()
	frame := time.Millisecond * 20
	for {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.Set(x, y, color.RGBA{255, 0, 0, 255})
				//c.Render()
			}
			c.Render()
			time.Sleep(frame)
		}

		for x := bounds.Max.X - 1; x > bounds.Min.X; x-- {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.Set(x, y, color.RGBA{255, 0, 0, 255})
				//c.Render()
			}
			c.Render()
			time.Sleep(frame)
		}
	}
}

func cylonGreen() {
	bounds := c.Bounds()
	frame := time.Millisecond * 20
	for {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.Set(x, y, color.RGBA{0, 255, 0, 255})
				//c.Render()
			}
			c.Render()
			time.Sleep(frame)
		}

		for x := bounds.Max.X - 1; x > bounds.Min.X; x-- {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.Set(x, y, color.RGBA{0, 255, 0, 255})
				//c.Render()
			}
			c.Render()
			time.Sleep(frame)
		}
	}
}

func cylonBlue() {
	bounds := c.Bounds()
	frame := time.Millisecond * 20
	for {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.Set(x, y, color.RGBA{0, 0, 255, 125})
				//c.Render()
			}
			c.Render()
			time.Sleep(frame)
		}

		for x := bounds.Max.X - 1; x > bounds.Min.X; x-- {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.Set(x, y, color.RGBA{0, 0, 255, 125})
				//c.Render()
			}
			c.Render()
			time.Sleep(frame)
		}
	}
}
