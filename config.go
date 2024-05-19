package main

import (
	"github.com/sashabaranov/go-openai"
)

const (
	windowTitle   = "Pictionary GPT"
	windowWidth   = 1900
	windowHeight  = 1000
	tempDirectory = "/tmp/pictionary"
)

const ( // advanced settings
	targetFramerate     = 999 // effectively disable framerate-limiting
	vSyncEnabled        = false
	gptGuessIntervalSec = 10
	gptGuessAbility     = openai.ImageURLDetailLow
)
