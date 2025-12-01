package main

import (
	"context"
	"log"
	"os"

	"google.golang.org/genai"
)

func checkAI(imageData []byte, mimeType string) string {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Print("AI: No GEMINI_API_KEY configured")
		return ""
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Printf("AI: Failed to create client: %v", err)
		return ""
	}

	parts := []*genai.Part{
		{Text: "You are a model designed to test if an image is winter/christmas themed or not. Upon review of an image you must STRICTLY respond with only 'yes' if it is winter/christmas themed or 'no' if it is not. Please check if this image is winter/christmas themed."},
		{InlineData: &genai.Blob{
			MIMEType: mimeType,
			Data:     imageData,
		}},
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		[]*genai.Content{{Parts: parts}},
		nil,
	)
	if err != nil {
		log.Printf("AI: Request failed: %v", err)
		return ""
	}

	text := result.Text()
	log.Printf("AI: Response: %s", text)
	return text
}
