package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	jwtSecret []byte
}

type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewService(cfg Config) (*Service, error) {
	return &Service{
		jwtSecret: []byte(cfg.JWTSecret),
	}, nil
}

func (s *Service) Register(email, password string) (*User, error) {
	// Hash password
	// In production, use bcrypt or argon2
	hashedPassword := password // TODO: hash password

	user := &User{
		ID:        uuid.New(),
		Email:     email,
		Password:  hashedPassword,
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// TODO: Save to database

	return user, nil
}

func (s *Service) Login(email, password string) (string, error) {
	// TODO: Verify credentials from database
	user := &User{
		ID:    uuid.New(),
		Email: email,
		Role:  "user",
	}

	// Generate JWT
	token, err := s.generateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) Logout(token string) error {
	// In production, blacklist the token
	return nil
}

func (s *Service) Refresh(refreshToken string) (string, error) {
	// TODO: Validate refresh token and generate new access token
	return "", nil
}

func (s *Service) ValidateToken(token string) (*Claims, error) {
	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	claims := &Claims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !t.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *Service) GetUser(userID string) (*User, error) {
	// TODO: Fetch from database
	return &User{
		ID:    uuid.MustParse(userID),
		Email: "user@example.com",
		Role:  "user",
	}, nil
}

func (s *Service) generateToken(user *User) (string, error) {
	claims := &Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tortoise-cloud",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

type Config struct {
	JWTSecret   string
	SuperTokens SuperTokensConfig
}

type SuperTokensConfig struct {
	ConnectionURI string
	APIKey        string
}
