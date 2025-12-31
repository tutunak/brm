package main

import (
	"strings"
	"testing"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "String shorter than maxLen",
			input:    "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "String equal to maxLen",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "String longer than maxLen",
			input:    "Hello, World!",
			maxLen:   5,
			expected: "Hello...",
		},
		{
			name:     "Empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "MaxLen is zero",
			input:    "Hello",
			maxLen:   0,
			expected: "...",
		},
		{
			name:     "MaxLen is one",
			input:    "Hello",
			maxLen:   1,
			expected: "H...",
		},
		{
			name:     "Very long string",
			input:    "This is a very long string that should be truncated at some point",
			maxLen:   20,
			expected: "This is a very long ...",
		},
		{
			name:     "Unicode string - within limit",
			input:    "Привет",
			maxLen:   20,
			expected: "Привет",
		},
		{
			name:     "String with emojis",
			input:    "Hello 😀 World",
			maxLen:   10,
			expected: "Hello 😀...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestSelectPromptType(t *testing.T) {
	// Run many times to ensure all types are returned
	typeCounts := make(map[PromptType]int)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		promptType := selectPromptType()
		typeCounts[promptType]++
	}

	// Check all types are present
	if typeCounts[PromptBullshit] == 0 {
		t.Error("PromptBullshit was never selected")
	}
	if typeCounts[PromptPositive] == 0 {
		t.Error("PromptPositive was never selected")
	}
	if typeCounts[PromptNegative] == 0 {
		t.Error("PromptNegative was never selected")
	}

	// Check approximate distribution (with some tolerance)
	// PromptBullshit should be ~10%, PromptPositive ~40%, PromptNegative ~50%
	bullshitPct := float64(typeCounts[PromptBullshit]) / float64(iterations) * 100
	positivePct := float64(typeCounts[PromptPositive]) / float64(iterations) * 100
	negativePct := float64(typeCounts[PromptNegative]) / float64(iterations) * 100

	// Allow 8% tolerance for random variation
	if bullshitPct < 2 || bullshitPct > 18 {
		t.Errorf("PromptBullshit percentage = %.1f%%, expected ~10%%", bullshitPct)
	}
	if positivePct < 32 || positivePct > 48 {
		t.Errorf("PromptPositive percentage = %.1f%%, expected ~40%%", positivePct)
	}
	if negativePct < 42 || negativePct > 58 {
		t.Errorf("PromptNegative percentage = %.1f%%, expected ~50%%", negativePct)
	}
}

func TestSelectPromptTypeReturnsValidType(t *testing.T) {
	validTypes := map[PromptType]bool{
		PromptBullshit: true,
		PromptPositive: true,
		PromptNegative: true,
	}

	for i := 0; i < 100; i++ {
		promptType := selectPromptType()
		if !validTypes[promptType] {
			t.Errorf("selectPromptType() returned invalid type: %v", promptType)
		}
	}
}

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name       string
		promptType PromptType
		contains   []string
	}{
		{
			name:       "Bullshit prompt",
			promptType: PromptBullshit,
			contains: []string{
				"bullshit",
				"short",
				"funny",
				"video",
			},
		},
		{
			name:       "Positive prompt",
			promptType: PromptPositive,
			contains: []string{
				"positive",
				"encouraging",
				"good aspects",
				"video",
			},
		},
		{
			name:       "Negative prompt",
			promptType: PromptNegative,
			contains: []string{
				"criticism",
				"weaknesses",
				"critical",
				"video",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPrompt(tt.promptType)

			for _, substr := range tt.contains {
				if !strings.Contains(strings.ToLower(result), strings.ToLower(substr)) {
					t.Errorf("buildPrompt(%v) = %q, expected to contain %q", tt.promptType, result, substr)
				}
			}
		})
	}
}

func TestBuildPromptCombinesBaseAndVideo(t *testing.T) {
	for _, promptType := range []PromptType{PromptBullshit, PromptPositive, PromptNegative} {
		t.Run(string(promptType), func(t *testing.T) {
			result := buildPrompt(promptType)
			base := basePrompts[promptType]
			video := videoPrompts[promptType]

			expected := base + video
			if result != expected {
				t.Errorf("buildPrompt(%v) = %q, want %q", promptType, result, expected)
			}
		})
	}
}

func TestBasePromptsExist(t *testing.T) {
	expectedTypes := []PromptType{PromptBullshit, PromptPositive, PromptNegative}

	for _, pt := range expectedTypes {
		if _, exists := basePrompts[pt]; !exists {
			t.Errorf("basePrompts missing entry for %v", pt)
		}
		if basePrompts[pt] == "" {
			t.Errorf("basePrompts[%v] is empty", pt)
		}
	}
}

func TestVideoPromptsExist(t *testing.T) {
	expectedTypes := []PromptType{PromptBullshit, PromptPositive, PromptNegative}

	for _, pt := range expectedTypes {
		if _, exists := videoPrompts[pt]; !exists {
			t.Errorf("videoPrompts missing entry for %v", pt)
		}
		if videoPrompts[pt] == "" {
			t.Errorf("videoPrompts[%v] is empty", pt)
		}
	}
}

func TestAnalyzeURLWithLLMNoAPIKey(t *testing.T) {
	// Save original googleAPIKey and restore after test
	originalKey := googleAPIKey
	defer func() { googleAPIKey = originalKey }()

	googleAPIKey = ""

	result, err := analyzeURLWithLLM("https://example.com")

	if err == nil {
		t.Error("analyzeURLWithLLM without API key: expected error, got nil")
	}

	if result != "" {
		t.Errorf("analyzeURLWithLLM without API key: result = %q, want empty string", result)
	}

	if !strings.Contains(err.Error(), "GOOGLE_API_KEY not configured") {
		t.Errorf("analyzeURLWithLLM error = %q, expected to contain 'GOOGLE_API_KEY not configured'", err.Error())
	}
}

func TestPromptTypeConstants(t *testing.T) {
	if PromptBullshit != "bullshit" {
		t.Errorf("PromptBullshit = %q, want %q", PromptBullshit, "bullshit")
	}
	if PromptPositive != "positive" {
		t.Errorf("PromptPositive = %q, want %q", PromptPositive, "positive")
	}
	if PromptNegative != "negative" {
		t.Errorf("PromptNegative = %q, want %q", PromptNegative, "negative")
	}
}

func TestLLMModelConstant(t *testing.T) {
	if llmModel == "" {
		t.Error("llmModel is empty")
	}
	if llmModel != "gemini-flash-latest" {
		t.Errorf("llmModel = %q, want %q", llmModel, "gemini-flash-latest")
	}
}

func TestTruncateStringPreservesShorterStrings(t *testing.T) {
	// Ensure strings shorter than or equal to maxLen are not modified
	testCases := []struct {
		input  string
		maxLen int
	}{
		{"", 10},
		{"a", 1},
		{"ab", 2},
		{"abc", 5},
		{"hello", 5},
		{"hello", 100},
	}

	for _, tc := range testCases {
		result := truncateString(tc.input, tc.maxLen)
		if result != tc.input {
			t.Errorf("truncateString(%q, %d) = %q, want %q (unchanged)", tc.input, tc.maxLen, result, tc.input)
		}
	}
}

func TestTruncateStringAddsSuffix(t *testing.T) {
	// Ensure truncated strings end with "..."
	result := truncateString("This is a long string", 10)

	if !strings.HasSuffix(result, "...") {
		t.Errorf("truncateString result %q does not end with '...'", result)
	}
}
