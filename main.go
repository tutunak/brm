package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/joho/godotenv"
    tele "gopkg.in/telebot.v3"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        logJSON("info", "No .env file found, using system environment variables", nil)
    }

    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    if botToken == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
    }

    // Load allowed chat IDs
    allowedChatsStr := os.Getenv("ALLOWED_CHAT_IDS")
    if allowedChatsStr == "" {
        log.Fatal("ALLOWED_CHAT_IDS environment variable is required (comma-separated list of chat IDs)")
    }

    groupLink := os.Getenv("GROUP_LINK")
    if groupLink == "" {
        groupLink = "your group"
    }

    allowedChatIDs := parseAllowedChatIDs(allowedChatsStr)
    logJSON("info", "Allowed chat IDs loaded", map[string]interface{}{
        "allowed_chats": allowedChatIDs,
        "group_link":    groupLink,
    })

    pref := tele.Settings{
        Token:  botToken,
        Poller: &tele.LongPoller{Timeout: 10 * time.Second},
    }

    bot, err := tele.NewBot(pref)
    if err != nil {
        log.Fatal(err)
    }

    logJSON("info", "Bot started successfully", map[string]interface{}{
        "bot_username": bot.Me.Username,
        "bot_id":       bot.Me.ID,
    })

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

    logJSON("info", "Bot is running and waiting for messages", nil)
    bot.Start()
}

func handleOpinionCommand(c tele.Context) error {
    // Check if in allowed group
    if !isAllowedChat(c) {
        chatType := string(c.Chat().Type)
        var message string
        
        if chatType == "private" {
            message = fmt.Sprintf("🤖 This bot is in early alpha and works only in the group: %s", os.Getenv("GROUP_LINK"))
        } else {
            message = fmt.Sprintf("🤖 This bot is in early alpha and works only in authorized groups. Join us at: %s", os.Getenv("GROUP_LINK"))
        }
        
        logJSON("warn", "Unauthorized chat access attempt", map[string]interface{}{
            "user":      getUserInfo(c),
            "chat":      getChatInfo(c),
            "chat_type": chatType,
        })
        
        return c.Reply(message)
    }

    // Check if this is a reply to another message
    if c.Message().ReplyTo == nil {
        logJSON("warn", "Command used without reply", map[string]interface{}{
            "user":    getUserInfo(c),
            "chat":    getChatInfo(c),
            "command": "/opinion",
        })
        return c.Reply("Please use /opinion as a reply to a message")
    }

    // Get the text from the replied message
    originalText := c.Message().ReplyTo.Text
    if originalText == "" {
        logJSON("warn", "Replied message has no text", map[string]interface{}{
            "user": getUserInfo(c),
            "chat": getChatInfo(c),
        })
        return c.Reply("The replied message has no text to analyze")
    }

    logJSON("info", "Processing opinion request", map[string]interface{}{
        "user":        getUserInfo(c),
        "chat":        getChatInfo(c),
        "text_length": len(originalText),
    })

    // Process the message through the opinion function
    opinion := getOpinion(originalText)

    logJSON("success", "Opinion sent successfully", map[string]interface{}{
        "user": getUserInfo(c),
        "chat": getChatInfo(c),
    })

    // Reply to the original message with the opinion
    return c.Send(opinion, &tele.SendOptions{
        ReplyTo: c.Message().ReplyTo,
    })
}

// logRequest logs information about incoming requests
func logRequest(c tele.Context, command string) {
    logJSON("request", "Command received", map[string]interface{}{
        "command":    command,
        "user":       getUserInfo(c),
        "chat":       getChatInfo(c),
        "message_id": c.Message().ID,
    })
}

// logJSON outputs structured JSON logs
func logJSON(level string, message string, data map[string]interface{}) {
    logEntry := map[string]interface{}{
        "timestamp": time.Now().Format(time.RFC3339),
        "level":     level,
        "message":   message,
    }
    
    if data != nil {
        for k, v := range data {
            logEntry[k] = v
        }
    }
    
    jsonBytes, err := json.Marshal(logEntry)
    if err != nil {
        log.Printf("Error marshaling log: %v", err)
        return
    }
    
    fmt.Println(string(jsonBytes))
}

// parseAllowedChatIDs parses comma-separated chat IDs from environment variable
func parseAllowedChatIDs(chatsStr string) []int64 {
    parts := strings.Split(chatsStr, ",")
    var chatIDs []int64
    
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if part == "" {
            continue
        }
        
        chatID, err := strconv.ParseInt(part, 10, 64)
        if err != nil {
            log.Printf("Warning: invalid chat ID '%s': %v", part, err)
            continue
        }
        
        chatIDs = append(chatIDs, chatID)
    }
    
    return chatIDs
}

// isAllowedChat checks if the current chat is in the allowed list
func isAllowedChat(c tele.Context) bool {
    allowedChatsStr := os.Getenv("ALLOWED_CHAT_IDS")
    allowedChatIDs := parseAllowedChatIDs(allowedChatsStr)
    
    currentChatID := c.Chat().ID
    
    for _, allowedID := range allowedChatIDs {
        if currentChatID == allowedID {
            return true
        }
    }
    
    return false
}

// getUserInfo returns formatted user information
func getUserInfo(c tele.Context) map[string]interface{} {
    user := c.Sender()
    if user == nil {
        return map[string]interface{}{
            "username":   "unknown",
            "first_name": "Unknown",
            "user_id":    0,
        }
    }
    
    username := user.Username
    if username == "" {
        username = "no_username"
    }
    
    firstName := user.FirstName
    if firstName == "" {
        firstName = "Unknown"
    }
    
    return map[string]interface{}{
        "username":   username,
        "first_name": firstName,
        "user_id":    user.ID,
    }
}

// getChatInfo returns formatted chat information
func getChatInfo(c tele.Context) map[string]interface{} {
    chat := c.Chat()
    if chat == nil {
        return map[string]interface{}{
            "chat_title": "Unknown",
            "chat_type":  "unknown",
            "chat_id":    0,
        }
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
    
    return map[string]interface{}{
        "chat_title": chatTitle,
        "chat_type":  chatType,
        "chat_id":    chat.ID,
    }
}
