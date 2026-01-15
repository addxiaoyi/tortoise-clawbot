package gateway

import (
	"time"
)

// Config - Gateway 配置
type Config struct {
	BindAddress        string
	Port               int
	TLSEnabled         bool
	TLSCert            string
	TLSKey             string
	MaxConnections     int
	ConnectionTimeout  time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	MaxSessions        int
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		BindAddress:       "0.0.0.0",
		Port:              18792,
		TLSEnabled:        false,
		MaxConnections:    10000,
		ConnectionTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		MaxSessions:       10000,
	}
}
