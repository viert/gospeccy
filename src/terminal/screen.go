package terminal

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"image"
	"image/color"
	"port"
	"runtime"
	"time"
)

var (
	window      *glfw.Window
	texture     uint32
	render      *image.RGBA
	renderCount int
	border      FloatColor
	flashState  bool
	needRedraw  bool
	memory      []byte
)

type FloatColor struct {
	R float32
	G float32
	B float32
}

const (
	borderFactor = 0.2
	maxSize      = 1 - borderFactor
	screenWidth  = 256
	screenHeight = 192
)

func init() {
	runtime.LockOSThread()
}

func createWindow() *glfw.Window {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(640, 480, "Screen", nil, nil)
	if err != nil {
		panic(err)
	}
	return window
}

func SetRedraw() {
	needRedraw = true
}

func InitScreenWindow(mem []byte) {
	memory = mem
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	window = createWindow()
	window.MakeContextCurrent()
	window.SetKeyCallback(keyCallback)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	createTexture()

	setupScene()
}

func createTexture() {
	render = image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))

	gl.Enable(gl.TEXTURE_2D)
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, screenWidth, screenHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(render.Pix))
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func byteToFloat(in byte) float32 {
	return float32(in) / 255
}

func setBorderColor(data byte) {
	cl := getInkColor(data)
	border.R = byteToFloat(cl.R)
	border.G = byteToFloat(cl.G)
	border.B = byteToFloat(cl.B)
}

func subscribeBorderColor() {
	c := port.SubscribeOut("screen_border", 0x00FF, 0x00FE)

	data := port.GetOut(254)
	setBorderColor(data)

	for data = range c {
		setBorderColor(data)
	}
}

func flashInverse() {
	c := time.Tick(500 * time.Millisecond)
	for range c {
		flashState = !flashState
		SetRedraw()
	}
}

func StartTerminal(onReady func()) {
	go subscribeBorderColor()
	go flashInverse()
	if onReady != nil {
		onReady()
	}
	mainLoop()
	window.Destroy()
	glfw.Terminate()
}

func mainLoop() {
	for !window.ShouldClose() {
		drawScene()
		window.SwapBuffers()
		glfw.PollEvents()
		if needRedraw {
			renderSpectrum()
		}
	}
}

func renderSpectrum() {
	if render == nil || texture == 0 {
		// not initialized yet
		return
	}
	needRedraw = false
	renderCount++
	screenDump := memory[16384:23296]

	var inkColor color.RGBA
	var paperColor color.RGBA

	for i := 0; i < 0x1800; i++ {
		data := screenDump[i]
		cOffset := getColorOffset(i)
		colorValue := screenDump[cOffset]
		flashBit := getFlashBit(colorValue)

		if flashBit && flashState {
			inkColor = getPaperColor(colorValue)
			paperColor = getInkColor(colorValue)
		} else {
			inkColor = getInkColor(colorValue)
			paperColor = getPaperColor(colorValue)
		}
		coords := getCoords(i)
		for xoff := 0; xoff < 8; xoff++ {
			bit := (data >> uint(7-xoff)) & 1
			var pointColor color.RGBA
			if bit == 1 {
				pointColor = inkColor
			} else {
				pointColor = paperColor
			}
			render.SetRGBA(coords.X+xoff, screenHeight-1-coords.Y, pointColor)
		}
	}

	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, screenWidth, screenHeight, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(render.Pix))
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func getInkColor(colorValue byte) color.RGBA {
	b := uint8(colorValue&1) * 255
	r := uint8(colorValue>>1&1) * 255
	g := uint8(colorValue>>2&1) * 255
	return color.RGBA{r, g, b, 255}
}

func getPaperColor(colorValue byte) color.RGBA {
	b := uint8(colorValue>>3&1) * 255
	r := uint8(colorValue>>4&1) * 255
	g := uint8(colorValue>>5&1) * 255
	return color.RGBA{r, g, b, 255}
}

func getFlashBit(colorValue byte) bool {
	return colorValue&0x80 == 0x80
}

func getColorOffset(screenOffset int) int {
	tr := (screenOffset >> 11) & 3
	ba := screenOffset & 0xFF
	return tr*256 + ba + 0x1800
}

func getCoords(screenOffset int) image.Point {
	// Format is DDCCCBBBAAAAA
	a := screenOffset & 0x1F
	b := (screenOffset >> 5) & 0x07
	c := (screenOffset >> 8) & 0x07
	d := (screenOffset >> 11) & 0x03
	x := a * 8
	y := d*64 + b*8 + c
	return image.Point{x, y}
}

func setupScene() {
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

func drawScene() {
	gl.ClearColor(border.R, border.G, border.B, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.Color4f(1, 1, 1, 1)

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Begin(gl.QUADS)

	gl.Normal3f(0, 0, 1)
	gl.TexCoord2f(0, 0)
	gl.Vertex2f(-maxSize, -maxSize)

	gl.Normal3f(0, 0, 1)
	gl.TexCoord2f(0, 1)
	gl.Vertex2f(-maxSize, maxSize)

	gl.Normal3f(0, 0, 1)
	gl.TexCoord2f(1, 1)
	gl.Vertex2f(maxSize, maxSize)

	gl.Normal3f(0, 0, 1)
	gl.TexCoord2f(1, 0)
	gl.Vertex2f(maxSize, -maxSize)
	gl.End()
}
