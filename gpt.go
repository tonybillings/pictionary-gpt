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
	gptPrompt = `Describe this drawing using just 1 to 4 words, preferably 
using a color if it is primarily comprised of one color (e.g., 'orange cat' or 
'red barn').  If the image is just a solid color or just has a few 'sketch marks' 
or dots, etc, then respond with an empty string, no special characters, until you 
can make out an object or scene and then you can describe it with 1 to 4 words as 
instructed previously.`
)

func guessRoutine(ctx context.Context, imageDirectory string, exportImageFunc func(), makeGuessFunc func(string)) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

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
				openai.ChatCompletionRequest{
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
									Text: gptPrompt,
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
				},
			)

			if err == nil {
				guess := resp.Choices[0].Message.Content
				guess = strings.ReplaceAll(guess, "\"", "")
				guess = strings.ReplaceAll(guess, "''", "")
				if len(guess) > 0 {
					if strings.HasSuffix(guess, ".") {
						guess = guess[:len(guess)-1]
					}
					guess = strings.TrimSpace(guess) + "?"
					makeGuessFunc(guess)
				} else {
					makeGuessFunc("")
				}
			} else {
				panic(fmt.Errorf("API error: %w\n", err))
			}
		}

		time.Sleep(gptGuessIntervalSec * time.Second)
		exportImageFunc()
	}
}
