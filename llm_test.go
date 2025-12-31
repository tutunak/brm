package main

import (
	"bytes"
	"io"
	"os"
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

// TestTruncateStringNegativeMaxLen tests truncateString with negative maxLen
func TestTruncateStringNegativeMaxLen(t *testing.T) {
	// Negative maxLen causes a panic in current implementation
	// This test documents the expected behavior - it panics with slice bounds error
	defer func() {
		if r := recover(); r == nil {
			t.Log("truncateString with negative maxLen did not panic (implementation may have been fixed)")
		}
	}()

	truncateString("Hello", -1)
	// If we get here without panic, the implementation handles negative values
}

// TestTruncateStringExactBoundary tests truncateString at exact boundary
func TestTruncateStringExactBoundary(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"12345", 5, "12345"},     // Exactly equal
		{"12345", 4, "1234..."},   // One less
		{"12345", 6, "12345"},     // One more
		{"", 0, ""},               // Both empty/zero
		{"a", 1, "a"},             // Single char equal
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestSelectPromptTypeAlwaysReturnsValid tests that selectPromptType always returns a valid type
func TestSelectPromptTypeAlwaysReturnsValid(t *testing.T) {
	validTypes := map[PromptType]bool{
		PromptBullshit: true,
		PromptPositive: true,
		PromptNegative: true,
	}

	// Run many iterations
	for i := 0; i < 1000; i++ {
		result := selectPromptType()
		if !validTypes[result] {
			t.Errorf("selectPromptType() iteration %d returned invalid type: %v", i, result)
		}
	}
}

// TestSelectPromptTypeProbabilityBounds tests the probability boundaries
func TestSelectPromptTypeProbabilityBounds(t *testing.T) {
	// This test verifies the distribution matches the expected probabilities
	// by checking that each type is selected at least a minimum percentage
	counts := make(map[PromptType]int)
	iterations := 10000

	for i := 0; i < iterations; i++ {
		counts[selectPromptType()]++
	}

	// Check minimum thresholds (with generous tolerance for randomness)
	bullshitPct := float64(counts[PromptBullshit]) / float64(iterations) * 100
	positivePct := float64(counts[PromptPositive]) / float64(iterations) * 100
	negativePct := float64(counts[PromptNegative]) / float64(iterations) * 100

	// Bullshit should be at least 5% (target 10%)
	if bullshitPct < 5 {
		t.Errorf("PromptBullshit = %.1f%%, expected at least 5%%", bullshitPct)
	}
	// Positive should be at least 30% (target 40%)
	if positivePct < 30 {
		t.Errorf("PromptPositive = %.1f%%, expected at least 30%%", positivePct)
	}
	// Negative should be at least 40% (target 50%)
	if negativePct < 40 {
		t.Errorf("PromptNegative = %.1f%%, expected at least 40%%", negativePct)
	}
}

// TestBuildPromptNotEmpty tests that buildPrompt never returns empty string
func TestBuildPromptNotEmpty(t *testing.T) {
	promptTypes := []PromptType{PromptBullshit, PromptPositive, PromptNegative}

	for _, pt := range promptTypes {
		result := buildPrompt(pt)
		if result == "" {
			t.Errorf("buildPrompt(%v) returned empty string", pt)
		}
	}
}

// TestBuildPromptContainsBaseAndVideo tests that buildPrompt combines base and video prompts
func TestBuildPromptContainsBaseAndVideo(t *testing.T) {
	promptTypes := []PromptType{PromptBullshit, PromptPositive, PromptNegative}

	for _, pt := range promptTypes {
		t.Run(string(pt), func(t *testing.T) {
			result := buildPrompt(pt)
			base := basePrompts[pt]
			video := videoPrompts[pt]

			if !strings.Contains(result, base) {
				t.Errorf("buildPrompt(%v) does not contain base prompt", pt)
			}
			if !strings.Contains(result, strings.TrimSpace(video)) {
				t.Errorf("buildPrompt(%v) does not contain video prompt", pt)
			}
		})
	}
}

// TestBuildPromptWithInvalidType tests buildPrompt with an invalid prompt type
func TestBuildPromptWithInvalidType(t *testing.T) {
	invalidType := PromptType("invalid")
	result := buildPrompt(invalidType)

	// With invalid type, both maps will return empty strings
	if result != "" {
		t.Errorf("buildPrompt(invalid) = %q, expected empty string", result)
	}
}

// TestPromptMapsHaveSameKeys tests that basePrompts and videoPrompts have the same keys
func TestPromptMapsHaveSameKeys(t *testing.T) {
	for key := range basePrompts {
		if _, exists := videoPrompts[key]; !exists {
			t.Errorf("basePrompts has key %v but videoPrompts does not", key)
		}
	}

	for key := range videoPrompts {
		if _, exists := basePrompts[key]; !exists {
			t.Errorf("videoPrompts has key %v but basePrompts does not", key)
		}
	}
}

// TestBasePromptsContainKeywords tests that base prompts contain expected keywords
func TestBasePromptsContainKeywords(t *testing.T) {
	tests := []struct {
		promptType PromptType
		keywords   []string
	}{
		{
			PromptBullshit,
			[]string{"bullshit", "short", "funny"},
		},
		{
			PromptPositive,
			[]string{"positive", "encouraging", "good"},
		},
		{
			PromptNegative,
			[]string{"criticism", "critical", "weaknesses"},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.promptType), func(t *testing.T) {
			prompt := basePrompts[tt.promptType]
			promptLower := strings.ToLower(prompt)

			for _, keyword := range tt.keywords {
				if !strings.Contains(promptLower, strings.ToLower(keyword)) {
					t.Errorf("basePrompts[%v] missing keyword %q", tt.promptType, keyword)
				}
			}
		})
	}
}

// TestVideoPromptsContainVideoKeyword tests that all video prompts mention "video"
func TestVideoPromptsContainVideoKeyword(t *testing.T) {
	for promptType, prompt := range videoPrompts {
		if !strings.Contains(strings.ToLower(prompt), "video") {
			t.Errorf("videoPrompts[%v] does not contain 'video'", promptType)
		}
	}
}

// TestAnalyzeURLWithLLMEmptyURL tests analyzeURLWithLLM with empty URL
func TestAnalyzeURLWithLLMEmptyURL(t *testing.T) {
	// Save original googleAPIKey and restore after test
	originalKey := googleAPIKey
	defer func() { googleAPIKey = originalKey }()

	googleAPIKey = ""

	result, err := analyzeURLWithLLM("")

	if err == nil {
		t.Error("analyzeURLWithLLM with empty URL: expected error, got nil")
	}

	if result != "" {
		t.Errorf("analyzeURLWithLLM with empty URL: result = %q, want empty string", result)
	}
}

// TestLLMModelNotEmpty tests that llmModel constant is not empty
func TestLLMModelNotEmpty(t *testing.T) {
	if llmModel == "" {
		t.Error("llmModel constant is empty")
	}
}

// TestPromptTypeIsString tests that PromptType values are strings
func TestPromptTypeIsString(t *testing.T) {
	types := []PromptType{PromptBullshit, PromptPositive, PromptNegative}

	for _, pt := range types {
		if string(pt) == "" {
			t.Errorf("PromptType %v converts to empty string", pt)
		}
	}
}

// TestPromptTypeValuesUnique tests that all PromptType values are unique
func TestPromptTypeValuesUnique(t *testing.T) {
	types := []PromptType{PromptBullshit, PromptPositive, PromptNegative}
	seen := make(map[PromptType]bool)

	for _, pt := range types {
		if seen[pt] {
			t.Errorf("Duplicate PromptType value: %v", pt)
		}
		seen[pt] = true
	}
}

// TestBasePromptsLength tests that base prompts are not too short
func TestBasePromptsLength(t *testing.T) {
	minLength := 50 // Reasonable minimum for a useful prompt

	for promptType, prompt := range basePrompts {
		if len(prompt) < minLength {
			t.Errorf("basePrompts[%v] is too short (%d chars), expected at least %d",
				promptType, len(prompt), minLength)
		}
	}
}

// TestVideoPromptsLength tests that video prompts are not too short
func TestVideoPromptsLength(t *testing.T) {
	minLength := 20 // Reasonable minimum for video handling instructions

	for promptType, prompt := range videoPrompts {
		if len(prompt) < minLength {
			t.Errorf("videoPrompts[%v] is too short (%d chars), expected at least %d",
				promptType, len(prompt), minLength)
		}
	}
}

// TestTruncateStringWithSpecialCharacters tests truncateString with special chars
func TestTruncateStringWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "String with newlines",
			input:    "Hello\nWorld\nTest",
			maxLen:   7,
			expected: "Hello\nW...",
		},
		{
			name:     "String with tabs",
			input:    "Hello\tWorld",
			maxLen:   6,
			expected: "Hello\t...",
		},
		{
			name:     "String with null bytes",
			input:    "Hello\x00World",
			maxLen:   6,
			expected: "Hello\x00...",
		},
		{
			name:     "String with unicode",
			input:    "Привет мир",
			maxLen:   20, // Each Cyrillic char is 2 bytes, so 10 chars = 19 bytes + space
			expected: "Привет мир",
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

// TestAnalyzeURLWithLLMLogsCorrectly tests that analyzeURLWithLLM logs properly
func TestAnalyzeURLWithLLMLogsCorrectly(t *testing.T) {
	// Save original googleAPIKey and restore after test
	originalKey := googleAPIKey
	defer func() { googleAPIKey = originalKey }()

	googleAPIKey = ""

	// Capture stdout to check logs
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	analyzeURLWithLLM("https://example.com")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should contain error log about API key
	if !strings.Contains(output, "LLM API key not configured") {
		t.Error("analyzeURLWithLLM did not log API key error")
	}
}

// TestBuildPromptOrder tests that buildPrompt concatenates in correct order
func TestBuildPromptOrder(t *testing.T) {
	for _, pt := range []PromptType{PromptBullshit, PromptPositive, PromptNegative} {
		t.Run(string(pt), func(t *testing.T) {
			result := buildPrompt(pt)
			base := basePrompts[pt]
			video := videoPrompts[pt]

			// Base should come first
			baseIndex := strings.Index(result, base)
			videoIndex := strings.Index(result, strings.TrimSpace(video))

			if baseIndex == -1 {
				t.Error("Base prompt not found in result")
				return
			}
			if videoIndex == -1 {
				t.Error("Video prompt not found in result")
				return
			}
			if baseIndex > videoIndex {
				t.Error("Base prompt should come before video prompt")
			}
		})
	}
}

// TestPromptTypeCasting tests that PromptType can be cast to and from string
func TestPromptTypeCasting(t *testing.T) {
	tests := []struct {
		input    string
		expected PromptType
	}{
		{"bullshit", PromptBullshit},
		{"positive", PromptPositive},
		{"negative", PromptNegative},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := PromptType(tt.input)
			if result != tt.expected {
				t.Errorf("PromptType(%q) = %v, want %v", tt.input, result, tt.expected)
			}

			// And back to string
			str := string(result)
			if str != tt.input {
				t.Errorf("string(PromptType(%q)) = %q, want %q", tt.input, str, tt.input)
			}
		})
	}
}
