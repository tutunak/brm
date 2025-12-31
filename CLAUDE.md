# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Run the bot locally
go run .

# Run tests
go test -v ./...

# Run a single test
go test -v -run TestExtractURL ./...

# Run tests with coverage
go test -cover ./...

# Detailed coverage report
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# Build binary
go build -o bot .

# Docker build and run
docker build -t telegram-opinion-bot .
docker run -d --name opinion-bot --env-file .env telegram-opinion-bot
```

## Environment Variables

Required (see `.env.example`):
- `TELEGRAM_BOT_TOKEN` - Telegram bot token from BotFather
- `GOOGLE_API_KEY` - Google API key for Gemini LLM
- `ALLOWED_CHAT_IDS` - Comma-separated list of authorized chat IDs

Optional:
- `GROUP_LINK` - Link shown to unauthorized users
- `EXCLUDED_USER_IDS` - Users who bypass rate limiting
- `REDIS_ADDR` - Redis/Valkey address for caching (defaults to localhost:6379)

## Architecture

Telegram bot that analyzes URLs using Google's Gemini LLM with randomized opinion tones.

### Core Flow
1. User replies to a message containing a URL with `/opinion` command
2. Bot extracts URL from the replied message (`opinion.go:extractURL`)
3. Bot sends URL to Gemini LLM with a randomly selected prompt tone (`llm.go:analyzeURLWithLLM`)
4. Response is sent back to the chat

### Key Components
- `main.go` - Bot initialization, Telegram handler, rate limiting, chat authorization
- `opinion.go` - URL extraction and response routing
- `llm.go` - Gemini API integration with prompt selection

### Prompt System
Randomly selects prompt type with different probabilities:
- Bullshit (10%) - Sarcastic/critical
- Positive (40%) - Encouraging
- Negative (50%) - Constructive criticism

### Redis Caching
- Tracks processed messages (30-day TTL) to prevent duplicate analysis
- Rate limits users to 20 opinions per 2 days (bypassed via `EXCLUDED_USER_IDS`)

### Logging
All logs are structured JSON via `logJSON()` function.

## Testing

Test files mirror source files:
- `main_test.go` - Tests for parsing, authorization, user/chat info functions
- `opinion_test.go` - Tests for URL extraction and opinion generation
- `llm_test.go` - Tests for prompt selection and string utilities

Tests use a `MockContext` struct to mock Telegram's `tele.Context` interface.
