# Telegram Opinion Bot

A Telegram bot written in Go that analyzes messages and provides sentiment opinions using the `/opinion` command.

## Features

- 🤖 Simple Telegram bot with command handling
- 💬 `/opinion` command for sentiment analysis
- 📊 Analyzes text for positive/negative sentiment
- ✅ Works only on reply messages
- 🐳 Docker support included

## Prerequisites

- Go 1.21 or higher
- A Telegram Bot Token (get it from [@BotFather](https://t.me/botfather))
- Docker (optional, for containerized deployment)

## Getting Your Bot Token

1. Open Telegram and search for [@BotFather](https://t.me/botfather)
2. Send `/newbot` command
3. Follow the instructions to create your bot
4. Copy the API token provided by BotFather

## Installation & Setup

### Method 1: Running Locally

1. **Clone the repository**
   ```bash
   cd /home/dk/projects/my/golang/brm
   ```

2. **Create a `.env` file**
   ```bash
   cp .env.example .env
   ```

3. **Edit `.env` and add your bot token**
   ```
   TELEGRAM_BOT_TOKEN=your_bot_token_here
   ```

4. **Install dependencies**
   ```bash
   go mod download
   ```

5. **Run the bot**
   ```bash
   go run .
   ```

### Method 2: Using Docker

1. **Create a `.env` file with your bot token**
   ```
   TELEGRAM_BOT_TOKEN=your_bot_token_here
   ```

2. **Build the Docker image**
   ```bash
   docker build -t telegram-opinion-bot .
   ```

3. **Run the container**
   ```bash
   docker run -d --name opinion-bot --env-file .env telegram-opinion-bot
   ```

4. **View logs**
   ```bash
   docker logs -f opinion-bot
   ```

5. **Stop the container**
   ```bash
   docker stop opinion-bot
   ```

## Running Tests

Run the unit tests with:

```bash
go test -v
```

For coverage report:

```bash
go test -cover
```

## Usage

### In Private Chat

1. Start a conversation with your bot
2. Send any message
3. Reply to that message with `/opinion`
4. The bot will analyze the original message and reply with sentiment analysis

### Adding Bot to a Channel

1. **Make your bot an admin** (required for it to read messages):
   - Open your channel
   - Go to channel settings → Administrators
   - Click "Add Administrator"
   - Search for your bot and add it
   - Grant necessary permissions (at minimum: "Post Messages" and "Delete Messages")

2. **Disable Privacy Mode** (to allow bot to see all messages):
   - Go to [@BotFather](https://t.me/botfather)
   - Send `/mybots`
   - Select your bot
   - Go to "Bot Settings" → "Group Privacy"
   - Click "Turn off"

3. **Use the bot in the channel**:
   - Reply to any message in the channel with `/opinion`
   - The bot will analyze and respond

### Adding Bot to a Group

1. Add the bot to your group
2. Make sure the bot has permission to read messages (Privacy Mode should be disabled)
3. Reply to any message with `/opinion`
4. The bot will analyze and respond

## Commands

- `/opinion` - Analyze sentiment of the replied message (must be used as a reply)

## How It Works

The `/opinion` command:
1. Must be used as a reply to another message
2. Extracts the text from the replied message
3. Performs simple sentiment analysis by counting positive and negative keywords
4. Returns an analysis including:
   - Overall sentiment (positive 😊, negative 😔, or neutral 😐)
   - Word count
   - Number of positive indicators
   - Number of negative indicators

## Project Structure

```
.
├── main.go           # Main bot logic and command handling
├── opinion.go        # Opinion/sentiment analysis function
├── opinion_test.go   # Unit tests
├── go.mod           # Go module definition
├── Dockerfile       # Docker configuration
├── .env.example     # Example environment variables
└── README.md        # This file
```

## Environment Variables

- `TELEGRAM_BOT_TOKEN` - Your Telegram bot API token (required)

## Development

To add new commands:

1. Add a new case in the `handleCommand` function in [main.go](main.go)
2. Implement the command handler function
3. Add unit tests for the new functionality

## Troubleshooting

**Bot doesn't respond in groups:**
- Ensure Privacy Mode is disabled in BotFather settings
- Make sure the bot has necessary permissions in the group/channel

**"TELEGRAM_BOT_TOKEN environment variable is required" error:**
- Check that your `.env` file exists and contains the token
- Ensure the environment variable is properly set

**Bot not seeing messages in channel:**
- The bot must be an administrator of the channel
- Privacy Mode must be disabled

## License

This project is open source and available under the MIT License.

## Contributing

Feel free to submit issues and enhancement requests!
