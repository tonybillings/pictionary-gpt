package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/tonybillings/gfx"
	"math"
)

/******************************************************************************
 RainbowButton
******************************************************************************/

type RainbowButton struct {
	gfx.Button
	texture       *gfx.Texture2D
	textureBuffer []uint8
	tick          int
	freqMod       float64
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (b *RainbowButton) Init() (ok bool) {
	b.texture = gfx.NewTexture2D("rainbow_texture", gfx.White)
	b.texture.Init()
	b.SetTexture(b.texture)
	b.textureBuffer = make([]uint8, b.texture.Width()*b.texture.Height()*4)

	return b.Button.Init()
}

func (b *RainbowButton) Update(deltaTime int64) (ok bool) {
	if ok = b.Button.Update(deltaTime); !ok {
		return
	}

	gl.BindTexture(gl.TEXTURE_2D, b.texture.GlName())
	gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&b.textureBuffer[0]))

	b.tick++
	for i := 0; i < 4; i++ {
		b.textureBuffer[i*4+0] = b.animateColor(i*0, b.tick)
		b.textureBuffer[i*4+1] = b.animateColor(i*2, b.tick)
		b.textureBuffer[i*4+2] = b.animateColor(i*4, b.tick)
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(b.texture.Width()), int32(b.texture.Height()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(b.textureBuffer))
	gl.BindTexture(gl.TEXTURE_2D, 0)

	return
}

func (b *RainbowButton) Close() {
	if !b.Initialized() {
		return
	}

	b.texture.Close()
	b.Button.Close()
}

/******************************************************************************
 RainbowButton Functions
******************************************************************************/

func (b *RainbowButton) defaultLayout() {
	b.Button.SetParent(b)
	b.freqMod = 2500.0
}

func (b *RainbowButton) animateColor(phase, tick int) uint8 {
	return uint8(math.Sin(math.Pi*float64(tick)/b.freqMod+(float64(phase)*math.Pi/4))*127 + 127)
}

func (b *RainbowButton) AnimationSpeed() float64 {
	return 1 / b.freqMod
}

func (b *RainbowButton) SetAnimationSpeed(speed float64) *RainbowButton {
	b.freqMod = 1 / speed
	return b
}

/******************************************************************************
 New RainbowButton Function
******************************************************************************/

func NewRainbowButton() *RainbowButton {
	b := &RainbowButton{
		Button: *gfx.NewButton(false),
	}

	b.defaultLayout()

	return b
}
