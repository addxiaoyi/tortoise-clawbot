package channel

import "errors"

var (
	// ErrChannelNotFound 渠道未找到
	ErrChannelNotFound = errors.New("channel not found")
	
	// ErrChannelExists 渠道已存在
	ErrChannelExists = errors.New("channel already exists")
	
	// ErrChannelClosed 渠道已关闭
	ErrChannelClosed = errors.New("channel closed")
	
	// ErrInvalidCredentials 无效的凭证
	ErrInvalidCredentials = errors.New("invalid credentials")
	
	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errors.New("connection failed")
)
