package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tonybillings/gfx"
	"github.com/tonybillings/gfx/obj"
	"image/color"
	"sync/atomic"
)

/******************************************************************************
 Star
******************************************************************************/

type Star struct {
	gfx.Shape3D

	material *obj.BasicMaterial
	rgba     mgl32.Vec4

	rotation uint64
	lastRot  float64
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (s *Star) Init() (ok bool) {
	if ok = s.Shape3D.Init(); !ok {
		return
	}

	s.initColor()
	return
}

func (s *Star) Update(deltaTime int64) (ok bool) {
	if ok = s.Shape3D.Update(deltaTime); !ok {
		return
	}

	s.updateRotation(deltaTime)
	return
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (s *Star) Draw(deltaTime int64) (ok bool) {
	if !s.Initialized() {
		return false
	}

	s.updateColor()

	return s.Shape3D.Draw(deltaTime)
}

/******************************************************************************
 Star Functions
******************************************************************************/

func (s *Star) initColor() {
	s.material = s.Meshes()[0].Faces()[0].Material().(*obj.BasicMaterial)
	s.rgba = gfx.RgbaToFloatArray(s.Color())
}

func (s *Star) updateRotation(deltaTime int64) {
	s.rotation++
	dt := float64(int(float64(deltaTime) * .001))
	rot := float64(s.rotation) * 0.001 * dt
	if rot > 0 && rot-s.lastRot < 1 {
		s.SetRotationY(float32(rot))
		s.lastRot = rot
	}
}

func (s *Star) updateColor() {
	s.rgba = gfx.RgbaToFloatArray(s.Color())
	s.material.Lock()
	s.material.Properties.Emissive = s.rgba
	s.material.Unlock()
}

/******************************************************************************
 New Star Function
******************************************************************************/

func NewStar() *Star {
	return &Star{
		Shape3D: *gfx.NewShape3D(),
	}
}

/******************************************************************************
 StarContainer
******************************************************************************/

type StarContainer struct {
	gfx.WindowObjectBase

	star1 *Star
	star2 *Star
	star3 *Star

	star1Visible bool
	star2Visible bool
	star3Visible bool

	stateChanged atomic.Bool
}

func (c *StarContainer) Init() (ok bool) {
	if ok = c.WindowObjectBase.Init(); !ok {
		return
	}

	c.star1 = c.Child("Star1").(*Star)
	c.star2 = c.Child("Star2").(*Star)
	c.star3 = c.Child("Star3").(*Star)

	c.star1.SetVisibility(c.star1Visible)
	c.star2.SetVisibility(c.star2Visible)
	c.star3.SetVisibility(c.star3Visible)

	return
}

func (c *StarContainer) Update(deltaTime int64) (ok bool) {
	if ok = c.WindowObjectBase.Update(deltaTime); !ok {
		return
	}

	if c.stateChanged.Load() {
		c.stateChanged.Store(false)
		rgba := c.WindowObjectBase.Color()
		c.star1.SetColor(rgba)
		c.star2.SetColor(rgba)
		c.star3.SetColor(rgba)
		c.star1.SetVisibility(c.star1Visible)
		c.star2.SetVisibility(c.star2Visible)
		c.star3.SetVisibility(c.star3Visible)
	}

	return
}

func (c *StarContainer) SetColor(rgba color.RGBA) gfx.WindowObject {
	c.WindowObjectBase.SetColor(rgba)
	c.stateChanged.Store(true)
	return c
}

func (c *StarContainer) SetStarVisibility(starNumber int, visible bool) *StarContainer {
	switch starNumber {
	case 1:
		c.star1Visible = visible
	case 2:
		c.star2Visible = visible
	case 3:
		c.star3Visible = visible
	}
	c.stateChanged.Store(true)
	return c
}

func (c *StarContainer) Reset() gfx.WindowObject {
	c.star1Visible = false
	c.star2Visible = false
	c.star3Visible = false
	c.stateChanged.Store(true)
	return c
}

func NewStarContainer() *StarContainer {
	return &StarContainer{
		WindowObjectBase: *gfx.NewWindowObject(),
	}
}
