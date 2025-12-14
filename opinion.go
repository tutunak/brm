package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"
)

// getOpinion analyzes a message and returns an opinion about it
func getOpinion(text string) string {
	if text == "" {
		return "No text to analyze."
	}

	// Extract URL from the message
	url := extractURL(text)
	
	if url == "" {
		// No URL found - return random angry/tired response
		return getRandomRefusalResponse()
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
func processURL(url string) string {
	// Call the LLM to analyze the URL
	analysis, err := analyzeURLWithLLM(url)
	if err != nil {
		return fmt.Sprintf("❌ Failed to analyze URL: %v", err)
	}
	
	return analysis
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
	
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
	
	return responses[rand.Intn(len(responses))]
}
