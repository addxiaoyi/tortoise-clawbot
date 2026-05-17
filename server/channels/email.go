package channels

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"tortoise/config"
)

// EmailChannel Email 渠道实现
type EmailChannel struct {
	config    config.EmailConfig
	smtpConn *smtp.Client
	imapConn *IMAPClient
	inbox    *EmailInbox
	handlers map[string]EmailHandler
	mu       sync.RWMutex
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// EmailInbox 邮箱收件箱
type EmailInbox struct {
	Emails      []*EmailMessage
	UnreadCount int
	LastSync    time.Time
}

// EmailMessage 邮件消息
type EmailMessage struct {
	ID          string
	From        *mail.Address
	To          []*mail.Address
	CC          []*mail.Address
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []*EmailAttachment
	Date        time.Time
	Read        bool
	Labels      []string
}

// EmailAttachment 邮件附件
type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
	Size        int
}

// IMAPClient IMAP 客户端
type IMAPClient struct {
	host string
	port int
	user string
	pass string
}

// EmailHandler 邮件处理器
type EmailHandler func(*EmailMessage) error

// NewEmailChannel 创建 Email 渠道
func NewEmailChannel(cfg config.EmailConfig) *EmailChannel {
	ctx, cancel := context.WithCancel(context.Background())
	return &EmailChannel{
		config:   cfg,
		handlers: make(map[string]EmailHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Connect 连接到邮件服务器
func (e *EmailChannel) Connect() error {
	if e.config.SMTPEnabled {
		if err := e.connectSMTP(); err != nil {
			return fmt.Errorf("smtp connect failed: %w", err)
		}
	}

	if e.config.IMAPEnabled {
		if err := e.connectIMAP(); err != nil {
			return fmt.Errorf("imap connect failed: %w", err)
		}
	}

	e.running = true
	return nil
}

// connectSMTP 连接 SMTP 服务器
func (e *EmailChannel) connectSMTP() error {
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)

	var err error
	if e.config.SMTPTLS {
		e.smtpConn, err = smtp.Dial(addr)
		if err != nil {
			return err
		}

		tlsConfig := &tls.Config{
			ServerName: e.config.SMTPHost,
		}
		if err := e.smtpConn.StartTLS(tlsConfig); err != nil {
			return err
		}
	} else {
		auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)
		err = smtp.SendMail(addr, auth, e.config.Username, []string{}, []byte("test"))
		if err != nil {
			return err
		}
	}

	if e.config.Username != "" && e.config.Password != "" {
		auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)
		if err := e.smtpConn.Auth(auth); err != nil {
			return err
		}
	}

	return nil
}

// connectIMAP 连接 IMAP 服务器
func (e *EmailChannel) connectIMAP() error {
	e.imapConn = &IMAPClient{
		host: e.config.IMAPHost,
		port: e.config.IMAPPort,
		user: e.config.Username,
		pass: e.config.Password,
	}
	return nil
}

// Start 开始监听邮件
func (e *EmailChannel) Start() error {
	if !e.running {
		if err := e.Connect(); err != nil {
			return err
		}
	}

	if e.config.IMAPEnabled {
		go e.emailLoop()
	}

	return nil
}

// emailLoop 邮件轮询循环
func (e *EmailChannel) emailLoop() {
	pollInterval := time.Duration(e.config.PollInterval) * time.Second
	if pollInterval == 0 {
		pollInterval = 30 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// 初始同步
	e.syncEmails()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.syncEmails()
		}
	}
}

// syncEmails 同步邮件
func (e *EmailChannel) syncEmails() {
	if e.imapConn == nil {
		return
	}

	// TODO: 实现 IMAP 同步逻辑
	e.mu.Lock()
	e.inbox = &EmailInbox{
		Emails:      []*EmailMessage{},
		UnreadCount: 0,
		LastSync:    time.Now(),
	}
	e.mu.Unlock()
}

// SendEmail 发送邮件
func (e *EmailChannel) SendEmail(to []string, subject, body string) error {
	return e.SendHTMLEmail(to, subject, body, body)
}

// SendHTMLEmail 发送 HTML 邮件
func (e *EmailChannel) SendHTMLEmail(to []string, subject, body, htmlBody string) error {
	if e.smtpConn == nil {
		return fmt.Errorf("smtp not connected")
	}

	from := mail.Address{Name: e.config.DisplayName, Address: e.config.Username}

	// 设置发件人
	if err := e.smtpConn.Mail(from.Address); err != nil {
		return fmt.Errorf("set sender failed: %w", err)
	}

	// 设置收件人
	for _, addr := range to {
		if err := e.smtpConn.Rcpt(addr); err != nil {
			return fmt.Errorf("set recipient failed: %w", err)
		}
	}

	// 发送邮件内容
	writer, err := e.smtpConn.Data()
	if err != nil {
		return fmt.Errorf("open data writer failed: %w", err)
	}
	defer writer.Close()

	msg := e.buildEmail(from, to, subject, body, htmlBody)
	_, err = writer.Write([]byte(msg))
	return err
}

// buildEmail 构建邮件内容
func (e *EmailChannel) buildEmail(from mail.Address, to []string, subject, body, htmlBody string) string {
	var sb strings.Builder

	// MIME 头
	sb.WriteString("From: " + from.String() + "\r\n")

	toAddrs := make([]string, len(to))
	for i, addr := range to {
		toAddrs[i] = addr
	}
	sb.WriteString("To: " + strings.Join(toAddrs, ", ") + "\r\n")

	sb.WriteString("Subject: " + subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")

	if htmlBody != "" {
		//  multipart/alternative
		boundary := fmt.Sprintf("----=_Part_%d_%d", time.Now().Unix(), time.Now().UnixNano())
		sb.WriteString(fmt.Sprintf("Content-Type: multipart/alternative;\r\n boundary=\"%s\"\r\n\r\n", boundary))

		// 纯文本
		sb.WriteString("--" + boundary + "\r\n")
		sb.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		sb.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		sb.WriteString(quotedprintable.EncodeReader(strings.NewReader(body)))
		sb.WriteString("\r\n\r\n")

		// HTML
		sb.WriteString("--" + boundary + "\r\n")
		sb.WriteString("Content-Type: text/html; charset=utf-8\r\n")
		sb.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
		sb.WriteString(base64.StdEncoding.EncodeToString([]byte(htmlBody)))
		sb.WriteString("\r\n")

		// 结束边界
		sb.WriteString("--" + boundary + "--\r\n")
	} else {
		sb.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		sb.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		sb.WriteString(quotedprintable.EncodeReader(strings.NewReader(body)))
	}

	return sb.String()
}

// SendEmailWithAttachment 发送带附件的邮件
func (e *EmailChannel) SendEmailWithAttachment(to []string, subject, body string, attachments []struct {
	Filename    string
	ContentType string
	Data        []byte
}) error {
	if e.smtpConn == nil {
		return fmt.Errorf("smtp not connected")
	}

	from := mail.Address{Name: e.config.DisplayName, Address: e.config.Username}

	if err := e.smtpConn.Mail(from.Address); err != nil {
		return err
	}

	for _, addr := range to {
		if err := e.smtpConn.Rcpt(addr); err != nil {
			return err
		}
	}

	writer, err := e.smtpConn.Data()
	if err != nil {
		return err
	}
	defer writer.Close()

	msg := e.buildEmailWithAttachments(from, to, subject, body, attachments)
	_, err = writer.Write([]byte(msg))
	return err
}

// buildEmailWithAttachments 构建带附件的邮件
func (e *EmailChannel) buildEmailWithAttachments(from mail.Address, to []string, subject, body string, attachments []struct {
	Filename    string
	ContentType string
	Data        []byte
}) string {
	var sb strings.Builder

	boundary := fmt.Sprintf("----=_Part_%d_%d", time.Now().Unix(), time.Now().UnixNano())

	sb.WriteString("From: " + from.String() + "\r\n")

	toAddrs := make([]string, len(to))
	for i, addr := range to {
		toAddrs[i] = addr
	}
	sb.WriteString("To: " + strings.Join(toAddrs, ", ") + "\r\n")

	sb.WriteString("Subject: " + subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString(fmt.Sprintf("Content-Type: multipart/mixed;\r\n boundary=\"%s\"\r\n\r\n", boundary))

	// 邮件正文
	sb.WriteString("--" + boundary + "\r\n")
	sb.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	sb.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	sb.WriteString(quotedprintable.EncodeReader(strings.NewReader(body)))
	sb.WriteString("\r\n\r\n")

	// 附件
	for _, att := range attachments {
		sb.WriteString("--" + boundary + "\r\n")

		encodedFilename := base64.StdEncoding.EncodeToString([]byte(att.Filename))
		sb.WriteString(fmt.Sprintf("Content-Type: %s;\r\n", att.ContentType))
		sb.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename*=UTF-8''%s\r\n", encodedFilename))
		sb.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
		sb.WriteString(base64.StdEncoding.EncodeToString(att.Data))
		sb.WriteString("\r\n")
	}

	sb.WriteString("--" + boundary + "--\r\n")

	return sb.String()
}

// SetHandler 设置邮件处理器
func (e *EmailChannel) SetHandler(address string, handler EmailHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[address] = handler
}

// GetInbox 获取收件箱
func (e *EmailChannel) GetInbox() *EmailInbox {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.inbox
}

// GetEmails 获取邮件列表
func (e *EmailChannel) GetEmails(limit, offset int) []*EmailMessage {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.inbox == nil {
		return []*EmailMessage{}
	}

	emails := e.inbox.Emails
	if offset >= len(emails) {
		return []*EmailMessage{}
	}

	end := offset + limit
	if end > len(emails) {
		end = len(emails)
	}

	return emails[offset:end]
}

// MarkAsRead 标记邮件为已读
func (e *EmailChannel) MarkAsRead(emailID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.inbox == nil {
		return nil
	}

	for _, email := range e.inbox.Emails {
		if email.ID == emailID {
			email.Read = true
			if e.inbox.UnreadCount > 0 {
				e.inbox.UnreadCount--
			}
			return nil
		}
	}

	return fmt.Errorf("email not found")
}

// Stop 停止渠道
func (e *EmailChannel) Stop() error {
	e.cancel()
	e.running = false

	if e.smtpConn != nil {
		e.smtpConn.Close()
	}

	if e.imapConn != nil {
		e.imapConn = nil
	}

	return nil
}

// GetUserID 获取用户邮箱
func (e *EmailChannel) GetUserID() string {
	return e.config.Username
}

// 回复邮件
func (e *EmailChannel) ReplyEmail(original *EmailMessage, body string) error {
	if original == nil || original.From == nil {
		return fmt.Errorf("original email is required")
	}

	to := []string{original.From.Address}
	subject := "Re: " + original.Subject

	// 添加引用前缀
	quotedBody := "> " + strings.ReplaceAll(original.Body, "\n", "\n> ")
	replyBody := body + "\n\n--- Original Message ---\n" + quotedBody

	return e.SendEmail(to, subject, replyBody)
}

// 转发邮件
func (e *EmailChannel) ForwardEmail(original *EmailMessage, to []string) error {
	subject := "Fwd: " + original.Subject
	body := fmt.Sprintf("--- Forwarded Message ---\nFrom: %s\nDate: %s\nSubject: %s\n\n%s",
		original.From.String(),
		original.Date.Format("Mon, 02 Jan 2006 15:04:05 -0700"),
		original.Subject,
		original.Body,
	)

	return e.SendEmail(to, subject, body)
}

// 解析邮件地址
func ParseEmailAddress(addr string) (*mail.Address, error) {
	return mail.ParseAddress(addr)
}

// 解析邮件内容类型
func ParseContentType(ct string) (string, map[string]string) {
	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return "", nil
	}
	return mediaType, params
}
