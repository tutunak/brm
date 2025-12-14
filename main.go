package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Me.Username)

	// Handle /opinion command
	bot.Handle("/opinion", func(c tele.Context) error {
		return handleOpinionCommand(c)
	})

	// Handle unknown commands
	bot.Handle(tele.OnText, func(c tele.Context) error {
		// Only respond to other commands, not regular text
		if c.Message().Text != "" && c.Message().Text[0] == '/' {
			return c.Reply("Unknown command. Available commands: /opinion")
		}
		return nil
	})

	log.Println("Bot is running...")
	bot.Start()
}

func handleOpinionCommand(c tele.Context) error {
	// Check if this is a reply to another message
	if c.Message().ReplyTo == nil {
		return c.Reply("Please use /opinion as a reply to a message")
	}

	// Get the text from the replied message
	originalText := c.Message().ReplyTo.Text
	if originalText == "" {
		return c.Reply("The replied message has no text to analyze")
	}

	// Process the message through the opinion function
	opinion := getOpinion(originalText)

	// Reply to the original message with the opinion
	return c.Send(opinion, &tele.SendOptions{
		ReplyTo: c.Message().ReplyTo,
	})
}
