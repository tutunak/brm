package main

import (
	"strings"
	"testing"
)

func TestGetOpinion(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedSentiment string
	}{
		{
			name:           "Empty text",
			input:          "",
			expectedSentiment: "No text to analyze.",
		},
		{
			name:           "Positive text",
			input:          "This is a great and wonderful day! I love it!",
			expectedSentiment: "positive 😊",
		},
		{
			name:           "Negative text",
			input:          "This is terrible and awful. I hate it!",
			expectedSentiment: "negative 😔",
		},
		{
			name:           "Neutral text",
			input:          "The weather today is cloudy.",
			expectedSentiment: "neutral 😐",
		},
		{
			name:           "Mixed sentiment with more positive",
			input:          "This is good but has some bad parts. Overall excellent!",
			expectedSentiment: "positive 😊",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOpinion(tt.input)
			
			if !strings.Contains(result, tt.expectedSentiment) {
				t.Errorf("getOpinion(%q) = %q, want to contain %q", tt.input, result, tt.expectedSentiment)
			}
		})
	}
}

func TestGetOpinionWordCount(t *testing.T) {
	input := "This is a test message with seven words"
	result := getOpinion(input)
	
	if !strings.Contains(result, "Word count: 7") {
		t.Errorf("Expected word count to be 7, got: %s", result)
	}
}

func TestGetOpinionFormat(t *testing.T) {
	input := "Test message"
	result := getOpinion(input)
	
	// Check if result contains expected sections
	expectedSections := []string{
		"📊 Opinion Analysis:",
		"Sentiment:",
		"Word count:",
		"Positive indicators:",
		"Negative indicators:",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(result, section) {
			t.Errorf("Expected result to contain %q, got: %s", section, result)
		}
	}
}
