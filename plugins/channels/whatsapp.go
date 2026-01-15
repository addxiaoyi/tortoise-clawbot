// WhatsApp 通道实现
package channels

import (
	"context"
	"fmt"
)

// WhatsAppConfig WhatsApp 配置
type WhatsAppConfig struct {
	PhoneNumber string `json:"phone_number"`
	APIURL      string `json:"api_url"`
	APIVersion  string `json:"api_version"`
}

// WhatsAppChannel WhatsApp 通道
type WhatsAppChannel struct {
	config *WhatsAppConfig
}

// NewWhatsAppChannel 创建 WhatsApp 通道
func NewWhatsAppChannel(config *WhatsAppConfig) *WhatsAppChannel {
	return &WhatsAppChannel{config: config}
}

// Start 启动通道
func (c *WhatsAppChannel) Start(ctx context.Context) error {
	fmt.Println("Starting WhatsApp channel...")
	return nil
}

// Stop 停止通道
func (c *WhatsAppChannel) Stop(ctx context.Context) error {
	fmt.Println("Stopping WhatsApp channel...")
	return nil
}

// Send 发送消息
func (c *WhatsAppChannel) Send(ctx context.Context, msg *WhatsAppMessage) (string, error) {
	return "msg_id", nil
}

// ChannelType 返回通道类型
func (c *WhatsAppChannel) ChannelType() string {
	return "whatsapp"
}

// WhatsAppMessage WhatsApp 消息
type WhatsAppMessage struct {
	To          string `json:"to"`
	Body        string `json:"body"`
	MediaURL    string `json:"media_url,omitempty"`
	ReplyToMsgID string `json:"reply_to,omitempty"`
}
