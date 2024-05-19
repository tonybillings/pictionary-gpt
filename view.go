package main

import (
	"github.com/tonybillings/gfx"
	"github.com/tonybillings/gfx/examples/ui/view"
)

func NewPictionaryView() gfx.WindowObject {
	pictView := gfx.NewWindowObject()

	canvasView := view.NewCanvasView()
	canvasView.SetPositionX(-.2)

	guess1 := gfx.NewLabel()
	guess1.SetName("GuessLabel1")
	guess1.
		SetText("").
		SetFontSize(.045).
		SetAlignment(gfx.Centered).
		SetColor(gfx.White).
		SetAnchor(gfx.MiddleRight).
		SetMarginRight(.005).
		SetScaleX(.5)
	guess1.OnResize(func(newWidth, newHeight int) {
		guess1.Window().ScaleX(.5)
	})

	guess2 := gfx.NewLabel()
	guess2.SetName("GuessLabel2")
	guess2.
		SetText("").
		SetFontSize(.045).
		SetAlignment(gfx.Centered).
		SetColor(gfx.White).
		SetAnchor(gfx.MiddleRight).
		SetMarginRight(.005).
		SetMarginTop(.15).
		SetScaleX(.5)
	guess2.OnResize(func(newWidth, newHeight int) {
		guess2.Window().ScaleX(.5)
	})

	pictView.AddChildren(canvasView, guess1, guess2)

	return pictView
}
