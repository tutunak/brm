package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/genai"
)

var llmModel = "gemini-3-pro-preview"
var googleAPIKey string

// analyzeURLWithLLM sends the URL to the LLM and returns the analysis
func analyzeURLWithLLM(url string) (string, error) {
	ctx := context.Background()
	startTime := time.Now()

	logJSON("info", "Starting LLM analysis", map[string]interface{}{
		"url":   url,
		"model": llmModel,
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
				genai.NewPartFromText("Write a short summary why the text provided by a link is a bullshit. Don't pass inroduction or something else, just answer started with. This is a bullshit because <your answer>"),
			},
		},
	}

	logJSON("debug", "Starting LLM stream request", map[string]interface{}{
		"model":          llmModel,
		"thinking_budget": 1024,
		"tools_count":    len(tools),
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

