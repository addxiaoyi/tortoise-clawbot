package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server   ServerConfig
	Auth     AuthConfig
	VectorDB VectorDBConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host            string   `json:"host"`
	Port            int      `json:"port"`
	AllowedOrigins  []string `json:"allowed_origins"`
	TLS             TLSConfig `json:"tls"`
}

type TLSConfig struct {
	Enabled bool   `json:"enabled"`
	Cert    string `json:"cert"`
	Key     string `json:"key"`
}

type AuthConfig struct {
	JWTSecret     string         `json:"jwt_secret"`
	SuperTokens   SuperTokensConfig `json:"supertokens"`
}

type SuperTokensConfig struct {
	ConnectionURI string `json:"connection_uri"`
	APIKey        string `json:"api_key"`
}

type VectorDBConfig struct {
	Provider string `json:"provider"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	APIKey   string `json:"api_key"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
