package main

import (
	"context"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tonybillings/gfx"
	"github.com/tonybillings/gfx/examples/ui/view"
	"github.com/tonybillings/gfx/obj"
	"github.com/tonybillings/pictionary-gpt/models"
	"github.com/tonybillings/pictionary-gpt/textures"
	"image/color"
	"strings"
	"sync"
)

var (
	bronzeStarColor = gfx.Darken(gfx.Brown, .5)
	silverStarColor = gfx.Gray
	goldStarColor   = gfx.Darken(gfx.Yellow, .5)
)

func getExportFunc(gameView gfx.WindowObject, imageDirectory string) func() {
	canvas := gameView.Child("Canvas").(*gfx.Canvas)
	exportFunc := func() {
		if canvas.Initialized() {
			canvas.Export(imageDirectory)
		}
	}
	return exportFunc
}

func getGuessFunc(gameView gfx.WindowObject) func(string) {
	guess1 := gameView.Child("GuessLabel1").(*gfx.Label)
	guess2 := gameView.Child("GuessLabel2").(*gfx.Label)
	starContainer := gameView.Child("StarContainer").(*StarContainer)
	challengeLabel := gameView.Child("ChallengeLabel").(*gfx.Label)
	timer := gameView.Child("Timer").(*Timer)

	guessFunc := func(gptGuess string) {
		words := strings.Split(gptGuess, " ")
		if len(words) > 2 { // crude text wrapping
			guess1.SetText(fmt.Sprintf("%s %s", words[0], words[1]))
			guess2.SetText(fmt.Sprintf("%s", strings.Join(words[2:], " ")))
		} else {
			guess1.SetText(gptGuess)
			guess2.SetText("")
		}

		gptGuess = strings.ToLower(gptGuess)
		gptGuess = strings.ReplaceAll(gptGuess, "?", "")
		gptGuess = strings.ReplaceAll(gptGuess, ",", "")
		words = strings.Split(gptGuess, " ")

		if guess1.Text() == "?" {
			starContainer.SetStarVisibility(1, false)
			starContainer.SetStarVisibility(2, false)
			starContainer.SetStarVisibility(3, false)
		} else {
			answer := strings.ToLower(challengeLabel.Text())
			if strings.Contains(gptGuess, answer) {
				starContainer.SetStarVisibility(1, true)
				starContainer.SetStarVisibility(2, true)
				starContainer.SetStarVisibility(3, true)
				timer.SetTimeRemaining(0)
			} else if strings.Contains(answer, words[0]) {
				starContainer.SetStarVisibility(1, true)
				starContainer.SetStarVisibility(2, false)
				starContainer.SetStarVisibility(3, false)
			} else {
				starContainer.SetStarVisibility(1, false)
				starContainer.SetStarVisibility(2, false)
				starContainer.SetStarVisibility(3, false)

				for _, word := range words[1:] {
					if strings.Contains(answer, word) {
						starContainer.SetStarVisibility(1, true)
						starContainer.SetStarVisibility(2, true)
						starContainer.SetStarVisibility(3, false)
						break
					}
				}
			}
		}
	}

	return guessFunc
}

func newInkMeter(rgba color.RGBA) (outer, inner gfx.WindowObject) {
	inkMeter := gfx.NewWindowObject()
	inkMeter.
		SetScale(mgl32.Vec3{.3, .6}).
		SetPositionY(.35)

	inkMeterInner := gfx.NewQuad()
	inkMeterInner.
		SetAnchor(gfx.BottomCenter).
		SetColor(rgba)

	inkMeterFrame := gfx.NewSquare(.04)

	inkMeter.AddChildren(inkMeterInner, inkMeterFrame)

	return inkMeter, inkMeterInner
}

func newBrushControls(brush *InkBrush) gfx.WindowObject {
	redInkMeter, redInkMeterInner := newInkMeter(gfx.Red)
	brush.OnRedInkChanged(func(newInkLevel float64) {
		redInkMeterInner.SetScaleY(float32(newInkLevel))
	})

	greenInkMeter, greenInkMeterInner := newInkMeter(gfx.Green)
	brush.OnGreenInkChanged(func(newInkLevel float64) {
		greenInkMeterInner.SetScaleY(float32(newInkLevel))
	})

	blueInkMeter, blueInkMeterInner := newInkMeter(gfx.Blue)
	brush.OnBlueInkChanged(func(newInkLevel float64) {
		blueInkMeterInner.SetScaleY(float32(newInkLevel))
	})

	brushControls := view.NewBrushControls(&brush.BasicBrush)
	brushControls.SetPositionY(.2)
	brushControls.Child("RedSlider").AddChild(redInkMeter)
	brushControls.Child("GreenSlider").AddChild(greenInkMeter)
	brushControls.Child("BlueSlider").AddChild(blueInkMeter)

	refillButton := NewRainbowButton()
	refillButton.
		SetText("Refill").
		SetFontSize(.3).
		SetMouseEnterBorderColor(gfx.White).
		SetBorderThickness(.1).
		SetBorderColor(gfx.Purple).
		SetAnchor(gfx.Center).
		SetMarginBottom(.23).
		SetScale(mgl32.Vec3{1, 3})
	refillButton.OnClick(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		brush.RefillInk()
	})
	brushControls.Child("ColorPreview").AddChild(refillButton)

	return brushControls
}

func newGameControls(challengeLabel *gfx.Label, brush *InkBrush, starContainer *StarContainer) gfx.WindowObject {
	gameControls := gfx.NewView()
	gameControls.
		SetBorderColor(gfx.Purple).
		SetBorderThickness(.05).
		SetFillColor(gfx.Opacity(gfx.Purple, .5)).
		SetPositionY(.85).
		SetScaleY(.1)

	newGameLabel := gfx.NewLabel()
	newGameLabel.
		SetText("   New Game").
		SetFontSize(.5).
		SetAlignment(gfx.Left)

	easyButton := gfx.NewButton()
	easyButton.
		SetText("Easy").
		SetFontSize(.5).
		SetMouseEnterBorderColor(gfx.White).
		SetBorderColor(gfx.Lighten(gfx.Purple, .5)).
		SetBorderThickness(.2).
		SetAnchor(gfx.MiddleLeft).
		SetMarginLeft(.5).
		SetScale(mgl32.Vec3{.15, .4})

	normalButton := gfx.NewButton()
	normalButton.
		SetText("Normal").
		SetFontSize(.5).
		SetMouseEnterBorderColor(gfx.White).
		SetBorderColor(gfx.Lighten(gfx.Purple, .5)).
		SetBorderThickness(.2).
		SetAnchor(gfx.MiddleLeft).
		SetMarginLeft(.75).
		SetScale(mgl32.Vec3{.15, .4})

	hardButton := gfx.NewButton()
	hardButton.
		SetText("Hard").
		SetFontSize(.5).
		SetMouseEnterBorderColor(gfx.White).
		SetBorderColor(gfx.Lighten(gfx.Purple, .5)).
		SetBorderThickness(.2).
		SetAnchor(gfx.MiddleLeft).
		SetMarginLeft(1.).
		SetScale(mgl32.Vec3{.15, .4})

	timer := NewTimer(timerCountdownSec, timerNormalSec)
	timer.SetFontSize(.5)
	timer.SetVisibility(false).SetEnabled(false)
	timer.OnTimerStop(func() {
		timer.SetVisibility(false).SetEnabled(false)
		newGameLabel.SetVisibility(true).SetEnabled(true)
		easyButton.SetVisibility(true).SetEnabled(true)
		normalButton.SetVisibility(true).SetEnabled(true)
		hardButton.SetVisibility(true).SetEnabled(true)
	})

	client := newGptClient()
	objectHistory := make([]string, 0)
	var gameMutex sync.Mutex

	startGame := func(difficulty int) {
		gameMutex.Lock()
		challengeLabel.SetText("")
		brush.RefillInkInstantly()

		newGameLabel.SetVisibility(false).SetEnabled(false)
		easyButton.SetVisibility(false).SetEnabled(false)
		normalButton.SetVisibility(false).SetEnabled(false)
		hardButton.SetVisibility(false).SetEnabled(false)

		go func() {
			defer gameMutex.Unlock()

			prompt := gptStartGamePrompt
			timerSec := int64(0)

			switch difficulty {
			case 1:
				timerSec = timerEasySec
				prompt += "Easy"
				starContainer.SetColor(bronzeStarColor)
			case 2:
				timerSec = timerNormalSec
				prompt += "Normal"
				starContainer.SetColor(silverStarColor)
			case 3:
				timerSec = timerHardSec
				prompt += "Hard"
				starContainer.SetColor(goldStarColor)
			}

			starContainer.Reset()

			resp, err := client.CreateChatCompletion(
				context.Background(),
				*newTextCompletionRequest(prompt, fmt.Sprintf("[%s]", strings.Join(objectHistory, "|"))),
			)

			if err == nil {
				challenge := formatChallenge(resp.Choices[0].Message.Content)
				challengeLabel.SetText(challenge)
				challengeWords := strings.Split(challenge, " ")
				objectHistory = append(objectHistory, challengeWords[1])
			} else {
				panic(fmt.Errorf("API error: %w\n", err))
			}

			timer.Reset(timerCountdownSec, timerSec)
			timer.SetVisibility(true).SetEnabled(true)
		}()
	}

	easyButton.OnClick(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		startGame(1)
	})
	normalButton.OnClick(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		startGame(2)
	})
	hardButton.OnClick(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		startGame(3)
	})

	gameControls.AddChildren(newGameLabel, easyButton, normalButton, hardButton, timer)
	return gameControls
}

func newStarContainer(win *gfx.Window) *StarContainer {
	win.Assets().AddEmbeddedFiles(models.Assets)
	win.Assets().AddEmbeddedFiles(textures.Assets)

	model := obj.NewModel("star", "star.obj")
	win.Assets().Add(model)
	model.SetDefaultShader(win.Assets().Get(gfx.Shape3DNoNormalSpecularMapsShader).(gfx.Shader))
	model.Load()

	camera := gfx.NewCamera()
	camera.SetProjection(45, win.AspectRatio(), .1, 1000)
	camera.Properties.Position = mgl32.Vec4{0, 0, 20}
	camera.Properties.Target = mgl32.Vec4{0, 0, 19}
	win.AddObject(camera)

	lighting := gfx.NewQuadDirectionalLighting()
	lighting.LightCount = 3
	lighting.Lights[0].Color = mgl32.Vec3{.6, .5, .5}
	lighting.Lights[0].Direction = mgl32.Vec3{.5, .3, -1}
	lighting.Lights[1].Color = mgl32.Vec3{.5, .6, .5}
	lighting.Lights[1].Direction = mgl32.Vec3{.0, 1, -1.3}
	lighting.Lights[2].Color = mgl32.Vec3{.5, .5, .6}
	lighting.Lights[2].Direction = mgl32.Vec3{.5, -.3, -.7}

	viewport := gfx.NewViewport(win.Width(), win.Height())
	viewport.Set(.35, .35, 1, 1)

	star1 := NewStar()
	star1.SetName("Star1")
	star1.
		SetModel(model).
		SetCamera(camera).
		SetLighting(lighting).
		SetViewport(viewport).
		SetPositionX(-2)

	star2 := NewStar()
	star2.SetName("Star2")
	star2.
		SetModel(model).
		SetCamera(camera).
		SetLighting(lighting).
		SetViewport(viewport).
		SetPositionX(0)

	star3 := NewStar()
	star3.SetName("Star3")
	star3.
		SetModel(model).
		SetCamera(camera).
		SetLighting(lighting).
		SetViewport(viewport).
		SetPositionX(2)

	container := NewStarContainer()
	container.SetName("StarContainer")
	container.AddChildren(star1, star2, star3)
	container.SetColor(bronzeStarColor)

	return container
}

func newGameView(win *gfx.Window, exportDirectory ...string) gfx.WindowObject {
	exportDir := ""
	if len(exportDirectory) > 0 {
		exportDir = exportDirectory[0]
	}

	canvas := gfx.NewCanvas()
	canvas.
		SetFillColor(gfx.White).
		SetBorderColor(gfx.Purple).
		SetBorderThickness(.02).
		SetScale(mgl32.Vec3{.75, .75}).
		SetPositionX(.2)

	brush := NewInkBrush()
	brush.
		SetBrushHead(gfx.RoundBrushHead).
		SetSize(0.005).
		SetColor(gfx.Black)
	brush.SetCanvas(canvas)
	canvas.AddChild(brush)

	brushControls := newBrushControls(brush)

	canvasControls := view.NewCanvasControls(canvas, brush, exportDir)

	challengeLabel := gfx.NewLabel()
	challengeLabel.SetName("ChallengeLabel")
	challengeLabel.
		SetFontSize(.1).
		SetAlignment(gfx.Centered).
		SetMaintainAspectRatio(false).
		SetPositionY(-.85)
	canvas.AddChild(challengeLabel)

	starContainer := newStarContainer(win)

	gameControls := newGameControls(challengeLabel, brush, starContainer)
	canvas.AddChild(gameControls)

	container := gfx.NewWindowObject()
	container.SetMaintainAspectRatio(false)
	container.AddChildren(brushControls, canvasControls, canvas, starContainer)

	return container
}

func newHomeView() gfx.WindowObject {
	container := gfx.NewView()
	container.
		SetBorderThickness(.05).
		SetBorderColor(gfx.Purple).
		SetFillColor(gfx.Opacity(gfx.Purple, .5)).
		SetMaintainAspectRatio(true).
		SetScale(mgl32.Vec3{.8, .3})

	help1 := gfx.NewLabel()
	help1.
		SetText("Use TAB/ARROW keys to switch").
		SetFontSize(.15).
		SetMaintainAspectRatio(false).
		SetPositionY(.1)

	help2 := gfx.NewLabel()
	help2.
		SetText("between Practice/Game modes").
		SetFontSize(.15).
		SetMaintainAspectRatio(false).
		SetPositionY(-.1)

	container.AddChildren(help1, help2)

	return container
}

func NewPictionaryView(win *gfx.Window, practiceMode bool) gfx.WindowObject {
	pictView := gfx.NewWindowObject()
	pictView.SetMaintainAspectRatio(false)

	var canvasView gfx.WindowObject
	if practiceMode {
		canvasView = view.NewCanvasView()
	} else {
		canvasView = newGameView(win)
	}

	canvasView.SetPositionX(-.2)

	guess1 := gfx.NewLabel()
	guess1.SetName("GuessLabel1")
	guess1.SetMaintainAspectRatio(false)
	guess1.
		SetText("").
		SetFontSize(.045).
		SetAlignment(gfx.Centered).
		SetColor(gfx.White).
		SetAnchor(gfx.MiddleRight).
		SetMarginRight(.01).
		SetScaleX(.3)

	guess2 := gfx.NewLabel()
	guess2.SetName("GuessLabel2")
	guess2.
		SetText("").
		SetFontSize(.045).
		SetAlignment(gfx.Centered).
		SetColor(gfx.White).
		SetMaintainAspectRatio(false).
		SetAnchor(gfx.MiddleRight).
		SetMarginRight(.01).
		SetMarginTop(.15).
		SetScaleX(.3)

	pictView.AddChildren(canvasView, guess1, guess2)

	return pictView
}
