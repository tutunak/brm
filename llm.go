package main

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

var llmModel = "gemini-3-pro-preview"
var googleAPIKey string

// analyzeURLWithLLM sends the URL to the LLM and returns the analysis
func analyzeURLWithLLM(url string) (string, error) {
	ctx := context.Background()

	if googleAPIKey == "" {
		return "", fmt.Errorf("GOOGLE_API_KEY not configured")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: googleAPIKey,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	var tools []*genai.Tool
	tools = append(tools, &genai.Tool{
		URLContext: &genai.URLContext{},
	})

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				genai.NewPartFromText(url),
			},
		},
	}

	config := &genai.GenerateContentConfig{
		ThinkingConfig: &genai.ThinkingConfig{
			ThinkingBudget: genai.Ptr[int32](1024),
		},
		Tools: tools,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				genai.NewPartFromText("Write a short summary why the text provided by a link is a bullshit. Don't pass inroduction or something else, just answer started with. This is a bullshit because <your answer>"),
			},
		},
	}

	var result strings.Builder
	for streamResult, err := range client.Models.GenerateContentStream(ctx, llmModel, contents, config) {
		if err != nil {
			return "", fmt.Errorf("stream error: %w", err)
		}

		if len(streamResult.Candidates) == 0 || streamResult.Candidates[0].Content == nil || len(streamResult.Candidates[0].Content.Parts) == 0 {
			continue
		}

		parts := streamResult.Candidates[0].Content.Parts
		for _, part := range parts {
			result.WriteString(part.Text)
		}
	}

	response := result.String()
	if response == "" {
		return "", fmt.Errorf("no response from LLM")
	}

	return response, nil
}

