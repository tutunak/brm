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

// TestParseAllowedChatIDsEdgeCases tests additional edge cases for parseAllowedChatIDs
func TestParseAllowedChatIDsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int64
	}{
		{
			name:     "Very large positive ID",
			input:    "9223372036854775807",
			expected: []int64{9223372036854775807},
		},
		{
			name:     "Very large negative ID",
			input:    "-9223372036854775808",
			expected: []int64{-9223372036854775808},
		},
		{
			name:     "Multiple commas",
			input:    "123,,456,,,789",
			expected: []int64{123, 456, 789},
		},
		{
			name:     "Tabs and newlines",
			input:    "123\t,\n456",
			expected: []int64{123, 456},
		},
		{
			name:     "Zero ID",
			input:    "0",
			expected: []int64{0},
		},
		{
			name:     "Only invalid values",
			input:    "abc,def,ghi",
			expected: nil,
		},
		{
			name:     "Float values (invalid)",
			input:    "123.456,789",
			expected: []int64{789},
		},
		{
			name:     "Hex values (invalid)",
			input:    "0x123,456",
			expected: []int64{456},
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

// TestParseExcludedUserIDsEdgeCases tests additional edge cases for parseExcludedUserIDs
func TestParseExcludedUserIDsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int64
	}{
		{
			name:     "Very large ID",
			input:    "9223372036854775807",
			expected: []int64{9223372036854775807},
		},
		{
			name:     "Multiple commas between values",
			input:    "123,,456,,,789",
			expected: []int64{123, 456, 789},
		},
		{
			name:     "Whitespace variations",
			input:    "  123  ,  456  ,  789  ",
			expected: []int64{123, 456, 789},
		},
		{
			name:     "Zero ID",
			input:    "0",
			expected: []int64{0},
		},
		{
			name:     "Negative ID (unusual but valid)",
			input:    "-123",
			expected: []int64{-123},
		},
		{
			name:     "Only spaces",
			input:    "   ",
			expected: []int64{},
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

// TestIsExcludedUserEdgeCases tests additional edge cases for isExcludedUser
func TestIsExcludedUserEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		userID   int64
		expected bool
	}{
		{
			name:     "Zero user ID in list - match",
			envValue: "0,123456789",
			userID:   0,
			expected: true,
		},
		{
			name:     "Very large user ID - match",
			envValue: "9223372036854775807",
			userID:   9223372036854775807,
			expected: true,
		},
		{
			name:     "Multiple IDs with spaces - match",
			envValue: "  123  ,  456  ,  789  ",
			userID:   456,
			expected: true,
		},
		{
			name:     "Invalid entries mixed - match valid",
			envValue: "abc,123,def",
			userID:   123,
			expected: true,
		},
		{
			name:     "Whitespace only env",
			envValue: "   ",
			userID:   123,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("EXCLUDED_USER_IDS", tt.envValue)
			defer os.Unsetenv("EXCLUDED_USER_IDS")

			result := isExcludedUser(tt.userID)

			if result != tt.expected {
				t.Errorf("isExcludedUser(%d) with env=%q = %v, want %v", tt.userID, tt.envValue, result, tt.expected)
			}
		})
	}
}

// TestIsAllowedChatEdgeCases tests additional edge cases for isAllowedChat
func TestIsAllowedChatEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		chatID   int64
		expected bool
	}{
		{
			name:     "Zero chat ID - match",
			envValue: "0",
			chatID:   0,
			expected: true,
		},
		{
			name:     "Very large negative chat ID - match",
			envValue: "-9223372036854775808",
			chatID:   -9223372036854775808,
			expected: true,
		},
		{
			name:     "Mixed valid and invalid entries - match valid",
			envValue: "abc,-1001234567890,def",
			chatID:   -1001234567890,
			expected: true,
		},
		{
			name:     "Whitespace only env",
			envValue: "   ",
			chatID:   123,
			expected: false,
		},
		{
			name:     "Multiple same IDs",
			envValue: "123,123,123",
			chatID:   123,
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

// TestGetUserInfoEdgeCases tests additional edge cases for getUserInfo
func TestGetUserInfoEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		user     *tele.User
		expected map[string]interface{}
	}{
		{
			name: "User with only ID",
			user: &tele.User{
				ID: 123456789,
			},
			expected: map[string]interface{}{
				"username":   "no_username",
				"first_name": "Unknown",
				"user_id":    int64(123456789),
			},
		},
		{
			name: "User with very long username",
			user: &tele.User{
				ID:        123456789,
				Username:  "very_long_username_that_goes_on_and_on",
				FirstName: "Test",
			},
			expected: map[string]interface{}{
				"username":   "very_long_username_that_goes_on_and_on",
				"first_name": "Test",
				"user_id":    int64(123456789),
			},
		},
		{
			name: "User with special characters in first name",
			user: &tele.User{
				ID:        123456789,
				Username:  "user",
				FirstName: "Test 🎉 User",
			},
			expected: map[string]interface{}{
				"username":   "user",
				"first_name": "Test 🎉 User",
				"user_id":    int64(123456789),
			},
		},
		{
			name: "User with zero ID",
			user: &tele.User{
				ID:        0,
				Username:  "user",
				FirstName: "Test",
			},
			expected: map[string]interface{}{
				"username":   "user",
				"first_name": "Test",
				"user_id":    int64(0),
			},
		},
		{
			name: "User with LastName (ignored)",
			user: &tele.User{
				ID:        123456789,
				Username:  "user",
				FirstName: "First",
				LastName:  "Last",
			},
			expected: map[string]interface{}{
				"username":   "user",
				"first_name": "First",
				"user_id":    int64(123456789),
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

// TestGetChatInfoEdgeCases tests additional edge cases for getChatInfo
func TestGetChatInfoEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		chat     *tele.Chat
		expected map[string]interface{}
	}{
		{
			name: "Chat with both username and first name (username takes priority)",
			chat: &tele.Chat{
				ID:        123456789,
				Type:      tele.ChatPrivate,
				Username:  "testuser",
				FirstName: "John",
			},
			expected: map[string]interface{}{
				"chat_title": "@testuser",
				"chat_type":  "private",
				"chat_id":    int64(123456789),
			},
		},
		{
			name: "Chat with Title takes precedence over username",
			chat: &tele.Chat{
				ID:       123456789,
				Type:     tele.ChatPrivate,
				Title:    "Chat Title",
				Username: "testuser",
			},
			expected: map[string]interface{}{
				"chat_title": "Chat Title",
				"chat_type":  "private",
				"chat_id":    int64(123456789),
			},
		},
		{
			name: "Chat with very long title",
			chat: &tele.Chat{
				ID:    -1001234567890,
				Type:  tele.ChatGroup,
				Title: "This is a very long group title that contains many words and characters",
			},
			expected: map[string]interface{}{
				"chat_title": "This is a very long group title that contains many words and characters",
				"chat_type":  "group",
				"chat_id":    int64(-1001234567890),
			},
		},
		{
			name: "Chat with zero ID",
			chat: &tele.Chat{
				ID:    0,
				Type:  tele.ChatPrivate,
				Title: "Zero Chat",
			},
			expected: map[string]interface{}{
				"chat_title": "Zero Chat",
				"chat_type":  "private",
				"chat_id":    int64(0),
			},
		},
		{
			name: "Chat with emoji in title",
			chat: &tele.Chat{
				ID:    -1001234567890,
				Type:  tele.ChatGroup,
				Title: "🎉 Fun Group 🎊",
			},
			expected: map[string]interface{}{
				"chat_title": "🎉 Fun Group 🎊",
				"chat_type":  "group",
				"chat_id":    int64(-1001234567890),
			},
		},
		{
			name: "Chat private channel type",
			chat: &tele.Chat{
				ID:    -1001234567890,
				Type:  tele.ChatChannelPrivate,
				Title: "Private Channel",
			},
			expected: map[string]interface{}{
				"chat_title": "Private Channel",
				"chat_type":  "privatechannel",
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

// TestLogJSONWithNilData tests logJSON with nil data
func TestLogJSONWithNilData(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logJSON("info", "Test message", nil)

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

	// Should have exactly timestamp, level, message
	expectedKeys := map[string]bool{"timestamp": true, "level": true, "message": true}
	for key := range logEntry {
		if !expectedKeys[key] {
			t.Errorf("Unexpected key %q in log output with nil data", key)
		}
	}
}

// TestLogJSONWithNestedData tests logJSON with nested map data
func TestLogJSONWithNestedData(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]interface{}{
		"nested": map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
		"array": []int{1, 2, 3},
	}
	logJSON("info", "Nested test", data)

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

	// Check nested map exists
	if _, ok := logEntry["nested"]; !ok {
		t.Error("Expected 'nested' key in log output")
	}

	// Check array exists
	if _, ok := logEntry["array"]; !ok {
		t.Error("Expected 'array' key in log output")
	}
}

// TestLogJSONTimestampFormat tests that the timestamp is in RFC3339 format
func TestLogJSONTimestampFormat(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logJSON("info", "Timestamp test", nil)

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

	timestamp, ok := logEntry["timestamp"].(string)
	if !ok {
		t.Error("Timestamp is not a string")
		return
	}

	// Try to parse as RFC3339
	_, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		t.Errorf("Timestamp %q is not in RFC3339 format: %v", timestamp, err)
	}
}

// TestLogJSONVariousLevels tests logJSON with different log levels
func TestLogJSONVariousLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "success", "request", "fatal"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			logJSON(level, "Test message", nil)

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

			if logEntry["level"] != level {
				t.Errorf("Log level = %v, want %v", logEntry["level"], level)
			}
		})
	}
}

// TestLogRequestWithDifferentCommands tests logRequest with various commands
func TestLogRequestWithDifferentCommands(t *testing.T) {
	commands := []string{"/opinion", "/start", "/help", "/settings", "custom_command"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
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

			logRequest(mockCtx, cmd)

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

			if logEntry["command"] != cmd {
				t.Errorf("logRequest command = %v, want %v", logEntry["command"], cmd)
			}
		})
	}
}

// TestLogRequestIncludesUserAndChatInfo tests that logRequest includes user and chat info
func TestLogRequestIncludesUserAndChatInfo(t *testing.T) {
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

	// Check user info is present
	user, ok := logEntry["user"].(map[string]interface{})
	if !ok {
		t.Error("Expected 'user' to be a map in log output")
		return
	}
	if user["username"] != "testuser" {
		t.Errorf("user.username = %v, want 'testuser'", user["username"])
	}

	// Check chat info is present
	chat, ok := logEntry["chat"].(map[string]interface{})
	if !ok {
		t.Error("Expected 'chat' to be a map in log output")
		return
	}
	if chat["chat_title"] != "Test Group" {
		t.Errorf("chat.chat_title = %v, want 'Test Group'", chat["chat_title"])
	}
}

// TestMockContextMethods tests that MockContext methods return expected values
func TestMockContextMethods(t *testing.T) {
	mockCtx := &MockContext{
		chat: &tele.Chat{
			ID:    123456789,
			Type:  tele.ChatPrivate,
			Title: "Test Chat",
		},
		sender: &tele.User{
			ID:        987654321,
			Username:  "testuser",
			FirstName: "Test",
		},
		message: &tele.Message{
			ID: 42,
		},
	}

	// Test Chat()
	if mockCtx.Chat().ID != 123456789 {
		t.Errorf("MockContext.Chat().ID = %d, want 123456789", mockCtx.Chat().ID)
	}

	// Test Sender()
	if mockCtx.Sender().ID != 987654321 {
		t.Errorf("MockContext.Sender().ID = %d, want 987654321", mockCtx.Sender().ID)
	}

	// Test Message()
	if mockCtx.Message().ID != 42 {
		t.Errorf("MockContext.Message().ID = %d, want 42", mockCtx.Message().ID)
	}

	// Test stub methods return nil/zero values
	if mockCtx.Bot() != nil {
		t.Error("MockContext.Bot() should return nil")
	}
	if mockCtx.Callback() != nil {
		t.Error("MockContext.Callback() should return nil")
	}
	if mockCtx.Query() != nil {
		t.Error("MockContext.Query() should return nil")
	}
	if mockCtx.Text() != "" {
		t.Error("MockContext.Text() should return empty string")
	}
	if mockCtx.Data() != "" {
		t.Error("MockContext.Data() should return empty string")
	}
	if mockCtx.Args() != nil {
		t.Error("MockContext.Args() should return nil")
	}
	if mockCtx.ThreadID() != 0 {
		t.Error("MockContext.ThreadID() should return 0")
	}
}

// TestHandleOpinionCommandUnauthorizedChat tests handleOpinionCommand with unauthorized chat
func TestHandleOpinionCommandUnauthorizedChat(t *testing.T) {
	// Save and restore environment
	originalAllowedChats := os.Getenv("ALLOWED_CHAT_IDS")
	originalGroupLink := os.Getenv("GROUP_LINK")
	defer func() {
		os.Setenv("ALLOWED_CHAT_IDS", originalAllowedChats)
		os.Setenv("GROUP_LINK", originalGroupLink)
	}()

	os.Setenv("ALLOWED_CHAT_IDS", "-1001111111111")
	os.Setenv("GROUP_LINK", "https://t.me/testgroup")

	tests := []struct {
		name     string
		chatType tele.ChatType
		chatID   int64
	}{
		{
			name:     "Unauthorized private chat",
			chatType: tele.ChatPrivate,
			chatID:   123456789,
		},
		{
			name:     "Unauthorized group chat",
			chatType: tele.ChatGroup,
			chatID:   -1002222222222,
		},
		{
			name:     "Unauthorized supergroup chat",
			chatType: tele.ChatSuperGroup,
			chatID:   -1003333333333,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout to suppress logs
			oldStdout := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			replyCalled := false
			replyMessage := ""

			mockCtx := &MockContextWithReply{
				MockContext: MockContext{
					chat: &tele.Chat{
						ID:   tt.chatID,
						Type: tt.chatType,
					},
					sender: &tele.User{
						ID:        123456789,
						Username:  "testuser",
						FirstName: "Test",
					},
					message: &tele.Message{
						ID: 42,
					},
				},
				replyFunc: func(what interface{}, opts ...interface{}) error {
					replyCalled = true
					replyMessage = what.(string)
					return nil
				},
			}

			err := handleOpinionCommand(mockCtx)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Errorf("handleOpinionCommand returned error: %v", err)
			}

			if !replyCalled {
				t.Error("Reply was not called for unauthorized chat")
			}

			if replyMessage == "" {
				t.Error("Reply message was empty")
			}
		})
	}
}

// MockContextWithReply extends MockContext with a custom Reply function
type MockContextWithReply struct {
	MockContext
	replyFunc func(what interface{}, opts ...interface{}) error
}

func (m *MockContextWithReply) Reply(what interface{}, opts ...interface{}) error {
	if m.replyFunc != nil {
		return m.replyFunc(what, opts...)
	}
	return nil
}

// TestHandleOpinionCommandNoReply tests handleOpinionCommand when message has no reply
func TestHandleOpinionCommandNoReply(t *testing.T) {
	// Save and restore environment
	originalAllowedChats := os.Getenv("ALLOWED_CHAT_IDS")
	defer func() {
		os.Setenv("ALLOWED_CHAT_IDS", originalAllowedChats)
	}()

	os.Setenv("ALLOWED_CHAT_IDS", "-1001234567890")

	// Capture stdout to suppress logs
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	replyCalled := false
	replyMessage := ""

	mockCtx := &MockContextWithReply{
		MockContext: MockContext{
			chat: &tele.Chat{
				ID:   -1001234567890,
				Type: tele.ChatGroup,
			},
			sender: &tele.User{
				ID:        123456789,
				Username:  "testuser",
				FirstName: "Test",
			},
			message: &tele.Message{
				ID:      42,
				ReplyTo: nil, // No reply
			},
		},
		replyFunc: func(what interface{}, opts ...interface{}) error {
			replyCalled = true
			replyMessage = what.(string)
			return nil
		},
	}

	err := handleOpinionCommand(mockCtx)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("handleOpinionCommand returned error: %v", err)
	}

	if !replyCalled {
		t.Error("Reply was not called when message has no reply")
	}

	expectedMessage := "Please use /opinion as a reply to a message"
	if replyMessage != expectedMessage {
		t.Errorf("Reply message = %q, want %q", replyMessage, expectedMessage)
	}
}

// TestHandleOpinionCommandEmptyReplyText tests handleOpinionCommand when replied message has no text
func TestHandleOpinionCommandEmptyReplyText(t *testing.T) {
	// Save and restore environment
	originalAllowedChats := os.Getenv("ALLOWED_CHAT_IDS")
	defer func() {
		os.Setenv("ALLOWED_CHAT_IDS", originalAllowedChats)
	}()

	os.Setenv("ALLOWED_CHAT_IDS", "-1001234567890")

	// Capture stdout to suppress logs
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	replyCalled := false
	replyMessage := ""

	mockCtx := &MockContextWithReply{
		MockContext: MockContext{
			chat: &tele.Chat{
				ID:   -1001234567890,
				Type: tele.ChatGroup,
			},
			sender: &tele.User{
				ID:        123456789,
				Username:  "testuser",
				FirstName: "Test",
			},
			message: &tele.Message{
				ID: 42,
				ReplyTo: &tele.Message{
					ID:   41,
					Text: "", // Empty text
				},
			},
		},
		replyFunc: func(what interface{}, opts ...interface{}) error {
			replyCalled = true
			replyMessage = what.(string)
			return nil
		},
	}

	err := handleOpinionCommand(mockCtx)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("handleOpinionCommand returned error: %v", err)
	}

	if !replyCalled {
		t.Error("Reply was not called when replied message has no text")
	}

	expectedMessage := "The replied message has no text to analyze"
	if replyMessage != expectedMessage {
		t.Errorf("Reply message = %q, want %q", replyMessage, expectedMessage)
	}
}
