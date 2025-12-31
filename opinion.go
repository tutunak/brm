package main

import (
	"math/rand"
	"regexp"
)

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
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	matches := urlRegex.FindStringSubmatch(text)
	
	if len(matches) > 0 {
		return matches[0]
	}
	
	return ""
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
