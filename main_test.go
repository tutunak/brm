package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	tele "gopkg.in/telebot.v3"
)

func TestParseAllowedChatIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int64
	}{
		{
			name:     "Single chat ID",
			input:    "-1001234567890",
			expected: []int64{-1001234567890},
		},
		{
			name:     "Multiple chat IDs",
			input:    "-1001234567890,-1009876543210,123456789",
			expected: []int64{-1001234567890, -1009876543210, 123456789},
		},
		{
			name:     "With spaces",
			input:    "-1001234567890, -1009876543210 , 123456789",
			expected: []int64{-1001234567890, -1009876543210, 123456789},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "Only spaces and commas",
			input:    " , , ",
			expected: nil,
		},
		{
			name:     "Invalid chat ID mixed with valid",
			input:    "-1001234567890,invalid,123456789",
			expected: []int64{-1001234567890, 123456789},
		},
		{
			name:     "Positive and negative IDs",
			input:    "123456789,-987654321",
			expected: []int64{123456789, -987654321},
		},
		{
			name:     "Trailing comma",
			input:    "-1001234567890,",
			expected: []int64{-1001234567890},
		},
		{
			name:     "Leading comma",
			input:    ",-1001234567890",
			expected: []int64{-1001234567890},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAllowedChatIDs(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("parseAllowedChatIDs(%q) returned %d elements, want %d", tt.input, len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseAllowedChatIDs(%q)[%d] = %d, want %d", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestParseExcludedUserIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int64
	}{
		{
			name:     "Single user ID",
			input:    "123456789",
			expected: []int64{123456789},
		},
		{
			name:     "Multiple user IDs",
			input:    "123456789,987654321,555555555",
			expected: []int64{123456789, 987654321, 555555555},
		},
		{
			name:     "With spaces",
			input:    "123456789, 987654321 , 555555555",
			expected: []int64{123456789, 987654321, 555555555},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []int64{},
		},
		{
			name:     "Invalid user ID mixed with valid",
			input:    "123456789,notanumber,987654321",
			expected: []int64{123456789, 987654321},
		},
		{
			name:     "Trailing comma",
			input:    "123456789,",
			expected: []int64{123456789},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseExcludedUserIDs(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("parseExcludedUserIDs(%q) returned %d elements, want %d", tt.input, len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseExcludedUserIDs(%q)[%d] = %d, want %d", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestIsExcludedUser(t *testing.T) {
	tests := []struct {
		name       string
		envValue   string
		userID     int64
		expected   bool
	}{
		{
			name:     "User is excluded",
			envValue: "123456789,987654321",
			userID:   123456789,
			expected: true,
		},
		{
			name:     "User is not excluded",
			envValue: "123456789,987654321",
			userID:   555555555,
			expected: false,
		},
		{
			name:     "Empty exclusion list",
			envValue: "",
			userID:   123456789,
			expected: false,
		},
		{
			name:     "Single excluded user - match",
			envValue: "123456789",
			userID:   123456789,
			expected: true,
		},
		{
			name:     "Single excluded user - no match",
			envValue: "123456789",
			userID:   987654321,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("EXCLUDED_USER_IDS", tt.envValue)
			defer os.Unsetenv("EXCLUDED_USER_IDS")

			result := isExcludedUser(tt.userID)

			if result != tt.expected {
				t.Errorf("isExcludedUser(%d) with env=%q = %v, want %v", tt.userID, tt.envValue, result, tt.expected)
			}
		})
	}
}

func TestLogJSON(t *testing.T) {
	tests := []struct {
		name           string
		level          string
		message        string
		data           map[string]interface{}
		expectedFields []string
	}{
		{
			name:           "Basic log without data",
			level:          "info",
			message:        "Test message",
			data:           nil,
			expectedFields: []string{"timestamp", "level", "message"},
		},
		{
			name:    "Log with additional data",
			level:   "error",
			message: "Error occurred",
			data: map[string]interface{}{
				"error_code": 500,
				"details":    "Something went wrong",
			},
			expectedFields: []string{"timestamp", "level", "message", "error_code", "details"},
		},
		{
			name:           "Log with empty data map",
			level:          "debug",
			message:        "Debug info",
			data:           map[string]interface{}{},
			expectedFields: []string{"timestamp", "level", "message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			logJSON(tt.level, tt.message, tt.data)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Parse the JSON output
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Errorf("Failed to parse log output as JSON: %v", err)
				return
			}

			// Check expected fields exist
			for _, field := range tt.expectedFields {
				if _, exists := logEntry[field]; !exists {
					t.Errorf("Expected field %q not found in log output", field)
				}
			}

			// Verify level and message values
			if logEntry["level"] != tt.level {
				t.Errorf("Log level = %v, want %v", logEntry["level"], tt.level)
			}
			if logEntry["message"] != tt.message {
				t.Errorf("Log message = %v, want %v", logEntry["message"], tt.message)
			}
		})
	}
}

func TestLogJSONDataValues(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]interface{}{
		"user_id":   int64(123456789),
		"chat_name": "Test Chat",
		"is_admin":  true,
	}
	logJSON("info", "Test with data", data)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Failed to parse log output as JSON: %v", err)
		return
	}

	// Check data values are correctly included
	if logEntry["chat_name"] != "Test Chat" {
		t.Errorf("chat_name = %v, want %v", logEntry["chat_name"], "Test Chat")
	}
	if logEntry["is_admin"] != true {
		t.Errorf("is_admin = %v, want %v", logEntry["is_admin"], true)
	}
}

// MockContext implements a minimal mock for tele.Context
type MockContext struct {
	chat    *tele.Chat
	sender  *tele.User
	message *tele.Message
}

func (m *MockContext) Chat() *tele.Chat {
	return m.chat
}

func (m *MockContext) Sender() *tele.User {
	return m.sender
}

func (m *MockContext) Message() *tele.Message {
	return m.message
}

// Implement other required interface methods with stubs
func (m *MockContext) Bot() *tele.Bot                                       { return nil }
func (m *MockContext) Update() tele.Update                                  { return tele.Update{} }
func (m *MockContext) Callback() *tele.Callback                             { return nil }
func (m *MockContext) Query() *tele.Query                                   { return nil }
func (m *MockContext) InlineResult() *tele.InlineResult                     { return nil }
func (m *MockContext) ShippingQuery() *tele.ShippingQuery                   { return nil }
func (m *MockContext) PreCheckoutQuery() *tele.PreCheckoutQuery             { return nil }
func (m *MockContext) Poll() *tele.Poll                                     { return nil }
func (m *MockContext) PollAnswer() *tele.PollAnswer                         { return nil }
func (m *MockContext) ChatMember() *tele.ChatMemberUpdate                   { return nil }
func (m *MockContext) ChatJoinRequest() *tele.ChatJoinRequest               { return nil }
func (m *MockContext) Migration() (int64, int64)                            { return 0, 0 }
func (m *MockContext) Topic() *tele.Topic                                   { return nil }
func (m *MockContext) Recipient() tele.Recipient                            { return nil }
func (m *MockContext) Text() string                                         { return "" }
func (m *MockContext) Entities() tele.Entities                              { return nil }
func (m *MockContext) Data() string                                         { return "" }
func (m *MockContext) Args() []string                                       { return nil }
func (m *MockContext) Send(what interface{}, opts ...interface{}) error     { return nil }
func (m *MockContext) SendAlbum(a tele.Album, opts ...interface{}) error    { return nil }
func (m *MockContext) Reply(what interface{}, opts ...interface{}) error    { return nil }
func (m *MockContext) Forward(msg tele.Editable, opts ...interface{}) error { return nil }
func (m *MockContext) ForwardTo(to tele.Recipient, opts ...interface{}) error { return nil }
func (m *MockContext) Edit(what interface{}, opts ...interface{}) error     { return nil }
func (m *MockContext) EditCaption(caption string, opts ...interface{}) error { return nil }
func (m *MockContext) EditOrSend(what interface{}, opts ...interface{}) error { return nil }
func (m *MockContext) EditOrReply(what interface{}, opts ...interface{}) error { return nil }
func (m *MockContext) Delete() error                                        { return nil }
func (m *MockContext) DeleteAfter(d time.Duration) *time.Timer              { return nil }
func (m *MockContext) Notify(action tele.ChatAction) error                  { return nil }
func (m *MockContext) Ship(what ...interface{}) error                       { return nil }
func (m *MockContext) Accept(errorMessage ...string) error                  { return nil }
func (m *MockContext) Answer(resp *tele.QueryResponse) error                { return nil }
func (m *MockContext) Respond(resp ...*tele.CallbackResponse) error         { return nil }
func (m *MockContext) Get(key string) interface{}                           { return nil }
func (m *MockContext) Set(key string, val interface{})                      {}
func (m *MockContext) ThreadID() int                                        { return 0 }

func TestIsAllowedChat(t *testing.T) {
	tests := []struct {
		name       string
		envValue   string
		chatID     int64
		expected   bool
	}{
		{
			name:     "Chat is allowed",
			envValue: "-1001234567890,-1009876543210",
			chatID:   -1001234567890,
			expected: true,
		},
		{
			name:     "Chat is not allowed",
			envValue: "-1001234567890,-1009876543210",
			chatID:   -1005555555555,
			expected: false,
		},
		{
			name:     "Empty allowed list",
			envValue: "",
			chatID:   -1001234567890,
			expected: false,
		},
		{
			name:     "Single allowed chat - match",
			envValue: "-1001234567890",
			chatID:   -1001234567890,
			expected: true,
		},
		{
			name:     "Single allowed chat - no match",
			envValue: "-1001234567890",
			chatID:   -1009876543210,
			expected: false,
		},
		{
			name:     "Positive chat ID allowed",
			envValue: "123456789",
			chatID:   123456789,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ALLOWED_CHAT_IDS", tt.envValue)
			defer os.Unsetenv("ALLOWED_CHAT_IDS")

			mockCtx := &MockContext{
				chat: &tele.Chat{ID: tt.chatID},
			}

			result := isAllowedChat(mockCtx)

			if result != tt.expected {
				t.Errorf("isAllowedChat() with chatID=%d and env=%q = %v, want %v",
					tt.chatID, tt.envValue, result, tt.expected)
			}
		})
	}
}

func TestGetUserInfo(t *testing.T) {
	tests := []struct {
		name     string
		user     *tele.User
		expected map[string]interface{}
	}{
		{
			name: "User with all fields",
			user: &tele.User{
				ID:        123456789,
				Username:  "testuser",
				FirstName: "Test",
			},
			expected: map[string]interface{}{
				"username":   "testuser",
				"first_name": "Test",
				"user_id":    int64(123456789),
			},
		},
		{
			name: "User without username",
			user: &tele.User{
				ID:        123456789,
				Username:  "",
				FirstName: "Test",
			},
			expected: map[string]interface{}{
				"username":   "no_username",
				"first_name": "Test",
				"user_id":    int64(123456789),
			},
		},
		{
			name: "User without first name",
			user: &tele.User{
				ID:       123456789,
				Username: "testuser",
			},
			expected: map[string]interface{}{
				"username":   "testuser",
				"first_name": "Unknown",
				"user_id":    int64(123456789),
			},
		},
		{
			name: "Nil user",
			user: nil,
			expected: map[string]interface{}{
				"username":   "unknown",
				"first_name": "Unknown",
				"user_id":    0, // untyped int, matches what getUserInfo returns for nil user
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := &MockContext{
				sender: tt.user,
			}

			result := getUserInfo(mockCtx)

			if result["username"] != tt.expected["username"] {
				t.Errorf("getUserInfo().username = %v, want %v", result["username"], tt.expected["username"])
			}
			if result["first_name"] != tt.expected["first_name"] {
				t.Errorf("getUserInfo().first_name = %v, want %v", result["first_name"], tt.expected["first_name"])
			}
			if result["user_id"] != tt.expected["user_id"] {
				t.Errorf("getUserInfo().user_id = %v, want %v", result["user_id"], tt.expected["user_id"])
			}
		})
	}
}

func TestGetChatInfo(t *testing.T) {
	tests := []struct {
		name     string
		chat     *tele.Chat
		expected map[string]interface{}
	}{
		{
			name: "Group chat with title",
			chat: &tele.Chat{
				ID:    -1001234567890,
				Type:  tele.ChatGroup,
				Title: "Test Group",
			},
			expected: map[string]interface{}{
				"chat_title": "Test Group",
				"chat_type":  "group",
				"chat_id":    int64(-1001234567890),
			},
		},
		{
			name: "Private chat with username",
			chat: &tele.Chat{
				ID:       123456789,
				Type:     tele.ChatPrivate,
				Username: "testuser",
			},
			expected: map[string]interface{}{
				"chat_title": "@testuser",
				"chat_type":  "private",
				"chat_id":    int64(123456789),
			},
		},
		{
			name: "Private chat with first name only",
			chat: &tele.Chat{
				ID:        123456789,
				Type:      tele.ChatPrivate,
				FirstName: "John",
			},
			expected: map[string]interface{}{
				"chat_title": "John",
				"chat_type":  "private",
				"chat_id":    int64(123456789),
			},
		},
		{
			name: "Private chat without username or first name",
			chat: &tele.Chat{
				ID:   123456789,
				Type: tele.ChatPrivate,
			},
			expected: map[string]interface{}{
				"chat_title": "Private Chat",
				"chat_type":  "private",
				"chat_id":    int64(123456789),
			},
		},
		{
			name: "Nil chat",
			chat: nil,
			expected: map[string]interface{}{
				"chat_title": "Unknown",
				"chat_type":  "unknown",
				"chat_id":    0, // untyped int, matches what getChatInfo returns for nil chat
			},
		},
		{
			name: "Supergroup chat",
			chat: &tele.Chat{
				ID:    -1001234567890,
				Type:  tele.ChatSuperGroup,
				Title: "Super Group",
			},
			expected: map[string]interface{}{
				"chat_title": "Super Group",
				"chat_type":  "supergroup",
				"chat_id":    int64(-1001234567890),
			},
		},
		{
			name: "Channel",
			chat: &tele.Chat{
				ID:    -1001234567890,
				Type:  tele.ChatChannel,
				Title: "Test Channel",
			},
			expected: map[string]interface{}{
				"chat_title": "Test Channel",
				"chat_type":  "channel",
				"chat_id":    int64(-1001234567890),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := &MockContext{
				chat: tt.chat,
			}

			result := getChatInfo(mockCtx)

			if result["chat_title"] != tt.expected["chat_title"] {
				t.Errorf("getChatInfo().chat_title = %v, want %v", result["chat_title"], tt.expected["chat_title"])
			}
			if result["chat_type"] != tt.expected["chat_type"] {
				t.Errorf("getChatInfo().chat_type = %v, want %v", result["chat_type"], tt.expected["chat_type"])
			}
			if result["chat_id"] != tt.expected["chat_id"] {
				t.Errorf("getChatInfo().chat_id = %v, want %v", result["chat_id"], tt.expected["chat_id"])
			}
		})
	}
}

func TestLogRequest(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	mockCtx := &MockContext{
		chat: &tele.Chat{
			ID:    -1001234567890,
			Type:  tele.ChatGroup,
			Title: "Test Group",
		},
		sender: &tele.User{
			ID:        123456789,
			Username:  "testuser",
			FirstName: "Test",
		},
		message: &tele.Message{
			ID: 42,
		},
	}

	logRequest(mockCtx, "/opinion")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Failed to parse log output as JSON: %v", err)
		return
	}

	// Verify key fields
	if logEntry["level"] != "request" {
		t.Errorf("logRequest level = %v, want 'request'", logEntry["level"])
	}
	if logEntry["command"] != "/opinion" {
		t.Errorf("logRequest command = %v, want '/opinion'", logEntry["command"])
	}
	if logEntry["message_id"] != float64(42) { // JSON numbers are float64
		t.Errorf("logRequest message_id = %v, want 42", logEntry["message_id"])
	}
}
