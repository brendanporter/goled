# goled
Raspberry Pi LED Matrix Controller

Controls one of these https://www.adafruit.com/product/420 with one of these https://www.raspberrypi.org/products/raspberry-pi-3-model-b-plus/ using a web application written in Go.

This project has a dependency on the C library from https://github.com/hzeller/rpi-rgb-led-matrix

As well as the Go binding to that C library here: https://github.com/mcuadros/go-rpi-rgb-led-matrix

I recommend following hzeller's fantastic instructions, then those of mcuadros.
Once those are satisfied, this code should compile on the Pi properly.

# Sanity warning

If using the Adafruit HAT or Bonnet, make sure to use HardwareConfiguration `adafruit-hat` or `adafruit-hat-pwm` as appropriate. If you use `regular` as found in many examples, you'll get no visual feedback, and no errors.


# Goals
- Create friendly web-based prototyping tool for testing designs
  - ~~Simple Web UI to assign colors to individual LEDs by clicking-to-paint on table cells~~
  - ~~"Drawing" mode to allow assignment of colors to LEDs by simply hovering over the table cells~~
  - Color selector using http://bgrins.github.io/spectrum/ (no more manual RGBA value inputs)
  - Clear canvas button
  - Write text with font selection
  - Create scrolling images and text
- Create frame capture and re-play logic to enable animation creation
  - Save canvas
  - Recall canvas
  - Create animation (slice of canvas frames)
  - Save animation frame
  - Recall animation frame
  - Play animation button
- Create API-hooking code to allow piping in data from outside sources

# Stretch Goals
- Create sign orchestration system to control many signs from a Single Pane of Glass
