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

    log.Printf("Bot started successfully - @%s (ID: %d)", bot.Me.Username, bot.Me.ID)

    // Handle /opinion command
    bot.Handle("/opinion", func(c tele.Context) error {
        logRequest(c, "/opinion")
        return handleOpinionCommand(c)
    })

    // Handle unknown commands
    bot.Handle(tele.OnText, func(c tele.Context) error {
        // Only respond to other commands, not regular text
        if c.Message().Text != "" && c.Message().Text[0] == '/' {
            logRequest(c, "unknown_command")
            return c.Reply("Unknown command. Available commands: /opinion")
        }
        return nil
    })

    log.Println("Bot is running and waiting for messages...")
    bot.Start()
}

func handleOpinionCommand(c tele.Context) error {
    // Check if this is a reply to another message
    if c.Message().ReplyTo == nil {
        log.Printf("[WARN] User %s tried to use /opinion without replying to a message", getUserInfo(c))
        return c.Reply("Please use /opinion as a reply to a message")
    }

    // Get the text from the replied message
    originalText := c.Message().ReplyTo.Text
    if originalText == "" {
        log.Printf("[WARN] User %s replied to a message with no text", getUserInfo(c))
        return c.Reply("The replied message has no text to analyze")
    }

    log.Printf("[INFO] Processing opinion for message from %s in %s. Text length: %d chars", 
        getUserInfo(c), getChatInfo(c), len(originalText))

    // Process the message through the opinion function
    opinion := getOpinion(originalText)

    log.Printf("[SUCCESS] Sending opinion result to %s", getChatInfo(c))

    // Reply to the original message with the opinion
    return c.Send(opinion, &tele.SendOptions{
        ReplyTo: c.Message().ReplyTo,
    })
}

// logRequest logs information about incoming requests
func logRequest(c tele.Context, command string) {
    log.Printf("[REQUEST] Command: %s | User: %s | Chat: %s | MessageID: %d",
        command,
        getUserInfo(c),
        getChatInfo(c),
        c.Message().ID)
}

// getUserInfo returns formatted user information
func getUserInfo(c tele.Context) string {
    user := c.Sender()
    if user == nil {
        return "Unknown"
    }
    
    username := user.Username
    if username == "" {
        username = "no_username"
    }
    
    firstName := user.FirstName
    if firstName == "" {
        firstName = "Unknown"
    }
    
    return user.Username + " (" + firstName + ", ID: " + string(rune(user.ID)) + ")"
}

// getChatInfo returns formatted chat information
func getChatInfo(c tele.Context) string {
    chat := c.Chat()
    if chat == nil {
        return "Unknown"
    }
    
    chatType := string(chat.Type)
    chatTitle := chat.Title
    
    if chatTitle == "" {
        // For private chats, use username or first name
        if chat.Username != "" {
            chatTitle = "@" + chat.Username
        } else if chat.FirstName != "" {
            chatTitle = chat.FirstName
        } else {
            chatTitle = "Private Chat"
        }
    }
    
    return chatTitle + " (" + chatType + ", ID: " + string(rune(chat.ID)) + ")"
}
