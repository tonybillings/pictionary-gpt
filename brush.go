package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/tonybillings/gfx"
	"image/color"
	"sync"
)

/******************************************************************************
 InkBrush
******************************************************************************/

type InkBrush struct {
	gfx.BasicBrush

	canvasBuffer []uint8

	redInk   float64
	greenInk float64
	blueInk  float64

	drainRate    float64
	drainRateMod float64

	refilling bool

	onRedInkChanged   func(float64)
	onGreenInkChanged func(float64)
	onBlueInkChanged  func(float64)

	stateMutex sync.Mutex
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (b *InkBrush) Init() (ok bool) {
	if ok = b.BasicBrush.Init(); !ok {
		return
	}

	b.initDrainRateMod()

	return true
}

func (b *InkBrush) Update(deltaTime int64) (ok bool) {
	if b.refilling {
		b.stateMutex.Lock()
		b.refillInk()
		b.stateMutex.Unlock()
		b.dispatchEvents()
		return true
	}

	return b.BasicBrush.Update(deltaTime)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (b *InkBrush) Resize(newWidth, newHeight int) {
	b.BasicBrush.Resize(newWidth, newHeight)
	b.stateMutex.Lock()
	b.initDrainRateMod()
	b.stateMutex.Unlock()
}

/******************************************************************************
 InkBrush Functions
******************************************************************************/

// initDrainRateMod ensures the drain rate scales with the window size for
// a consistent experience across different screen resolutions, etc.
func (b *InkBrush) initDrainRateMod() {
	winWidth, winHeight := 1000.0, 1000.0
	if win := b.Window(); win != nil {
		winWidth = float64(win.Width())
		winHeight = float64(win.Height())
	}
	b.drainRateMod = 1 / (winWidth * winHeight)
}

func (b *InkBrush) getBrushProperties() (textureColor color.RGBA, redDrain, greenDrain, blueDrain float64) {
	textureColor = b.Color()
	redDrain = (float64(textureColor.R) / 255.0) * b.drainRate * b.drainRateMod
	greenDrain = (float64(textureColor.G) / 255.0) * b.drainRate * b.drainRateMod
	blueDrain = (float64(textureColor.B) / 255.0) * b.drainRate * b.drainRateMod

	if b.redInk <= 0 {
		textureColor.R = 0
	}

	if b.greenInk <= 0 {
		textureColor.G = 0
	}

	if b.blueInk <= 0 {
		textureColor.B = 0
	}

	return
}

func (b *InkBrush) refillInk() {
	if b.redInk < 1.0 {
		b.redInk += b.drainRate * .0001
		if b.redInk > 1.0 {
			b.redInk = 1.0
		}
	}
	if b.greenInk < 1.0 {
		b.greenInk += b.drainRate * .0001
		if b.greenInk > 1.0 {
			b.greenInk = 1.0
		}
	}
	if b.blueInk < 1.0 {
		b.blueInk += b.drainRate * .0001
		if b.blueInk > 1.0 {
			b.blueInk = 1.0
		}
	}

	if b.redInk == 1.0 && b.greenInk == 1.0 && b.blueInk == 1.0 {
		b.refilling = false
	}
}

func (b *InkBrush) updateCanvas(mouse *gfx.MouseState) {
	surface := b.Canvas().Surface()
	width := surface.Width()
	height := surface.Height()

	textureColor, redDrain, greenDrain, blueDrain := b.getBrushProperties()
	tx := int((mouse.X + 1) / 2 * float32(width))
	ty := int((mouse.Y + 1) / 2 * float32(height))
	radius := int(b.Size() * (float32(width) * 0.5))

	b.stateMutex.Lock()

	if b.canvasBuffer == nil {
		b.canvasBuffer = make([]uint8, width*height*4)
	}

	gl.BindTexture(gl.TEXTURE_2D, surface.GlName())
	gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&b.canvasBuffer[0]))

	switch b.BrushHead() {
	case gfx.RoundBrushHead:
		b.updateCanvasRoundHead(width, height, textureColor, redDrain, greenDrain, blueDrain, radius, tx, ty)
	case gfx.SquareBrushHead:
		b.updateCanvasSquareHead(width, height, textureColor, redDrain, greenDrain, blueDrain, radius, tx, ty)
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(b.canvasBuffer))
	gl.BindTexture(gl.TEXTURE_2D, 0)

	b.stateMutex.Unlock()

	b.dispatchEvents()
}

func (b *InkBrush) updateCanvasRoundHead(surfaceWidth, surfaceHeight int,
	textureColor color.RGBA, redDrain, greenDrain, blueDrain float64, radius, tx, ty int) {
	for i := -radius; i <= radius; i++ {
		for j := -radius; j <= radius; j++ {
			if i*i+j*j <= radius*radius {
				px := tx + i
				py := ty + j
				if px >= 0 && px < surfaceWidth && py >= 0 && py < surfaceHeight {
					index := (py*surfaceWidth + px) * 4
					b.paintCanvas(index, &textureColor, redDrain, greenDrain, blueDrain)
				}
			}
		}
	}
}

func (b *InkBrush) updateCanvasSquareHead(surfaceWidth, surfaceHeight int,
	textureColor color.RGBA, redDrain, greenDrain, blueDrain float64, radius, tx, ty int) {
	for i := -radius; i <= radius; i++ {
		for j := -radius; j <= radius; j++ {
			px := tx + i
			py := ty + j
			if px >= 0 && px < surfaceWidth && py >= 0 && py < surfaceHeight {
				index := (py*surfaceWidth + px) * 4
				b.paintCanvas(index, &textureColor, redDrain, greenDrain, blueDrain)
			}
		}
	}
}

func (b *InkBrush) paintCanvas(index int, textureColor *color.RGBA,
	redDrain, greenDrain, blueDrain float64) {
	if b.redInk > 0 {
		b.redInk -= redDrain
		if b.redInk <= 0 {
			b.redInk = 0
			textureColor.R = 0
		}
	}
	b.canvasBuffer[index] = textureColor.R

	if b.greenInk > 0 {
		b.greenInk -= greenDrain
		if b.greenInk <= 0 {
			b.greenInk = 0
			textureColor.G = 0
		}
	}
	b.canvasBuffer[index+1] = textureColor.G

	if b.blueInk > 0 {
		b.blueInk -= blueDrain
		if b.blueInk <= 0 {
			b.blueInk = 0
			textureColor.B = 0
		}
	}
	b.canvasBuffer[index+2] = textureColor.B

	b.canvasBuffer[index+3] = textureColor.A
}

func (b *InkBrush) dispatchEvents() {
	if b.onRedInkChanged != nil {
		b.onRedInkChanged(b.redInk)
	}

	if b.onGreenInkChanged != nil {
		b.onGreenInkChanged(b.greenInk)
	}

	if b.onBlueInkChanged != nil {
		b.onBlueInkChanged(b.blueInk)
	}
}

func (b *InkBrush) DrainRate() (rate float64) {
	b.stateMutex.Lock()
	rate = b.drainRate
	b.stateMutex.Unlock()
	return
}

func (b *InkBrush) SetDrainRate(rate float64) *InkBrush {
	b.stateMutex.Lock()
	b.drainRate = rate
	b.stateMutex.Unlock()
	return b
}

func (b *InkBrush) RedInk() (level float64) {
	b.stateMutex.Lock()
	level = b.redInk
	b.stateMutex.Unlock()
	return
}

func (b *InkBrush) GreenInk() (level float64) {
	b.stateMutex.Lock()
	level = b.greenInk
	b.stateMutex.Unlock()
	return
}

func (b *InkBrush) BlueInk() (level float64) {
	b.stateMutex.Lock()
	level = b.blueInk
	b.stateMutex.Unlock()
	return
}

func (b *InkBrush) RefillInk() {
	b.stateMutex.Lock()
	b.refilling = true
	b.stateMutex.Unlock()
}

func (b *InkBrush) RefillInkInstantly() {
	b.stateMutex.Lock()
	b.redInk = 1.0
	b.greenInk = 1.0
	b.blueInk = 1.0
	b.refilling = false
	b.stateMutex.Unlock()
	b.dispatchEvents()
}

func (b *InkBrush) OnRedInkChanged(handler func(newInkLevel float64)) {
	b.onRedInkChanged = handler
}

func (b *InkBrush) OnGreenInkChanged(handler func(newInkLevel float64)) {
	b.onGreenInkChanged = handler
}

func (b *InkBrush) OnBlueInkChanged(handler func(newInkLevel float64)) {
	b.onBlueInkChanged = handler
}

/******************************************************************************
 New InkBrush Function
******************************************************************************/

func NewInkBrush() *InkBrush {
	b := &InkBrush{
		BasicBrush: *gfx.NewBasicBrush(),
		redInk:     1.0,
		greenInk:   1.0,
		blueInk:    1.0,
		drainRate:  2.5,
	}

	b.SetName("InkBrush")
	b.OverrideUpdateCanvas(b.updateCanvas)
	return b
}
