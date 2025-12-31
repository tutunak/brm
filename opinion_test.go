package main

import (
	"strings"
	"testing"
)

func TestExtractURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple HTTP URL",
			input:    "Check this out http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "Simple HTTPS URL",
			input:    "Check this out https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "URL with path",
			input:    "Visit https://github.com/user/repo for more info",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "URL with query parameters",
			input:    "Link: https://example.com/page?param=value&other=123",
			expected: "https://example.com/page?param=value&other=123",
		},
		{
			name:     "URL with fragment",
			input:    "See https://docs.example.com/guide#section",
			expected: "https://docs.example.com/guide#section",
		},
		{
			name:     "Multiple URLs - returns first",
			input:    "First https://first.com then https://second.com",
			expected: "https://first.com",
		},
		{
			name:     "No URL in text",
			input:    "This is just plain text without any links",
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "URL only",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "URL with port",
			input:    "Server at http://localhost:8080/api",
			expected: "http://localhost:8080/api",
		},
		{
			name:     "Complex GitHub URL",
			input:    "Check https://github.com/anthropics/claude-code/blob/main/README.md",
			expected: "https://github.com/anthropics/claude-code/blob/main/README.md",
		},
		{
			name:     "YouTube URL",
			input:    "Watch this https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:     "URL at start of text",
			input:    "https://example.com is a great site",
			expected: "https://example.com",
		},
		{
			name:     "URL at end of text",
			input:    "Great site: https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "FTP URL (not matched)",
			input:    "Download from ftp://files.example.com",
			expected: "",
		},
		{
			name:     "Malformed URL without protocol",
			input:    "Visit example.com for more",
			expected: "",
		},
		{
			name:     "Text with newlines and URL",
			input:    "Line one\nhttps://example.com\nLine three",
			expected: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractURL(tt.input)
			if result != tt.expected {
				t.Errorf("extractURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetOpinionEmptyText(t *testing.T) {
	result, success := getOpinion("")

	if success {
		t.Error("getOpinion(\"\") success = true, want false")
	}
	if result != "No text to analyze." {
		t.Errorf("getOpinion(\"\") = %q, want %q", result, "No text to analyze.")
	}
}

func TestGetOpinionNoURL(t *testing.T) {
	// Test various inputs without URLs
	inputs := []string{
		"This is just plain text",
		"Hello world!",
		"Some random message without any links",
		"Check out example.com (no protocol)",
	}

	// Known refusal responses from getRandomRefusalResponse
	validResponses := []string{
		"I'm tired 😴",
		"I don't want to talk 😤",
		"NO 😠",
		"Not today 😑",
		"Leave me alone 🙄",
		"I'm not in the mood 😒",
		"Go away 😡",
		"Seriously? 🤨",
		"Don't bother me 💢",
		"Ask someone else 😾",
		"I refuse 🚫",
		"Absolutely not 😤",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			result, success := getOpinion(input)

			if success {
				t.Errorf("getOpinion(%q) success = true, want false", input)
			}

			// Check that result is one of the valid refusal responses
			found := false
			for _, valid := range validResponses {
				if result == valid {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("getOpinion(%q) = %q, not a valid refusal response", input, result)
			}
		})
	}
}

func TestGetRandomRefusalResponse(t *testing.T) {
	validResponses := map[string]bool{
		"I'm tired 😴":         true,
		"I don't want to talk 😤": true,
		"NO 😠":                 true,
		"Not today 😑":          true,
		"Leave me alone 🙄":     true,
		"I'm not in the mood 😒": true,
		"Go away 😡":            true,
		"Seriously? 🤨":         true,
		"Don't bother me 💢":    true,
		"Ask someone else 😾":   true,
		"I refuse 🚫":           true,
		"Absolutely not 😤":     true,
	}

	// Call multiple times to check randomness and validity
	for i := 0; i < 50; i++ {
		result := getRandomRefusalResponse()
		if !validResponses[result] {
			t.Errorf("getRandomRefusalResponse() = %q, not a valid response", result)
		}
	}
}

func TestGetRandomRefusalResponseDistribution(t *testing.T) {
	// Run many times to ensure we get multiple different responses (basic randomness check)
	responseCounts := make(map[string]int)
	iterations := 100

	for i := 0; i < iterations; i++ {
		response := getRandomRefusalResponse()
		responseCounts[response]++
	}

	// Should have at least 3 different responses in 100 iterations
	if len(responseCounts) < 3 {
		t.Errorf("getRandomRefusalResponse() only produced %d unique responses in %d iterations, expected more variety",
			len(responseCounts), iterations)
	}
}

func TestGetOpinionWithURL(t *testing.T) {
	// Save original googleAPIKey and restore after test
	originalKey := googleAPIKey
	defer func() { googleAPIKey = originalKey }()

	// Test with no API key configured
	googleAPIKey = ""

	result, success := getOpinion("Check out https://example.com")

	// Without API key, it should fail with the tired response
	if success {
		t.Error("getOpinion with URL but no API key: success = true, want false")
	}
	if result != "I'm tired dude, next time 😴" {
		t.Errorf("getOpinion with URL but no API key = %q, want %q", result, "I'm tired dude, next time 😴")
	}
}

func TestProcessURLWithoutAPIKey(t *testing.T) {
	// Save original googleAPIKey and restore after test
	originalKey := googleAPIKey
	defer func() { googleAPIKey = originalKey }()

	googleAPIKey = ""

	result, success := processURL("https://example.com")

	if success {
		t.Error("processURL without API key: success = true, want false")
	}
	if result != "I'm tired dude, next time 😴" {
		t.Errorf("processURL without API key = %q, want %q", result, "I'm tired dude, next time 😴")
	}
}

func TestExtractURLEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL with special characters in path",
			input:    "See https://example.com/path/with%20spaces",
			expected: "https://example.com/path/with%20spaces",
		},
		{
			name:     "URL with subdomain",
			input:    "Visit https://api.example.com/v1/users",
			expected: "https://api.example.com/v1/users",
		},
		{
			name:     "URL with authentication",
			input:    "https://user:pass@example.com/path",
			expected: "https://user:pass@example.com/path",
		},
		{
			name:     "Very long URL",
			input:    "Link: https://example.com/very/long/path/that/goes/on/and/on/with/many/segments",
			expected: "https://example.com/very/long/path/that/goes/on/and/on/with/many/segments",
		},
		{
			name:     "URL surrounded by parentheses",
			input:    "Check this (https://example.com)",
			expected: "https://example.com",
		},
		{
			name:     "URL followed by punctuation",
			input:    "Visit https://example.com.",
			expected: "https://example.com",
		},
		{
			name:     "URL in markdown format",
			input:    "[Link](https://example.com)",
			expected: "https://example.com",
		},
		{
			name:     "Localhost URL",
			input:    "Test at http://localhost:3000/api/test",
			expected: "http://localhost:3000/api/test",
		},
		{
			name:     "IP address URL",
			input:    "Server at http://192.168.1.1:8080",
			expected: "http://192.168.1.1:8080",
		},
		{
			name:     "URL followed by comma",
			input:    "Visit https://example.com, then proceed",
			expected: "https://example.com",
		},
		{
			name:     "URL followed by exclamation mark",
			input:    "Check out https://awesome-site.com!",
			expected: "https://awesome-site.com",
		},
		{
			name:     "URL followed by question mark",
			input:    "Have you seen https://example.com?",
			expected: "https://example.com",
		},
		{
			name:     "URL followed by semicolon",
			input:    "First https://example.com; then next",
			expected: "https://example.com",
		},
		{
			name:     "URL with multiple trailing punctuation",
			input:    "Amazing site: https://example.com!.",
			expected: "https://example.com",
		},
		{
			name:     "URL in square brackets",
			input:    "See [https://docs.example.com]",
			expected: "https://docs.example.com",
		},
		{
			name:     "URL in curly braces",
			input:    "Template: {https://api.example.com}",
			expected: "https://api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractURL(tt.input)
			if result != tt.expected {
				t.Errorf("extractURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetOpinionIntegration(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectSuccess   bool
		expectContains  string
	}{
		{
			name:           "Empty input",
			input:          "",
			expectSuccess:  false,
			expectContains: "No text to analyze",
		},
		{
			name:           "Text without URL",
			input:          "Just some regular text here",
			expectSuccess:  false,
			expectContains: "", // Will be a random refusal
		},
		{
			name:           "Whitespace only",
			input:          "   \t\n   ",
			expectSuccess:  false,
			expectContains: "", // Will be a random refusal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, success := getOpinion(tt.input)

			if success != tt.expectSuccess {
				t.Errorf("getOpinion(%q) success = %v, want %v", tt.input, success, tt.expectSuccess)
			}

			if tt.expectContains != "" && !strings.Contains(result, tt.expectContains) {
				t.Errorf("getOpinion(%q) = %q, want to contain %q", tt.input, result, tt.expectContains)
			}
		})
	}
}

// TestTrimTrailingPunctuation tests the trimTrailingPunctuation function
func TestTrimTrailingPunctuation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No trailing punctuation",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "Trailing period",
			input:    "https://example.com.",
			expected: "https://example.com",
		},
		{
			name:     "Trailing comma",
			input:    "https://example.com,",
			expected: "https://example.com",
		},
		{
			name:     "Trailing closing parenthesis",
			input:    "https://example.com)",
			expected: "https://example.com",
		},
		{
			name:     "Trailing closing bracket",
			input:    "https://example.com]",
			expected: "https://example.com",
		},
		{
			name:     "Trailing closing brace",
			input:    "https://example.com}",
			expected: "https://example.com",
		},
		{
			name:     "Trailing semicolon",
			input:    "https://example.com;",
			expected: "https://example.com",
		},
		{
			name:     "Trailing exclamation mark",
			input:    "https://example.com!",
			expected: "https://example.com",
		},
		{
			name:     "Trailing question mark",
			input:    "https://example.com?",
			expected: "https://example.com",
		},
		{
			name:     "Trailing colon",
			input:    "https://example.com:",
			expected: "https://example.com",
		},
		{
			name:     "Multiple trailing punctuation",
			input:    "https://example.com.)!",
			expected: "https://example.com",
		},
		{
			name:     "Mixed trailing punctuation",
			input:    "https://example.com!?.",
			expected: "https://example.com",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only punctuation",
			input:    ".,!?",
			expected: "",
		},
		{
			name:     "URL with query parameter ending with valid char",
			input:    "https://example.com/path?foo=bar",
			expected: "https://example.com/path?foo=bar",
		},
		{
			name:     "URL with fragment",
			input:    "https://example.com/path#section",
			expected: "https://example.com/path#section",
		},
		{
			name:     "URL with path ending in slash",
			input:    "https://example.com/path/",
			expected: "https://example.com/path/",
		},
		{
			name:     "URL with port",
			input:    "https://example.com:8080",
			expected: "https://example.com:8080",
		},
		{
			name:     "All punctuation chars sequentially",
			input:    "test.,)];!?:",
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimTrailingPunctuation(tt.input)
			if result != tt.expected {
				t.Errorf("trimTrailingPunctuation(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestContainsChar tests the containsChar function
func TestContainsChar(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		c        byte
		expected bool
	}{
		{
			name:     "Character present at start",
			s:        "abc",
			c:        'a',
			expected: true,
		},
		{
			name:     "Character present in middle",
			s:        "abc",
			c:        'b',
			expected: true,
		},
		{
			name:     "Character present at end",
			s:        "abc",
			c:        'c',
			expected: true,
		},
		{
			name:     "Character not present",
			s:        "abc",
			c:        'd',
			expected: false,
		},
		{
			name:     "Empty string",
			s:        "",
			c:        'a',
			expected: false,
		},
		{
			name:     "Single character string - match",
			s:        "a",
			c:        'a',
			expected: true,
		},
		{
			name:     "Single character string - no match",
			s:        "a",
			c:        'b',
			expected: false,
		},
		{
			name:     "Punctuation character present",
			s:        ".,!?",
			c:        '.',
			expected: true,
		},
		{
			name:     "Punctuation character not present",
			s:        ".,!?",
			c:        ';',
			expected: false,
		},
		{
			name:     "Space character",
			s:        "hello world",
			c:        ' ',
			expected: true,
		},
		{
			name:     "Newline character",
			s:        "hello\nworld",
			c:        '\n',
			expected: true,
		},
		{
			name:     "Tab character",
			s:        "hello\tworld",
			c:        '\t',
			expected: true,
		},
		{
			name:     "Numeric character",
			s:        "abc123",
			c:        '1',
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsChar(tt.s, tt.c)
			if result != tt.expected {
				t.Errorf("containsChar(%q, %q) = %v, want %v", tt.s, tt.c, result, tt.expected)
			}
		})
	}
}

// TestExtractURLWithQueryParams tests URL extraction with various query parameters
func TestExtractURLWithQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL with single query param",
			input:    "Link: https://example.com?foo=bar",
			expected: "https://example.com?foo=bar",
		},
		{
			name:     "URL with multiple query params",
			input:    "Link: https://example.com?foo=bar&baz=qux&num=123",
			expected: "https://example.com?foo=bar&baz=qux&num=123",
		},
		{
			name:     "URL with encoded query param",
			input:    "Link: https://example.com?q=hello%20world",
			expected: "https://example.com?q=hello%20world",
		},
		{
			name:     "URL with query and fragment",
			input:    "Link: https://example.com?foo=bar#section",
			expected: "https://example.com?foo=bar#section",
		},
		{
			name:     "URL with empty query param value",
			input:    "Link: https://example.com?foo=",
			expected: "https://example.com?foo=",
		},
		{
			name:     "URL with query param without value",
			input:    "Link: https://example.com?flag",
			expected: "https://example.com?flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractURL(tt.input)
			if result != tt.expected {
				t.Errorf("extractURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestExtractURLProtocols tests URL extraction with different protocols
func TestExtractURLProtocols(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTTP protocol",
			input:    "Visit http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "HTTPS protocol",
			input:    "Visit https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "FTP protocol (not supported)",
			input:    "Download ftp://files.example.com",
			expected: "",
		},
		{
			name:     "FILE protocol (not supported)",
			input:    "Open file:///path/to/file",
			expected: "",
		},
		{
			name:     "MAILTO (not supported)",
			input:    "Contact mailto:test@example.com",
			expected: "",
		},
		{
			name:     "SSH (not supported)",
			input:    "Connect ssh://user@host.com",
			expected: "",
		},
		{
			name:     "Mixed case HTTP",
			input:    "Visit HTTP://example.com",
			expected: "",
		},
		{
			name:     "Mixed case HTTPS",
			input:    "Visit HTTPS://example.com",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractURL(tt.input)
			if result != tt.expected {
				t.Errorf("extractURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestURLRegexCompiled tests that the URL regex is properly compiled
func TestURLRegexCompiled(t *testing.T) {
	if urlRegex == nil {
		t.Error("urlRegex is nil, expected to be compiled")
	}
}

// TestGetOpinionVariousURLTypes tests getOpinion with different URL types
func TestGetOpinionVariousURLTypes(t *testing.T) {
	// Save original googleAPIKey and restore after test
	originalKey := googleAPIKey
	defer func() { googleAPIKey = originalKey }()
	googleAPIKey = ""

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "GitHub URL",
			input: "Check https://github.com/user/repo",
		},
		{
			name:  "YouTube URL",
			input: "Watch https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:  "Twitter URL",
			input: "See https://twitter.com/user/status/123",
		},
		{
			name:  "Medium URL",
			input: "Read https://medium.com/@user/article",
		},
		{
			name:  "Localhost URL",
			input: "Test at http://localhost:3000/api",
		},
		{
			name:  "IP address URL",
			input: "Server at http://192.168.1.1:8080/admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, success := getOpinion(tt.input)

			// Without API key, it should fail with tired message
			if success {
				t.Errorf("getOpinion(%q) with no API key: success = true, want false", tt.input)
			}
			if result != "I'm tired dude, next time 😴" {
				t.Errorf("getOpinion(%q) = %q, want %q", tt.input, result, "I'm tired dude, next time 😴")
			}
		})
	}
}

// TestProcessURLReturnsError tests processURL error handling
func TestProcessURLReturnsError(t *testing.T) {
	// Save original googleAPIKey and restore after test
	originalKey := googleAPIKey
	defer func() { googleAPIKey = originalKey }()
	googleAPIKey = ""

	urls := []string{
		"https://example.com",
		"https://github.com/test/repo",
		"http://localhost:8080",
		"https://api.example.com/v1/data",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			result, success := processURL(url)

			if success {
				t.Errorf("processURL(%q) without API key: success = true, want false", url)
			}
			if result != "I'm tired dude, next time 😴" {
				t.Errorf("processURL(%q) without API key = %q, want %q", url, result, "I'm tired dude, next time 😴")
			}
		})
	}
}

// TestRefusalResponsesHaveEmojis tests that all refusal responses contain emojis
func TestRefusalResponsesHaveEmojis(t *testing.T) {
	// Common emoji patterns (expanded ranges)
	hasEmoji := func(s string) bool {
		for _, r := range s {
			// Check for common emoji ranges
			if r >= 0x1F600 && r <= 0x1F64F { // Emoticons
				return true
			}
			if r >= 0x1F300 && r <= 0x1F5FF { // Misc Symbols and Pictographs
				return true
			}
			if r >= 0x1F680 && r <= 0x1F6FF { // Transport and Map Symbols
				return true
			}
			if r >= 0x1F900 && r <= 0x1F9FF { // Supplemental Symbols and Pictographs
				return true
			}
			if r >= 0x1FA00 && r <= 0x1FA6F { // Chess Symbols
				return true
			}
			if r >= 0x2600 && r <= 0x26FF { // Misc symbols
				return true
			}
			if r >= 0x2700 && r <= 0x27BF { // Dingbats
				return true
			}
		}
		return false
	}

	// Test multiple times to check various responses
	for i := 0; i < 50; i++ {
		response := getRandomRefusalResponse()
		if !hasEmoji(response) {
			t.Errorf("getRandomRefusalResponse() = %q, expected to contain an emoji", response)
		}
	}
}
