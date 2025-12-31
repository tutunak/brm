# Gemini Context File (GEMINI.md)

This file provides context for AI agents working on the `brm` project.

## Project Overview

`brm` is a Telegram bot written in Go that provides "opinions" on URLs shared in chat messages. It leverages Google's Gemini LLM to analyze content and generates responses with varying tones (Positive, Negative, or "Bullshit"). The system includes Redis-based caching and rate limiting.

**Key Technologies:**
*   **Language:** Go (v1.24)
*   **Frameworks/Libraries:**
    *   `gopkg.in/telebot.v3` (Telegram Bot API)
    *   `google.golang.org/genai` (Google Gemini API)
    *   `github.com/redis/go-redis/v9` (Redis client)
*   **Infrastructure:** Docker, Docker Compose, Valkey (Redis alternative)

## Architecture

### Core Components

1.  **Entry Point (`main.go`):**
    *   Initializes the Telegram bot and Redis connection.
    *   Handles the `/opinion` command.
    *   Implements rate limiting (5 requests/day for non-excluded users) and authorization (allowed chat IDs).
    *   Uses structured JSON logging.

2.  **LLM Integration (`llm.go`):**
    *   Interacts with the Google Gemini API (`gemini-flash-latest`).
    *   **Prompt System:** Randomly selects a persona/tone for the response:
        *   **Bullshit (10%):** Sarcastic, dismissive.
        *   **Positive (40%):** Encouraging, highlights good aspects.
        *   **Negative (50%):** Critical, constructive.
    *   Streaming response handling.

3.  **Content Extraction (`opinion.go`):**
    *   Extracts URLs from replied messages using Regex.
    *   If no URL is found, returns a random "refusal" message (e.g., "I'm tired").

### Data Flow

1.  User replies to a message containing a URL with `/opinion`.
2.  Bot checks authorization (Chat ID) and rate limits (Redis).
3.  Bot extracts the URL from the original message.
4.  If a URL is found:
    *   Checks Redis cache for existing analysis of this specific message.
    *   If not cached, calls Gemini API with a randomized prompt.
    *   Caches the result in Redis (30-day TTL) and replies to the user.
5.  If no URL is found:
    *   Returns a canned refusal response.

## Building and Running

### Prerequisites
*   Go 1.24+
*   Docker & Docker Compose (optional but recommended)
*   Valid `.env` file (see below)

### Local Development
```bash
# Install dependencies
go mod download

# Run the bot
go run .

# Run tests
go test -v ./...
```

### Docker
```bash
# Build and run with Compose (includes Valkey/Redis)
docker-compose up -d --build
```

## Configuration (`.env`)

The application relies on environment variables:

| Variable | Description | Required |
| :--- | :--- | :--- |
| `TELEGRAM_BOT_TOKEN` | Telegram Bot API Token | Yes |
| `GOOGLE_API_KEY` | Google Gemini API Key | Yes |
| `ALLOWED_CHAT_IDS` | Comma-separated list of authorized chat IDs | Yes |
| `GROUP_LINK` | Link to the main group (displayed in error messages) | No |
| `EXCLUDED_USER_IDS` | Comma-separated list of User IDs to bypass rate limits | No |
| `REDIS_ADDR` | Redis address (default: `localhost:6379`) | No |

## Testing

*   **Unit Tests:** exist for `main`, `opinion`, and `llm`.
*   **Mocks:** Tests use a `MockContext` to simulate Telegram interactions.
*   **Coverage:** Run `go test -cover ./...` to check coverage.

## Conventions

*   **Logging:** Use `logJSON` for all logs to ensure structured output.
*   **Error Handling:** Errors are logged with context but often fail gracefully (e.g., falling back to a generic message if LLM fails).
*   **Code Style:** Standard Go idioms. 
*   **Architecture:** Separation of concerns between Bot logic (`main`), Business logic (`opinion`), and AI integration (`llm`).
