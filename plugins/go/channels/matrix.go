// Matrix channel plugin for Tortoise
package channels

import (
	"context"
	"encoding/json"
	"fmt"
)

// MatrixConfig contains Matrix plugin configuration
type MatrixConfig struct {
	HomeserverURL string `json:"homeserver_url"`
	AccessToken   string `json:"access_token"`
	UserID       string `json:"user_id,omitempty"`
	DeviceID     string `json:"device_id,omitempty"`
}

// MatrixPlugin implements the Matrix channel plugin
type MatrixPlugin struct {
	config MatrixConfig
}

// NewMatrixPlugin creates a new Matrix plugin
func NewMatrixPlugin() *MatrixPlugin {
	return &MatrixPlugin{}
}

// Metadata returns the plugin metadata
func (p *MatrixPlugin) Metadata() PluginMetadata {
	return PluginMetadata{
		ID:          "matrix-channel",
		Name:        "Matrix Channel",
		Version:     "1.0.0",
		Description: "Matrix/Element messaging channel integration",
		Author:      "Tortoise Team",
		License:     "MIT",
		Type:        TypeChannel,
		Tags:        []string{"matrix", "element", "messaging", "federation"},
	}
}

// Initialize initializes the plugin with configuration
func (p *MatrixPlugin) Initialize(ctx context.Context, config json.RawMessage) error {
	if err := json.Unmarshal(config, &p.config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	
	if p.config.HomeserverURL == "" {
		return fmt.Errorf("Matrix homeserver URL is required")
	}
	
	if p.config.AccessToken == "" {
		return fmt.Errorf("Matrix access token is required")
	}
	
	return nil
}

// Start starts the Matrix connection
func (p *MatrixPlugin) Start(ctx context.Context) error {
	fmt.Println("Starting Matrix plugin...")
	return nil
}

// Stop stops the Matrix connection
func (p *MatrixPlugin) Stop(ctx context.Context) error {
	fmt.Println("Stopping Matrix plugin...")
	return nil
}

// Execute handles Matrix-specific operations
func (p *MatrixPlugin) Execute(ctx context.Context, method string, args json.RawMessage) (json.RawMessage, error) {
	switch method {
	case "send_message":
		return p.sendMessage(args)
	case "send_reply":
		return p.sendReply(args)
	case "join_room":
		return p.joinRoom(args)
	case "leave_room":
		return p.leaveRoom(args)
	case "get_room_state":
		return p.getRoomState(args)
	case "upload_media":
		return p.uploadMedia(args)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

type MatrixSendMessageArgs struct {
	RoomID string `json:"room_id"`
	Body   string `json:"body"`
	Format string `json:"format,omitempty"`
	HTMLBody string `json:"html_body,omitempty"`
}

type MatrixSendReplyArgs struct {
	RoomID   string `json:"room_id"`
	Body     string `json:"body"`
	RelatesTo string `json:"relates_to"`
	Format   string `json:"format,omitempty"`
	HTMLBody string `json:"html_body,omitempty"`
}

type MatrixJoinRoomArgs struct {
	RoomIDOrAlias string `json:"room_id_or_alias"`
}

type MatrixLeaveRoomArgs struct {
	RoomID string `json:"room_id"`
	Reason string `json:"reason,omitempty"`
}

type MatrixGetRoomStateArgs struct {
	RoomID string `json:"room_id"`
}

type MatrixUploadMediaArgs struct {
	Filename string `json:"filename"`
	Content  string `json:"content"` // Base64 encoded
	MimeType string `json:"mime_type,omitempty"`
}

func (p *MatrixPlugin) sendMessage(args json.RawMessage) (json.RawMessage, error) {
	var msgArgs MatrixSendMessageArgs
	if err := json.Unmarshal(args, &msgArgs); err != nil {
		return nil, fmt.Errorf("failed to parse args: %w", err)
	}
	
	return json.Marshal(map[string]interface{}{
		"success":       true,
		"room_id":      msgArgs.RoomID,
		"event_id":     "$" + "random_event_id",
		"timestamp":    1234567890000,
	})
}

func (p *MatrixPlugin) sendReply(args json.RawMessage) (json.RawMessage, error) {
	var replyArgs MatrixSendReplyArgs
	if err := json.Unmarshal(args, &replyArgs); err != nil {
		return nil, fmt.Errorf("failed to parse args: %w", err)
	}
	
	return json.Marshal(map[string]interface{}{
		"success":       true,
		"room_id":      replyArgs.RoomID,
		"event_id":     "$" + "random_reply_id",
		"timestamp":    1234567890000,
	})
}

func (p *MatrixPlugin) joinRoom(args json.RawMessage) (json.RawMessage, error) {
	var joinArgs MatrixJoinRoomArgs
	if err := json.Unmarshal(args, &joinArgs); err != nil {
		return nil, fmt.Errorf("failed to parse args: %w", err)
	}
	
	return json.Marshal(map[string]interface{}{
		"success":       true,
		"room_id":      "!random:matrix.org",
		"room_alias":   joinArgs.RoomIDOrAlias,
	})
}

func (p *MatrixPlugin) leaveRoom(args json.RawMessage) (json.RawMessage, error) {
	var leaveArgs MatrixLeaveRoomArgs
	if err := json.Unmarshal(args, &leaveArgs); err != nil {
		return nil, fmt.Errorf("failed to parse args: %w", err)
	}
	
	return json.Marshal(map[string]interface{}{
		"success": true,
		"room_id": leaveArgs.RoomID,
	})
}

func (p *MatrixPlugin) getRoomState(args json.RawMessage) (json.RawMessage, error) {
	var stateArgs MatrixGetRoomStateArgs
	if err := json.Unmarshal(args, &stateArgs); err != nil {
		return nil, fmt.Errorf("failed to parse args: %w", err)
	}
	
	return json.Marshal(map[string]interface{}{
		"room_id": stateArgs.RoomID,
		"state": []map[string]interface{}{
			{
				"type":      "m.room.name",
				"state_key": "",
				"content":   map[string]string{"name": "Tortoise Room"},
			},
		},
	})
}

func (p *MatrixPlugin) uploadMedia(args json.RawMessage) (json.RawMessage, error) {
	var uploadArgs MatrixUploadMediaArgs
	if err := json.Unmarshal(args, &uploadArgs); err != nil {
		return nil, fmt.Errorf("failed to parse args: %w", err)
	}
	
	return json.Marshal(map[string]interface{}{
		"success":     true,
		"content_uri": "mxc://matrix.org/" + "random_media_id",
		"filename":   uploadArgs.Filename,
	})
}
