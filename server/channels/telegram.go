// Telegram channel implementation

package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken     string
	APIID       string
	APIHash     string
	AllowedUsers []int64
}

// TelegramChannel implements ChannelHandler for Telegram
type TelegramChannel struct {
	config TelegramConfig
	client *http.Client
}

// NewTelegramChannel creates a new Telegram channel handler
func NewTelegramChannel(config TelegramConfig) *TelegramChannel {
	return &TelegramChannel{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChannelType returns the channel type
func (t *TelegramChannel) ChannelType() ChannelType {
	return ChannelTypeTelegram
}

// Name returns the channel name
func (t *TelegramChannel) Name() string {
	return "telegram"
}

// Connect connects to Telegram
func (t *TelegramChannel) Connect(ctx context.Context) error {
	// Set webhook or long polling
	// For now, just verify the bot token
	resp, err := t.client.Get(fmt.Sprintf(
		"https://api.telegram.org/bot%s/getMe",
		t.config.BotToken,
	))
	if err != nil {
		return fmt.Errorf("failed to connect to Telegram: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API error: %d", resp.StatusCode)
	}

	var result struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("bot token invalid")
	}

	log.Info().Msg("Connected to Telegram")
	return nil
}

// Disconnect disconnects from Telegram
func (t *TelegramChannel) Disconnect() error {
	// Cleanup webhook or polling
	log.Info().Msg("Disconnected from Telegram")
	return nil
}

// Status returns the connection status
func (t *TelegramChannel) Status() ChannelStatus {
	return ChannelStatusConnected
}

// Send sends a message
func (t *TelegramChannel) Send(ctx context.Context, msg *ChannelMessage) error {
	// Parse Telegram chat ID and message
	chatID := msg.ChannelID
	content := msg.Content

	// Send via Telegram API
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		t.config.BotToken,
	)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    content,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(
		io.NopCloser(nil)),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API error: %d", resp.StatusCode)
	}

	return nil
}

// SendTyping sends a typing indicator
func (t *TelegramChannel) SendTyping(ctx context.Context, userID string, typing bool) error {
	// Telegram doesn't support typing indicators directly
	// Could use "sendChatAction" with "typing"
	return nil
}

// Subscribe returns the event channel
func (t *TelegramChannel) Subscribe() chan *ChannelEvent {
	// Would implement long polling or webhook handler
	return make(chan *ChannelEvent)
}

// HandleUpdate processes an incoming Telegram update
func (t *TelegramChannel) HandleUpdate(update *TelegramUpdate) error {
	// Convert Telegram update to ChannelMessage
	if update.Message == nil {
		return nil
	}

	msg := &ChannelMessage{
		ID:          fmt.Sprintf("%d", update.Message.MessageID),
		ChannelType: ChannelTypeTelegram,
		ChannelID:  fmt.Sprintf("%d", update.Message.Chat.ID),
		UserID:     fmt.Sprintf("%d", update.Message.From.ID),
		UserName:   update.Message.From.Username,
		Content:    update.Message.Text,
		Timestamp:  time.Unix(update.Message.Date, 0),
	}

	// Check if user is allowed
	if len(t.config.AllowedUsers) > 0 {
		allowed := false
		for _, id := range t.config.AllowedUsers {
			if id == update.Message.From.ID {
				allowed = true
				break
			}
		}
		if !allowed {
			log.Warn().Int64("user_id", update.Message.From.ID).Msg("User not allowed")
			return nil
		}
	}

	// Send to event channel
	// In real implementation, would use a channel
	return nil
}

// Telegram API types

type TelegramUpdate struct {
	UpdateID int64          `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

type TelegramMessage struct {
	MessageID int64           `json:"message_id"`
	From     TelegramUser     `json:"from"`
	Chat     TelegramChat     `json:"chat"`
	Text     string           `json:"text,omitempty"`
	Date     int64            `json:"date"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

type TelegramChat struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	Type      string `json:"type"`
}
