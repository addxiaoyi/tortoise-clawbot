package channel

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
	"sync"
	"time"

	"tortoise-server/internal/ai"
)

// ============ Email Channel (SMTP/IMAP) ============

// EmailChannel Email 渠道
type EmailChannel struct {
	config      *EmailConfig
	aiEngine   *ai.Engine
	running    bool
	mu         sync.RWMutex
	httpClient *http.Client
	messageQueue chan *EmailMessage
}

// EmailConfig Email 配置
type EmailConfig struct {
	// IMAP 配置 (接收邮件)
	IMAPHost     string // imap.gmail.com
	IMAPPort     int    // 993
	IMAPUser     string
	IMAPPassword string
	UseTLS      bool
	
	// SMTP 配置 (发送邮件)
	SMTPHost     string // smtp.gmail.com
	SMTPPort     int    // 587 (TLS) 或 465 (SSL)
	SMTPUser     string
	SMTPPassword string
	UseStartTLS bool   // true for port 587, false for port 465
	
	// 应用配置
	EmailAddress string // bot@example.com
	BotName     string // Tortoise Bot
	PollInterval time.Duration // 邮件轮询间隔
	
	// 过滤配置
	AllowedSenders []string // 允许的发件人
	BlockedSubjects []string // 屏蔽的主题关键词
	
	// AI 配置
	AIProvider string
	AIModel   string
}

// EmailMessage Email 消息
type EmailMessage struct {
	ID          string
	From        string
	To          string
	Subject     string
	Body        string
	HTMLBody    string
	Date        time.Time
	HasAttachments bool
	Attachments   []EmailAttachment
	Headers     map[string]string
	IsRead      bool
	IsReplied   bool
	IsStarred   bool
	Labels      []string
}

// EmailAttachment Email 附件
type EmailAttachment struct {
	Filename    string
	ContentType string
	Size        int
	Data        []byte
}

// EmailFolder Email 文件夹
type EmailFolder struct {
	Name      string
	Path      string
	Total     int
	Unread    int
}

// NewEmailChannel 创建 Email 渠道
func NewEmailChannel(config *EmailConfig) *EmailChannel {
	if config.PollInterval == 0 {
		config.PollInterval = 60 * time.Second
	}
	if config.IMAPPort == 0 {
		config.IMAPPort = 993
	}
	if config.SMTPPort == 0 {
		config.SMTPPort = 587
	}
	
	return &EmailChannel{
		config: config,
		messageQueue: make(chan *EmailMessage, 100),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetAIEngine 设置 AI 引擎
func (c *EmailChannel) SetAIEngine(engine *ai.Engine) {
	c.aiEngine = engine
}

// Start 启动 Email 渠道
func (c *EmailChannel) Start() error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	// 启动轮询
	go c.pollEmails()
	// 启动消息处理器
	go c.processMessages()

	log.Printf("✅ Email 渠道已启动 (%s)", c.config.EmailAddress)
	return nil
}

// Stop 停止 Email 渠道
func (c *EmailChannel) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.running {
		return
	}
	c.running = false
	
	close(c.messageQueue)
	log.Printf("🛑 Email 渠道已停止")
}

// pollEmails 轮询新邮件
func (c *EmailChannel) pollEmails() {
	ticker := time.NewTicker(c.config.PollInterval)
	defer ticker.Stop()

	for c.mu.RLock(); c.running; c.mu.RUnlock() {
		select {
		case <-ticker.C:
			c.fetchNewEmails()
		}
	}
}

// fetchNewEmails 获取新邮件
func (c *EmailChannel) fetchNewEmails() {
	// 使用 IMAP 获取新邮件
	// 简化实现 - 实际使用 go-imap 库
	
	log.Printf("📧 检查新邮件...")
	
	// 测试连接
	if err := c.testIMAPConnection(); err != nil {
		log.Printf("❌ IMAP 连接失败: %v", err)
		return
	}
	
	log.Printf("✅ 邮箱连接正常")
}

// testIMAPConnection 测试 IMAP 连接
func (c *EmailChannel) testIMAPConnection() error {
	// 构建 IMAP 地址
	addr := fmt.Sprintf("%s:%d", c.config.IMAPHost, c.config.IMAPPort)
	
	// 连接
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: c.config.IMAPHost,
	})
	if err != nil {
		return err
	}
	defer conn.Close()
	
	return nil
}

// processMessages 处理消息队列
func (c *EmailChannel) processMessages() {
	for msg := range c.messageQueue {
		// 过滤
		if !c.shouldProcessEmail(msg) {
			continue
		}
		
		// AI 处理
		go c.handleEmail(msg)
	}
}

// shouldProcessEmail 判断是否应该处理邮件
func (c *EmailChannel) shouldProcessEmail(msg *EmailMessage) bool {
	// 检查发件人白名单
	if len(c.config.AllowedSenders) > 0 {
		allowed := false
		for _, sender := range c.config.AllowedSenders {
			if strings.Contains(strings.ToLower(msg.From), strings.ToLower(sender)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}
	
	// 检查屏蔽主题
	for _, keyword := range c.config.BlockedSubjects {
		if strings.Contains(strings.ToLower(msg.Subject), strings.ToLower(keyword)) {
			return false
		}
	}
	
	return true
}

// handleEmail 处理邮件
func (c *EmailChannel) handleEmail(msg *EmailMessage) {
	var response string
	
	if c.aiEngine != nil {
		// 构建提示
		prompt := fmt.Sprintf("用户发送邮件:\n主题: %s\n内容: %s\n\n请回复此邮件。", 
			msg.Subject, msg.Body)
		
		req := &ai.ChatRequest{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   2048,
			Messages: []ai.Message{
				{Role: "user", Content: prompt},
			},
		}
		
		resp, err := c.aiEngine.Chat(nil, req)
		if err != nil {
			response = fmt.Sprintf("抱歉，处理您的邮件时出错: %v", err)
		} else {
			response = resp.Content
		}
	} else {
		response = "AI 服务未配置"
	}
	
	// 自动回复
	if err := c.sendReply(msg, response); err != nil {
		log.Printf("❌ 自动回复失败: %v", err)
	}
}

// sendReply 发送回复
func (c *EmailChannel) sendReply(original *EmailMessage, body string) error {
	// 构建回复邮件
	reply := &EmailMessage{
		To:      original.From,
		Subject: fmt.Sprintf("Re: %s", original.Subject),
		Body:    body,
	}
	
	return c.SendEmail(reply)
}

// SendEmail 发送邮件
func (c *EmailChannel) SendEmail(msg *EmailMessage) error {
	if msg.To == "" {
		return fmt.Errorf("收件人不能为空")
	}
	
	// 构建邮件内容
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", c.config.BotName, c.config.EmailAddress)
	headers["To"] = msg.To
	headers["Subject"] = msg.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	
	var content strings.Builder
	for k, v := range headers {
		content.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	content.WriteString("\r\n")
	content.WriteString(msg.Body)
	
	// SMTP 发送
	return c.sendViaSMTP(msg.To, content.String())
}

// sendViaSMTP 通过 SMTP 发送
func (c *EmailChannel) sendViaSMTP(to, body string) error {
	auth := smtp.PlainAuth("", c.config.SMTPUser, c.config.SMTPPassword, c.config.SMTPHost)
	
	addr := fmt.Sprintf("%s:%d", c.config.SMTPHost, c.config.SMTPPort)
	
	if c.config.UseStartTLS {
		// STARTTLS
		err := smtp.SendMail(addr, auth, c.config.EmailAddress, []string{to}, []byte(body))
		if err != nil {
			// 尝试 TLS
			return c.sendViaSMTPTLS(to, body, auth)
		}
		return err
	}
	
	// 直接 TLS
	return c.sendViaSMTPTLS(to, body, auth)
}

// sendViaSMTPTLS 通过 TLS SMTP 发送
func (c *EmailChannel) sendViaSMTPTLS(to, body string, auth smtp.Auth) error {
	tlsConfig := &tls.Config{
		ServerName: c.config.SMTPHost,
	}
	
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.config.SMTPHost, c.config.SMTPPort), tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()
	
	client, err := smtp.NewClient(conn, c.config.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()
	
	// 认证
	if err := client.Auth(auth); err != nil {
		return err
	}
	
	// 发件人
	if err := client.Mail(c.config.EmailAddress); err != nil {
		return err
	}
	
	// 收件人
	if err := client.Rcpt(to); err != nil {
		return err
	}
	
	// 邮件内容
	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()
	
	_, err = w.Write([]byte(body))
	return err
}

// SendHTMLEmail 发送 HTML 邮件
func (c *EmailChannel) SendHTMLEmail(to, subject, htmlBody string) error {
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", c.config.BotName, c.config.EmailAddress)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""
	
	var content strings.Builder
	for k, v := range headers {
		content.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	content.WriteString("\r\n")
	content.WriteString(htmlBody)
	
	return c.sendViaSMTP(to, content.String())
}

// SendEmailWithAttachments 发送带附件的邮件
func (c *EmailChannel) SendEmailWithAttachments(to, subject, body string, attachments []EmailAttachment) error {
	// 构建 multipart 邮件
	boundary := generateBoundary()
	
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", c.config.BotName, c.config.EmailAddress)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=\"%s\"", boundary)
	
	var content strings.Builder
	for k, v := range headers {
		content.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	content.WriteString("\r\n")
	
	// 文本部分
	content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	content.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	content.WriteString("\r\n")
	content.WriteString(body)
	content.WriteString("\r\n")
	
	// 附件
	for _, att := range attachments {
		content.WriteString(fmt.Sprintf("\n--%s\r\n", boundary))
		content.WriteString(fmt.Sprintf("Content-Type: %s\r\n", att.ContentType))
		content.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename))
		content.WriteString("Content-Transfer-Encoding: base64\r\n")
		content.WriteString("\r\n")
		
		// Base64 编码
		encoded := base64.StdEncoding.EncodeToString(att.Data)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			content.WriteString(encoded[i:end])
			content.WriteString("\r\n")
		}
		content.WriteString("\r\n")
	}
	
	// 结束标记
	content.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	
	return c.sendViaSMTP(to, content.String())
}

// GetFolders 获取邮箱文件夹列表
func (c *EmailChannel) GetFolders() ([]EmailFolder, error) {
	folders := []EmailFolder{
		{Name: "INBOX", Path: "INBOX", Total: 0, Unread: 0},
		{Name: "Sent", Path: "[Gmail]/Sent Mail", Total: 0, Unread: 0},
		{Name: "Drafts", Path: "[Gmail]/Drafts", Total: 0, Unread: 0},
		{Name: "Trash", Path: "[Gmail]/Trash", Total: 0, Unread: 0},
		{Name: "Spam", Path: "[Gmail]/Spam", Total: 0, Unread: 0},
		{Name: "Starred", Path: "[Gmail]/Starred", Total: 0, Unread: 0},
	}
	
	return folders, nil
}

// MarkAsRead 标记邮件为已读
func (c *EmailChannel) MarkAsRead(emailID string) error {
	log.Printf("📧 标记邮件 %s 为已读", emailID)
	return nil
}

// MarkAsUnread 标记邮件为未读
func (c *EmailChannel) MarkAsUnread(emailID string) error {
	log.Printf("📧 标记邮件 %s 为未读", emailID)
	return nil
}

// StarEmail 标星邮件
func (c *EmailChannel) StarEmail(emailID string) error {
	log.Printf("📧 标星邮件 %s", emailID)
	return nil
}

// UnstarEmail 取消标星
func (c *EmailChannel) UnstarEmail(emailID string) error {
	log.Printf("📧 取消标星邮件 %s", emailID)
	return nil
}

// MoveToFolder 移动邮件到文件夹
func (c *EmailChannel) MoveToFolder(emailID, folder string) error {
	log.Printf("📧 移动邮件 %s 到 %s", emailID, folder)
	return nil
}

// DeleteEmail 删除邮件
func (c *EmailChannel) DeleteEmail(emailID string) error {
	log.Printf("📧 删除邮件 %s", emailID)
	return nil
}

// SearchEmails 搜索邮件
func (c *EmailChannel) SearchEmails(query string) ([]*EmailMessage, error) {
	log.Printf("📧 搜索邮件: %s", query)
	return []*EmailMessage{}, nil
}

// SendTemplate 发送邮件模板
func (c *EmailChannel) SendTemplate(to, templateID string, data map[string]string) error {
	templates := map[string]string{
		"welcome": "欢迎 {{name}} 加入 Tortoise！\n\n您的账户已创建成功。",
		"reset": "您好 {{name}}，\n\n请点击以下链接重置密码: {{link}}\n\n链接有效期: 24小时",
		"notification": "{{name}}，您有新的通知:\n\n{{message}}",
	}
	
	template, ok := templates[templateID]
	if !ok {
		return fmt.Errorf("未找到模板: %s", templateID)
	}
	
	// 替换变量
	body := template
	for k, v := range data {
		body = strings.ReplaceAll(body, fmt.Sprintf("{{%s}}", k), v)
	}
	
	return c.SendEmail(&EmailMessage{
		To:      to,
		Subject: "Tortoise 通知",
		Body:    body,
	})
}

// generateBoundary 生成 MIME boundary
func generateBoundary() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// ValidateEmail 验证邮箱地址
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ParseEmailAddress 解析邮箱地址
func ParseEmailAddress(raw string) (name, address string) {
	// 格式: "Name <email@example.com>" 或 "email@example.com"
	if strings.Contains(raw, "<") {
		parts := strings.Split(raw, "<")
		name = strings.TrimSpace(parts[0])
		address = strings.Trim(strings.TrimSpace(parts[1]), ">")
	} else {
		address = strings.TrimSpace(raw)
	}
	return
}
