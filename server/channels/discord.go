// Discord channel implementation

package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// DiscordConfig holds Discord bot configuration
type DiscordConfig struct {
	BotToken        string
	AllowedGuilds   []string
	AllowedChannels []string
}

// DiscordChannel implements ChannelHandler for Discord
type DiscordChannel struct {
	config  DiscordConfig
	client  *http.Client
	wsConn  *websocket.Conn
	session *DiscordSession
}

// NewDiscordChannel creates a new Discord channel handler
func NewDiscordChannel(config DiscordConfig) *DiscordChannel {
	return &DiscordChannel{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChannelType returns the channel type
func (d *DiscordChannel) ChannelType() ChannelType {
	return ChannelTypeDiscord
}

// Name returns the channel name
func (d *DiscordChannel) Name() string {
	return "discord"
}

// Connect connects to Discord
func (d *DiscordChannel) Connect(ctx context.Context) error {
	// Get gateway URL
	gateway, err := d.getGateway()
	if err != nil {
		return fmt.Errorf("failed to get gateway: %w", err)
	}

	// Connect to WebSocket
	header := http.Header{}
	header.Set("Authorization", "Bot "+d.config.BotToken)

	conn, _, err := websocket.DefaultDialer.Dial(gateway+"/?v=10&encoding=json", header)
	if err != nil {
		return fmt.Errorf("failed to connect to Discord gateway: %w", err)
	}

	d.wsConn = conn

	// Start reading messages
	go d.readMessages()

	log.Info().Msg("Connected to Discord")
	return nil
}

// Disconnect disconnects from Discord
func (d *DiscordChannel) Disconnect() error {
	if d.wsConn != nil {
		d.wsConn.Close()
	}
	log.Info().Msg("Disconnected from Discord")
	return nil
}

// Status returns the connection status
func (d *DiscordChannel) Status() ChannelStatus {
	if d.wsConn != nil {
		return ChannelStatusConnected
	}
	return ChannelStatusDisconnected
}

// Send sends a message
func (d *DiscordChannel) Send(ctx context.Context, msg *ChannelMessage) error {
	// Send via Discord REST API
	url := fmt.Sprintf(
		"https://discord.com/api/v10/channels/%s/messages",
		msg.ChannelID,
	)

	payload := map[string]interface{}{
		"content": msg.Content,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bot "+d.config.BotToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Discord API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendTyping sends a typing indicator
func (d *DiscordChannel) SendTyping(ctx context.Context, userID string, typing bool) error {
	if !typing {
		return nil
	}

	url := fmt.Sprintf(
		"https://discord.com/api/v10/channels/%s/typing",
		userID, // Channel ID
	)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+d.config.BotToken)

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Subscribe returns the event channel
func (d *DiscordChannel) Subscribe() chan *ChannelEvent {
	return make(chan *ChannelEvent)
}

// Helper methods

func (d *DiscordChannel) getGateway() (string, error) {
	url := "https://discord.com/api/v10/gateway/bot"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bot "+d.config.BotToken)

	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.URL, nil
}

func (d *DiscordChannel) readMessages() {
	for {
		_, msg, err := d.wsConn.ReadMessage()
		if err != nil {
			log.Error().Err(err).Msg("Discord read error")
			break
		}

		var event DiscordEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			continue
		}

		d.handleEvent(&event)
	}
}

func (d *DiscordChannel) handleEvent(event *DiscordEvent) {
	switch event.Op {
	case DiscordOpDispatch:
		d.handleDispatch(event)
	}
}

func (d *DiscordChannel) handleDispatch(event *DiscordEvent) {
	switch event.Type {
	case "MESSAGE_CREATE":
		var msg DiscordMessage
		if err := json.Unmarshal(event.Data, &msg); err != nil {
			return
		}
		d.handleMessage(&msg)
	}
}

func (d *DiscordChannel) handleMessage(msg *DiscordMessage) {
	if msg.Author.Bot {
		return // Ignore bot messages
	}

	channel := &ChannelMessage{
		ID:          msg.ID,
		ChannelType: ChannelTypeDiscord,
		ChannelID:   msg.ChannelID,
		UserID:      msg.Author.ID,
		UserName:    msg.Author.Username,
		Content:     msg.Content,
		Timestamp:   time.Unix(msg.Timestamp, 0),
	}

	// Send to channel
	// In real implementation, would use the Subscribe channel
	log.Info().Str("user", channel.UserName).Str("content", channel.Content).Msg("Discord message")
}

// Discord API types

type DiscordSession struct {
	SessionID string `json:"session_id"`
	Seq       int    `json:"seq"`
}

type DiscordEvent struct {
	Op   int             `json:"op"`
	Type string          `json:"t"`
	Seq  int             `json:"s"`
	Data json.RawMessage `json:"d"`
}

type DiscordMessage struct {
	ID        string          `json:"id"`
	ChannelID string          `json:"channel_id"`
	Content   string          `json:"content"`
	Timestamp int64           `json:"timestamp"`
	Author    DiscordUser     `json:"author"`
	Mentions  []DiscordUser   `json:"mentions"`
	Member    *DiscordMember   `json:"member,omitempty"`
}

type DiscordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}

type DiscordMember struct {
	Nick string `json:"nick"`
}

const (
	DiscordOpDispatch             = 0
	DiscordOpHeartbeat            = 1
	DiscordOpIdentify            = 2
	DiscordOpPresenceUpdate      = 3
	DiscordOpVoiceStateUpdate    = 4
	DiscordOpResume              = 6
	DiscordOpReconnect           = 7
	DiscordOpRequestGuildMembers = 8
	DiscordOpInvalidSession      = 9
	DiscordOpHello               = 10
	DiscordOpHeartbeatAck        = 11
)
