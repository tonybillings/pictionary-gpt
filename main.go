package main

import (
	"context"
	"fmt"
	"github.com/tonybillings/gfx"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func waitForInterruptSignal(ctx context.Context, cancelFunc context.CancelFunc) {
	sigIntChan := make(chan os.Signal, 1)
	signal.Notify(sigIntChan, syscall.SIGINT)

	select {
	case <-ctx.Done():
		return
	case <-sigIntChan:
		cancelFunc()
		return
	}
}

func main() {
	panicOnErr(gfx.Init())
	defer gfx.Close()

	gfx.SetTargetFramerate(targetFramerate)
	gfx.SetVSyncEnabled(vSyncEnabled)

	win := gfx.NewWindow().
		SetTitle(windowTitle).
		SetWidth(windowWidth).
		SetHeight(windowHeight)

	ctx, cancelFunc := context.WithCancel(context.Background())

	imgDir := prepareImageDirectory(tempDirectory)

	view := NewPictionaryView()
	win.AddObjects(view)

	win.EnableQuitKey()
	win.EnableFullscreenKey()

	gfx.InitWindowAsync(win)

	canvas := view.Child("Canvas").(*gfx.Canvas)
	exportFunc := func() {
		if canvas.Initialized() {
			canvas.Export(imgDir)
		}
	}

	guess1 := view.Child("GuessLabel1").(*gfx.Label)
	guess2 := view.Child("GuessLabel2").(*gfx.Label)
	guessFunc := func(gptGuess string) {
		words := strings.Split(gptGuess, " ")
		if len(words) == 4 {
			guess1.SetText(fmt.Sprintf("%s %s", words[0], words[1]))
			guess2.SetText(fmt.Sprintf("%s %s", words[2], words[3]))
		} else {
			guess1.SetText(gptGuess)
			guess2.SetText("")
		}
	}
	go guessRoutine(ctx, imgDir, exportFunc, guessFunc)

	go waitForInterruptSignal(ctx, cancelFunc)
	gfx.Run(ctx, cancelFunc)
}
