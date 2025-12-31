package main

import (
	"math/rand"
	"regexp"
)

var urlRegex = regexp.MustCompile(`https?://[^\s]+`)

// getOpinion analyzes a message and returns an opinion about it
// Returns the opinion and a boolean indicating if processing was successful
func getOpinion(text string) (string, bool) {
	if text == "" {
		return "No text to analyze.", false
	}

	// Extract URL from the message
	url := extractURL(text)
	
	if url == "" {
		// No URL found - return random angry/tired response
		return getRandomRefusalResponse(), false
	}
	
	// URL found - process it
	return processURL(url)
}

// extractURL extracts the first URL from the text
func extractURL(text string) string {
	// Regex to match URLs
	matches := urlRegex.FindStringSubmatch(text)
	
	if len(matches) > 0 {
		url := matches[0]
		// Trim common trailing punctuation that's not part of the URL
		url = trimTrailingPunctuation(url)
		return url
	}
	
	return ""
}

// trimTrailingPunctuation removes trailing punctuation characters that are
// commonly not part of URLs (periods, commas, parentheses, brackets, etc.)
func trimTrailingPunctuation(url string) string {
	// Common punctuation marks that often appear after URLs in text
	const trailingPunctuation = ".,)]};!?:"
	
	for len(url) > 0 {
		lastChar := url[len(url)-1]
		if containsChar(trailingPunctuation, lastChar) {
			url = url[:len(url)-1]
		} else {
			break
		}
	}
	return url
}

// containsChar checks if a string contains a specific byte character
func containsChar(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}

// processURL processes the URL (currently does nothing)
func processURL(url string) (string, bool) {
	// Call the LLM to analyze the URL
	analysis, err := analyzeURLWithLLM(url)
	if err != nil {
		return "I'm tired dude, next time 😴", false
	}
	
	return analysis, true
}

// getRandomRefusalResponse returns a random refusal/angry response
func getRandomRefusalResponse() string {
	responses := []string{
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
	
	return responses[rand.Intn(len(responses))]
}
