package main

import (
	"fmt"
	"strings"
)

// getOpinion analyzes a message and returns an opinion about it
func getOpinion(text string) string {
	if text == "" {
		return "No text to analyze."
	}

	// Simple sentiment analysis based on keywords
	text = strings.ToLower(text)
	
	positiveWords := []string{"good", "great", "excellent", "wonderful", "amazing", "love", "awesome", "fantastic", "perfect", "happy", "best"}
	negativeWords := []string{"bad", "terrible", "awful", "horrible", "hate", "worst", "poor", "sad", "angry", "disappointed"}
	
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		positiveCount += strings.Count(text, word)
	}
	
	for _, word := range negativeWords {
		negativeCount += strings.Count(text, word)
	}
	
	wordCount := len(strings.Fields(text))
	
	var sentiment string
	if positiveCount > negativeCount {
		sentiment = "positive 😊"
	} else if negativeCount > positiveCount {
		sentiment = "negative 😔"
	} else {
		sentiment = "neutral 😐"
	}
	
	opinion := fmt.Sprintf("📊 Opinion Analysis:\n\n"+
		"Sentiment: %s\n"+
		"Word count: %d\n"+
		"Positive indicators: %d\n"+
		"Negative indicators: %d",
		sentiment, wordCount, positiveCount, negativeCount)
	
	return opinion
}
