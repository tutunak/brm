package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
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

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Check if the message is a command
		if update.Message.IsCommand() {
			handleCommand(bot, update.Message)
		}
	}
}

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	switch message.Command() {
	case "opinion":
		handleOpinionCommand(bot, message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Available commands: /opinion")
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
	}
}

func handleOpinionCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Check if this is a reply to another message
	if message.ReplyToMessage == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please use /opinion as a reply to a message")
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
		return
	}

	// Get the text from the replied message
	originalText := message.ReplyToMessage.Text
	if originalText == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "The replied message has no text to analyze")
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
		return
	}

	// Process the message through the opinion function
	opinion := getOpinion(originalText)

	// Reply to the original message with the opinion
	msg := tgbotapi.NewMessage(message.Chat.ID, opinion)
	msg.ReplyToMessageID = message.ReplyToMessage.MessageID
	bot.Send(msg)
}
