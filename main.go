package main

import (
	"context"
	"github.com/tonybillings/gfx"
	"os"
	"os/signal"
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

	gameView := NewPictionaryView(win, false)
	practiceView := NewPictionaryView(win, true)
	win.AddObjects(gfx.NewTabGroup(newHomeView(), practiceView, gameView))

	win.EnableQuitKey()
	win.EnableFullscreenKey()

	gfx.InitWindowAsync(win)

	exportFunc := getExportFunc(gameView, imgDir)
	guessFunc := getGuessFunc(gameView)
	go guessRoutine(ctx, imgDir, exportFunc, guessFunc)

	go waitForInterruptSignal(ctx, cancelFunc)
	gfx.Run(ctx, cancelFunc)
}
