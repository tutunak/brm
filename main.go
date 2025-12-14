package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/joho/godotenv"
    "github.com/redis/go-redis/v9"
    tele "gopkg.in/telebot.v3"
)

var redisClient *redis.Client

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        logJSON("info", "No .env file found, using system environment variables", nil)
    }

    // Initialize Redis client
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }
    
    redisClient = redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: os.Getenv("REDIS_PASSWORD"),
        DB:       0,
    })
    
    ctx := context.Background()
    if err := redisClient.Ping(ctx).Err(); err != nil {
        logJSON("warn", "Redis connection failed, continuing without cache", map[string]interface{}{
            "error": err.Error(),
        })
        redisClient = nil
    } else {
        logJSON("info", "Redis connected successfully", map[string]interface{}{
            "address": redisAddr,
        })
    }

    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    if botToken == "" {
        logFatal("TELEGRAM_BOT_TOKEN environment variable is required", nil)
    }

    // Load Google API key for LLM
    googleAPIKey = os.Getenv("GOOGLE_API_KEY")
    if googleAPIKey == "" {
        logJSON("warn", "GOOGLE_API_KEY not set, URL analysis will be disabled", nil)
    }

    // Load allowed chat IDs
    allowedChatsStr := os.Getenv("ALLOWED_CHAT_IDS")
    if allowedChatsStr == "" {
        logFatal("ALLOWED_CHAT_IDS environment variable is required", map[string]interface{}{
            "hint": "Provide comma-separated list of chat IDs",
        })
    }

    groupLink := os.Getenv("GROUP_LINK")
    if groupLink == "" {
        groupLink = "your group"
    }

    allowedChatIDs := parseAllowedChatIDs(allowedChatsStr)
    
    // Load excluded user IDs (users who bypass rate limiting)
    excludedUsersStr := os.Getenv("EXCLUDED_USER_IDS")
    excludedUserIDs := parseExcludedUserIDs(excludedUsersStr)
    
    logJSON("info", "Configuration loaded", map[string]interface{}{
        "allowed_chats":   allowedChatIDs,
        "excluded_users":  excludedUserIDs,
        "group_link":      groupLink,
    })

    pref := tele.Settings{
        Token:  botToken,
        Poller: &tele.LongPoller{Timeout: 10 * time.Second},
    }

    bot, err := tele.NewBot(pref)
    if err != nil {
        logFatal("Failed to create bot", map[string]interface{}{
            "error": err.Error(),
        })
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

    // Check if we've already processed this message
    messageID := c.Message().ReplyTo.ID
    userID := c.Sender().ID
    ctx := context.Background()
    
    alreadyProcessed := false
    if redisClient != nil {
        cacheKey := fmt.Sprintf("opinion:%d:%d", c.Chat().ID, messageID)
        exists, err := redisClient.Exists(ctx, cacheKey).Result()
        if err == nil && exists > 0 {
            logJSON("info", "Duplicate opinion request detected", map[string]interface{}{
                "user":       getUserInfo(c),
                "chat":       getChatInfo(c),
                "message_id": messageID,
            })
            alreadyProcessed = true
            return c.Reply("I've already answered, try to use search")
        }
    }
    
    // Rate limiting: only apply to NEW messages (not already processed) and non-excluded users
    if !alreadyProcessed && redisClient != nil && !isExcludedUser(userID) {
        rateLimitKey := fmt.Sprintf("ratelimit:%d", userID)
        now := time.Now()
        twoDaysAgo := now.Add(-48 * time.Hour)
        
        // Remove old entries (older than 2 days)
        redisClient.ZRemRangeByScore(ctx, rateLimitKey, "0", fmt.Sprintf("%d", twoDaysAgo.Unix()))
        
        // Count recent attempts
        count, err := redisClient.ZCount(ctx, rateLimitKey, fmt.Sprintf("%d", twoDaysAgo.Unix()), "+inf").Result()
        if err == nil && count >= 5 {
            logJSON("warn", "Rate limit exceeded", map[string]interface{}{
                "user":  getUserInfo(c),
                "chat":  getChatInfo(c),
                "count": count,
            })
            return c.Reply("⚠️ You've reached the limit of 5 opinions per 2 days for new messages. Already analyzed messages can still be searched.")
        }
        
        // Add current attempt to rate limit tracking
        redisClient.ZAdd(ctx, rateLimitKey, redis.Z{
            Score:  float64(now.Unix()),
            Member: fmt.Sprintf("%d:%d", c.Chat().ID, messageID),
        })
        // Set expiration to 2 days
        redisClient.Expire(ctx, rateLimitKey, 48*time.Hour)
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
    
    // Check if URL was found in the message
    hasURL := extractURL(originalText) != ""

    // Store in Redis that we've processed this message (only if URL was found)
    if hasURL && redisClient != nil {
        cacheKey := fmt.Sprintf("opinion:%d:%d", c.Chat().ID, messageID)
        // Store for 30 days
        err := redisClient.Set(ctx, cacheKey, time.Now().Unix(), 30*24*time.Hour).Err()
        if err != nil {
            logJSON("warn", "Failed to cache opinion result", map[string]interface{}{
                "error": err.Error(),
            })
        }
    }

    logJSON("success", "Opinion sent successfully", map[string]interface{}{
        "user": getUserInfo(c),
        "chat": getChatInfo(c),
    })

    // Reply to the original message if URL found, otherwise reply to command message
    replyTo := c.Message().ReplyTo
    if !hasURL {
        replyTo = c.Message()
    }
    
    return c.Send(opinion, &tele.SendOptions{
        ReplyTo:           replyTo,
        DisableWebPagePreview: true,
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

// logFatal outputs structured JSON fatal log and exits the program
func logFatal(message string, data map[string]interface{}) {
    logEntry := map[string]interface{}{
        "timestamp": time.Now().Format(time.RFC3339),
        "level":     "fatal",
        "message":   message,
    }
    
    if data != nil {
        for k, v := range data {
            logEntry[k] = v
        }
    }
    
    jsonBytes, err := json.Marshal(logEntry)
    if err != nil {
        log.Fatalf("Error marshaling fatal log: %v", err)
    }
    
    fmt.Println(string(jsonBytes))
    os.Exit(1)
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

// parseExcludedUserIDs parses comma-separated user IDs from environment variable
func parseExcludedUserIDs(usersStr string) []int64 {
    if usersStr == "" {
        return []int64{}
    }
    
    parts := strings.Split(usersStr, ",")
    var userIDs []int64
    
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if part == "" {
            continue
        }
        
        userID, err := strconv.ParseInt(part, 10, 64)
        if err != nil {
            log.Printf("Warning: invalid excluded user ID '%s': %v", part, err)
            continue
        }
        
        userIDs = append(userIDs, userID)
    }
    
    return userIDs
}

// isExcludedUser checks if the user is in the excluded list (bypasses rate limiting)
func isExcludedUser(userID int64) bool {
    excludedUsersStr := os.Getenv("EXCLUDED_USER_IDS")
    excludedUserIDs := parseExcludedUserIDs(excludedUsersStr)
    
    for _, excludedID := range excludedUserIDs {
        if userID == excludedID {
            return true
        }
    }
    
    return false
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
