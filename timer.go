package main

import (
	"fmt"
	"github.com/tonybillings/gfx"
	"image/color"
	"sync"
	"time"
)

/******************************************************************************
 Timer
******************************************************************************/

type Timer struct {
	gfx.Label

	countdownMilli     int64
	countdownLastMilli int64
	countdownLeftMilli int64
	countingDown       bool

	timeMilli      int64
	timeLastMilli  int64
	timeLeftMilli  int64
	timeRunningOut bool

	onTimerStop func()

	stateMutex sync.Mutex
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (t *Timer) Init() (ok bool) {
	if t.Initialized() {
		return true
	}

	t.timeLeftMilli = t.timeMilli
	t.SetText(fmt.Sprintf("%.3f", float32(t.timeLeftMilli)*.001))

	return t.Label.Init()
}

func (t *Timer) Update(deltaTime int64) (ok bool) {
	if ok = t.Label.Update(deltaTime); !ok {
		return
	}

	t.tick()
	return
}

/******************************************************************************
 Timer Functions
******************************************************************************/

func (t *Timer) defaultLayout() {
	t.Label.SetParent(t)

	t.SetColor(gfx.White)
	t.SetFillColor(color.RGBA{})

	t.SetFillColor(gfx.White)
	t.SetAnchor(gfx.Center)
}

func (t *Timer) tick() {
	t.stateMutex.Lock()

	now := time.Now().UnixMilli()

	if t.countingDown {
		t.countdownLeftMilli -= now - t.countdownLastMilli
		t.countdownLastMilli = now
		t.timeLastMilli = now
		if t.countdownLeftMilli <= 1000 { // make display/human-friendly for spoken countdowns
			t.countingDown = false
		} else {
			t.SetText(fmt.Sprintf("%d", int(float64(t.countdownLeftMilli)*.001)))
		}
		t.stateMutex.Unlock()
		return
	}

	t.timeLeftMilli -= now - t.timeLastMilli
	t.timeLastMilli = now

	if !t.timeRunningOut && t.timeLeftMilli < 11000 {
		t.timeRunningOut = true
		t.SetColor(gfx.Red)
	}

	if t.timeLeftMilli <= 0 {
		t.timeLeftMilli = 0
		t.SetText(fmt.Sprintf("%.3f", float32(t.timeLeftMilli)*.001))
		t.stateMutex.Unlock()
		t.SetEnabled(false)
		if t.onTimerStop != nil {
			t.onTimerStop()
		}
		return
	}

	t.SetText(fmt.Sprintf("%.3f", float32(t.timeLeftMilli)*.001))
	t.stateMutex.Unlock()
}

func (t *Timer) Reset(countdownSec, timeSec int64) {
	t.stateMutex.Lock()

	t.SetColor(gfx.White)

	t.countingDown = true
	t.countdownMilli = countdownSec * 1000
	t.countdownLeftMilli = t.countdownMilli + 999 // make display/human-friendly for spoken countdowns
	t.countdownLastMilli = time.Now().UnixMilli()

	t.timeRunningOut = false
	t.timeMilli = timeSec * 1000
	t.timeLeftMilli = t.timeMilli
	t.timeLastMilli = time.Now().UnixMilli()

	t.stateMutex.Unlock()
}

func (t *Timer) OnTimerStop(handler func()) {
	t.stateMutex.Lock()
	t.onTimerStop = handler
	t.stateMutex.Unlock()
}

func (t *Timer) SetTimeRemaining(timeMilli int64) {
	t.stateMutex.Lock()
	t.timeLeftMilli = timeMilli
	t.stateMutex.Unlock()
}

/******************************************************************************
 New Timer Function
******************************************************************************/

func NewTimer(countdownSec, timeSec int64) *Timer {
	t := &Timer{
		Label: *gfx.NewLabel(),
	}

	t.SetName("Timer")
	t.Reset(countdownSec, timeSec)

	return t
}
