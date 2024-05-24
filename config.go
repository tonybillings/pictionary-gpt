package main

import (
	"github.com/sashabaranov/go-openai"
)

const (
	windowTitle       = "Pictionary GPT"
	windowWidth       = 1900 // best to set to near/at native resolution
	windowHeight      = 1000 // best to set to near/at native resolution
	tempDirectory     = "/tmp/pictionary"
	timerCountdownSec = 5
	timerEasySec      = 60
	timerNormalSec    = 45
	timerHardSec      = 30
)

const ( // advanced settings
	targetFramerate     = 999 // effectively disable framerate-limiting
	vSyncEnabled        = false
	gptGuessIntervalSec = 5
	gptGuessAbility     = openai.ImageURLDetailLow
)
