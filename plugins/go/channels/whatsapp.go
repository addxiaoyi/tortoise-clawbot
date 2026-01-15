// WhatsApp channel plugin for Tortoise
package channels

import (
	"context"
	"encoding/json"
	"fmt"
)

// WhatsAppConfig contains WhatsApp plugin configuration
type WhatsAppConfig struct {
	PhoneNumber   string   `json:"phone_number"`
	AccessToken   string   `json:"access_token"`
	BusinessID    string   `json:"business_id"`
	WebhookURL    string   `json:"webhook_url,omitempty"`
	VerifyToken   string   `json:"verify_token,omitempty"`
	AllowedPhones []string `json:"allowed_phones,omitempty"`
}

// WhatsAppPlugin implements the WhatsApp channel plugin
type WhatsAppPlugin struct {
	config WhatsAppConfig
}

// NewWhatsAppPlugin creates a new WhatsApp plugin
func NewWhatsAppPlugin() *WhatsAppPlugin {
	return &WhatsAppPlugin{}
}

// Metadata returns the plugin metadata
func (p *WhatsAppPlugin) Metadata() PluginMetadata {
	return PluginMetadata{
		ID:          "whatsapp-channel",
		Name:        "WhatsApp Channel",
		Version:     "1.0.0",
		Description: "WhatsApp Business API integration",
		Author:      "Tortoise Team",
		License:     "MIT",
		Type:        TypeChannel,
		Tags:        []string{"whatsapp", "messaging", "social", "business"},
	}
}

// Initialize initializes the plugin with configuration
func (p *WhatsAppPlugin) Initialize(ctx context.Context, config json.RawMessage) error {
	if err := json.Unmarshal(config, &p.config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	
	if p.config.AccessToken == "" {
		return fmt.Errorf("WhatsApp access token is required")
	}
	
	return nil
}

// Start starts the WhatsApp connection
func (p *WhatsAppPlugin) Start(ctx context.Context) error {
	fmt.Println("Starting WhatsApp plugin...")
	return nil
}

// Stop stops the WhatsApp connection
func (p *WhatsAppPlugin) Stop(ctx context.Context) error {
	fmt.Println("Stopping WhatsApp plugin...")
	return nil
}

// Execute handles WhatsApp-specific operations
func (p *WhatsAppPlugin) Execute(ctx context.Context, method string, args json.RawMessage) (json.RawMessage, error) {
	switch method {
	case "send_message":
		return p.sendMessage(args)
	case "send_template":
		return p.sendTemplate(args)
	case "send_media":
		return p.sendMedia(args)
	case "mark_read":
		return p.markRead(args)
	case "get_profile":
		return p.getProfile(args)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

type WhatsAppSendMessageArgs struct {
	To       string `json:"to"`
	Body     string `json:"body"`
	PreviewURL bool  `json:"preview_url,omitempty"`
}

type WhatsAppTemplateArgs struct {
	To       string                 `json:"to"`
	Template string                 `json:"template"`
	LangCode string                 `json:"lang_code,omitempty"`
	Components []TemplateComponent  `json:"components,omitempty"`
}

type TemplateComponent struct {
	Type    string            `json:"type"`
	SubType string           `json:"sub_type,omitempty"`
	Index   int              `json:"index,omitempty"`
	Parameters []Parameter    `json:"parameters,omitempty"`
}

type Parameter struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type WhatsAppMediaArgs struct {
	To       string `json:"to"`
	MediaURL string `json:"media_url"`
	Caption  string `json:"caption,omitempty"`
	MediaType string `json:"media_type"` // image