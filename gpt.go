package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
	"strings"
	"time"
)

const (
	gptGuessPrompt = `Describe this drawing using just 1 to 4 words, preferably 
using a color if it is primarily comprised of one color (e.g., 'Orange cat' or 
'Red barn').  If the image is just a solid color or just has a few 'sketch marks' 
or dots, etc, then respond with an empty string, no special characters, until you 
can make out an object or scene and then you can describe it with 1 to 4 words as 
instructed previously. The colors you are allowed to use when describing the object 
must be one of [Black|White|Red|Green|Blue|Yellow|Orange|Purple|Teal|Pink|Brown].`

	gptStartGamePrompt = `Here are the game rules.  If the Difficulty is set to 
Easy, you shall choose a Color that is one of [Black|White|Red|Green|Blue]; if 
the Difficulty is set to Normal, you shall choose a Color that is one of 
[Black|White|Red|Green|Blue|Yellow|Orange|Purple|Teal|Pink|Brown]; if the 
Difficulty is set to Hard, you shall choose a Color that is one of 
[Orange|Purple|Teal|Pink|Brown]. After you have chosen a Color, you shall then 
choose an Object that can be described in 1 or 2 words; the Difficulty should 
influence the Object chosen, such that "easy objects" are those that would take  
fewer brush strokes to draw/paint, while "hard objects" would take more strokes, 
may require multiple colors/sub-shapes to make the object distinguished, etc. An 
example Easy response would be 'Red ball' while an example Hard response would be 
'Pink football field'. Notice that the only adjective used to describe the object 
is the Color (i.e., you should avoid object descriptions like: 'tall hat' or 
'Black tall hat', etc). Objects can come from any context, like nature, sports, 
games, fiction, office/home spaces, etc. Do not respond with any preceding comments 
when we play the game and do not repeat previous responses (do not choose the same 
object twice, even if using a different color; review the current context/history of 
this conversation to learn which objects should not be used again).  OK, let's play 
the game now. The Difficulty has been set to `
)

func guessRoutine(ctx context.Context, imageDirectory string, exportImageFunc func(), makeGuessFunc func(string)) {
	client := newGptClient()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if img := getLatestDrawing(imageDirectory); img != "" {
			base64Image := getImageB64(img)

			resp, err := client.CreateChatCompletion(
				context.Background(),
				*newImageCompletionRequest(base64Image),
			)

			if err == nil {
				guess := formatGuess(resp.Choices[0].Message.Content)
				makeGuessFunc(guess)
			} else {
				panic(fmt.Errorf("API error: %w\n", err))
			}
		}

		time.Sleep(gptGuessIntervalSec * time.Second)
		exportImageFunc()
	}
}

func newGptClient() *openai.Client {
	return openai.NewClient(os.Getenv("OPENAI_API_KEY"))
}

func newTextCompletionRequest(text, previousObjects string) *openai.ChatCompletionRequest {
	return &openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeText,
		},
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Objects you have used already, which should not be chosen again: " + previousObjects,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
	}
}

func newImageCompletionRequest(base64Image string) *openai.ChatCompletionRequest {
	return &openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeText,
		},
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: gptGuessPrompt,
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL:    fmt.Sprintf("data:image/png;base64,%s", base64Image),
							Detail: gptGuessAbility,
						},
					},
				},
			},
		},
	}
}

func formatChallenge(gptChallenge string) (formattedChallenge string) {
	formattedChallenge = gptChallenge
	formattedChallenge = strings.ReplaceAll(formattedChallenge, "*", "")
	formattedChallenge = strings.ReplaceAll(formattedChallenge, "_", "")
	formattedChallenge = strings.ReplaceAll(formattedChallenge, "'", "")
	formattedChallenge = strings.ReplaceAll(formattedChallenge, "\"", "")
	return
}

func formatGuess(gptGuess string) (formattedGuess string) {
	formattedGuess = gptGuess

	formattedGuess = strings.ReplaceAll(formattedGuess, "\"", "")
	formattedGuess = strings.ReplaceAll(formattedGuess, "''", "")
	formattedGuess = strings.ReplaceAll(formattedGuess, "*", "")
	formattedGuess = strings.ReplaceAll(formattedGuess, "`", "")
	formattedGuess = strings.ReplaceAll(formattedGuess, "_", "")

	formattedGuess = strings.ReplaceAll(formattedGuess, "drawing", "")
	formattedGuess = strings.ReplaceAll(formattedGuess, "sketch", "")
	formattedGuess = strings.ReplaceAll(formattedGuess, "outline", "")

	formattedGuess = strings.TrimSpace(formattedGuess)

	if strings.HasSuffix(formattedGuess, ".") {
		formattedGuess = formattedGuess[:len(formattedGuess)-1]
	}
	formattedGuess += "?"

	if len(formattedGuess) > 1 {
		formattedGuess = strings.ToUpper(formattedGuess[0:1]) + formattedGuess[1:]
	}

	return
}
