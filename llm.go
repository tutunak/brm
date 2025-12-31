package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"google.golang.org/genai"
)

var llmModel = "gemini-flash-latest"
var googleAPIKey string

// Prompt types with their probabilities
type PromptType string

const (
	PromptBullshit PromptType = "bullshit" // 10%
	PromptPositive PromptType = "positive" // 40%
	PromptNegative PromptType = "negative" // 50%
)

// Base prompts for different tones
var basePrompts = map[PromptType]string{
	PromptBullshit: "Write a short summary why the text provided by a link is a bullshit. Don't write introduction or something else, just answer. If it's a github project - analyze it, and provide based arguments why it's a bullshit. Keep the answer short and funny.",
	PromptPositive: "Write a short summary with positive and well-argumented feedback about the content provided by a link. Don't write introduction, just answer. If it's a github project - analyze it and highlight the good aspects with solid arguments. Keep the answer short and encouraging.",
	PromptNegative: "Write a short summary with argumented criticism about why the content provided by a link is not good. Don't write introduction, just answer. If it's a github project - analyze it and provide solid arguments about its weaknesses. Keep the answer short and constructive but critical.",
}

// Video handling prompts for different tones
var videoPrompts = map[PromptType]string{
	PromptBullshit: " If it's a video - don't think long and answer that you will not watch such bullshit (make the answer random and creative each time).",
	PromptPositive: " If it's a video - politely explain that you can't watch videos but you're sure it must be interesting content.",
	PromptNegative: " If it's a video - rudely refuse to watch it and make a sarcastic comment about people who share videos instead of text.",
}

// selectPromptType randomly selects a prompt type based on probabilities
func selectPromptType() PromptType {
	r := rand.Float64() * 100
	
	if r < 10 {
		return PromptBullshit // 10%
	} else if r < 50 {
		return PromptPositive // 40% (10-50)
	} else {
		return PromptNegative // 40% (50-90) + remaining 10% = 50%
	}
}

// buildPrompt constructs the full prompt based on type
func buildPrompt(promptType PromptType) string {
	base := basePrompts[promptType]
	video := videoPrompts[promptType]
	return base + video
}

// analyzeURLWithLLM sends the URL to the LLM and returns the analysis
func analyzeURLWithLLM(url string) (string, error) {
	ctx := context.Background()
	startTime := time.Now()

	// Select prompt type based on probability
	promptType := selectPromptType()
	prompt := buildPrompt(promptType)

	logJSON("info", "Starting LLM analysis", map[string]interface{}{
		"url":         url,
		"model":       llmModel,
		"prompt_type": string(promptType),
	})

	if googleAPIKey == "" {
		logJSON("error", "LLM API key not configured", nil)
		return "", fmt.Errorf("GOOGLE_API_KEY not configured")
	}

	logJSON("debug", "Creating LLM client", map[string]interface{}{
		"api_key_length": len(googleAPIKey),
	})

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: googleAPIKey,
	})
	if err != nil {
		logJSON("error", "Failed to create LLM client", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	logJSON("debug", "LLM client created successfully", nil)

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
				genai.NewPartFromText(prompt),
			},
		},
	}

	logJSON("debug", "Starting LLM stream request", map[string]interface{}{
		"model":           llmModel,
		"thinking_budget": 1024,
		"tools_count":     len(tools),
		"prompt_type":     string(promptType),
	})

	var result strings.Builder
	chunkCount := 0
	for streamResult, err := range client.Models.GenerateContentStream(ctx, llmModel, contents, config) {
		if err != nil {
			logJSON("error", "LLM stream error", map[string]interface{}{
				"error":      err.Error(),
				"chunk":      chunkCount,
				"elapsed_ms": time.Since(startTime).Milliseconds(),
			})
			return "", fmt.Errorf("stream error: %w", err)
		}

		chunkCount++
		
		logJSON("debug", "Received LLM chunk", map[string]interface{}{
			"chunk_number":    chunkCount,
			"candidates":      len(streamResult.Candidates),
			"elapsed_ms":      time.Since(startTime).Milliseconds(),
		})

		if len(streamResult.Candidates) == 0 || streamResult.Candidates[0].Content == nil || len(streamResult.Candidates[0].Content.Parts) == 0 {
			logJSON("debug", "Empty chunk, skipping", map[string]interface{}{
				"chunk_number": chunkCount,
			})
			continue
		}

		parts := streamResult.Candidates[0].Content.Parts
		for i, part := range parts {
			result.WriteString(part.Text)
			logJSON("debug", "Processing part", map[string]interface{}{
				"chunk":       chunkCount,
				"part_index":  i,
				"text_length": len(part.Text),
				"text_preview": truncateString(part.Text, 50),
			})
		}
	}

	response := result.String()
	if response == "" {
		logJSON("error", "LLM returned empty response", map[string]interface{}{
			"url":        url,
			"chunks":     chunkCount,
			"elapsed_ms": time.Since(startTime).Milliseconds(),
		})
		return "", fmt.Errorf("no response from LLM")
	}

	logJSON("success", "LLM analysis completed", map[string]interface{}{
		"url":             url,
		"response_length": len(response),
		"chunks":          chunkCount,
		"elapsed_ms":      time.Since(startTime).Milliseconds(),
		"response_preview": truncateString(response, 100),
	})

	return response, nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

