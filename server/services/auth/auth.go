package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/ldap.v3"
)

// ========== 企业认证服务 ==========

// EnterpriseAuthService 企业认证服务
type EnterpriseAuthService struct {
	providers map[string]EnterpriseProvider
	jwtSecret []byte
	db        interface{}
}

// EnterpriseProvider 企业认证提供商
type EnterpriseProvider interface {
	Authenticate(ctx context.Context, credentials Credentials) (*User, error)
	GetUserInfo(ctx context.Context, token string) (*User, error)
}

// User 用户信息
type User struct {
	ID           string            `json:"id"`
	Email        string            `json:"email"`
	Name         string            `json:"name"`
	Username     string            `json:"username"`
	Groups       []string          `json:"groups"`
	Roles        []string          `json:"roles"`
	Department   string            `json:"department"`
	Organization string            `json:"organization"`
	Metadata     map[string]string `json:"metadata"`
}

// Credentials 凭证
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ========== LDAP 认证 ==========

// LDAPProvider LDAP 认证提供商
type LDAPProvider struct {
	Server     string
	Port       int
	BaseDN     string
	BindDN     string
	BindPW     string
	UserFilter string
	GroupFilter string
}

// NewLDAPProvider 创建 LDAP 提供商
func NewLDAPProvider(server string, port int, baseDN, bindDN, bindPW string) *LDAPProvider {
	return &LDAPProvider{
		Server:     server,
		Port:       port,
		BaseDN:     baseDN,
		BindDN:     bindDN,
		BindPW:     bindPW,
		UserFilter: "(uid=%s)",
	}
}

// Authenticate LDAP 认证
func (p *LDAPProvider) Authenticate(ctx context.Context, creds Credentials) (*User, error) {
	// 连接 LDAP 服务器
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", p.Server, p.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	defer conn.Close()

	// 绑定服务账户
	if err := conn.Bind(p.BindDN, p.BindPW); err != nil {
		return nil, fmt.Errorf("failed to bind: %w", err)
	}

	// 搜索用户
	userFilter := fmt.Sprintf(p.UserFilter, ldap.EscapeFilter(creds.Username))
	searchRequest := ldap.NewSearchRequest(
		p.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		userFilter,
		[]string{"dn", "cn", "mail", "uid", "memberOf", "departmentNumber", "o"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search user: %w", err)
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	userDN := result.Entries[0].DN

	// 验证密码
	if err := conn.Bind(userDN, creds.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// 获取用户信息
	entry := result.Entries[0]
	user := &User{
		ID:           entry.GetAttributeValue("uid"),
		Email:        entry.GetAttributeValue("mail"),
		Name:         entry.GetAttributeValue("cn"),
		Username:     entry.GetAttributeValue("uid"),
		Department:   entry.GetAttributeValue("departmentNumber"),
		Organization: entry.GetAttributeValue("o"),
		Groups:       entry.GetAttributeValues("memberOf"),
	}

	return user, nil
}

// GetUserInfo 获取用户信息
func (p *LDAPProvider) GetUserInfo(ctx context.Context, token string) (*User, error) {
	return nil, fmt.Errorf("not implemented")
}

// ========== SAML 认证 ==========

// SAMLProvider SAML 认证提供商
type SAMLProvider struct {
	EntityID         string
	SSOURL           string
	Certificate      *rsa.PublicKey
	CertificatePEM   string
	ACSURL           string
	MetadataURL      string
	LogoutURL        string
}

// SAMLResponse SAML 响应
type SAMLResponse struct {
	XMLName xml.Name `xml:"Response"`
	ID      string   `xml:"ID,attr"`
	Status  string   `xml:"Status>StatusCode,attr"`
	AttrStatement *SAMLAssertion `xml:"Assertion"`
}

// SAMLAssertion SAML 断言
type SAMLAssertion struct {
	Subject    string   `xml:"Subject>NameID"`
	Attributes []string `xml:"AttributeStatement>Attribute"`
}

// NewSAMLProvider 创建 SAML 提供商
func NewSAMLProvider(entityID, ssoURL, acsURL, certificatePEM string) (*SAMLProvider, error) {
	return &SAMLProvider{
		EntityID:       entityID,
		SSOURL:         ssoURL,
		ACSURL:         acsURL,
		CertificatePEM:  certificatePEM,
	}, nil
}

// Authenticate SAML 认证
func (p *SAMLProvider) Authenticate(ctx context.Context, creds Credentials) (*User, error) {
	return nil, fmt.Errorf("use InitiateAuth for SAML")
}

// InitiateAuth 发起 SAML 认证
func (p *SAMLProvider) InitiateAuth(c *gin.Context) error {
	// 生成 SAML AuthnRequest
	request := p.buildAuthnRequest()
	
	// 编码并重定向到 IdP
	encoded := base64.StdEncoding.EncodeToString([]byte(request))
	redirectURL := fmt.Sprintf("%s?SAMLRequest=%s", p.SSOURL, url.QueryEscape(encoded))
	
	c.Redirect(http.StatusFound, redirectURL)
	return nil
}

// HandleACS 处理 ACS 回调
func (p *SAMLProvider) HandleACS(c *gin.Context) (*User, error) {
	samlResponse := c.PostForm("SAMLResponse")
	
	decoded, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}
	
	var response SAMLResponse
	if err := xml.Unmarshal(decoded, &response); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}
	
	if response.Status != "urn:oasis:names:tc:SAML:2.0:status:Success" {
		return nil, fmt.Errorf("SAML authentication failed")
	}
	
	user := &User{
		ID:       response.AttrStatement.Subject,
		Email:    response.AttrStatement.Subject,
		Username: response.AttrStatement.Subject,
	}
	
	return user, nil
}

// GetUserInfo 获取用户信息
func (p *SAMLProvider) GetUserInfo(ctx context.Context, token string) (*User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *SAMLProvider) buildAuthnRequest() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
    ID="%s" Version="2.0" IssueInstant="%s"
    AssertionConsumerServiceURL="%s" ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST">
    <saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">%s</saml:Issuer>
</samlp:AuthnRequest>`,
		generateID(), time.Now().UTC().Format(time.RFC3339),
		p.ACSURL, p.EntityID)
}

// ========== JWT 服务 ==========

// JWTService JWT 服务
type JWTService struct {
	secret []byte
	issuer string
}

// NewJWTService 创建 JWT 服务
func NewJWTService(secret, issuer string) *JWTService {
	return &JWTService{
		secret: []byte(secret),
		issuer: issuer,
	}
}

// GenerateToken 生成 JWT
func (s *JWTService) GenerateToken(user *User, expiration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":    user.ID,
		"name":   user.Name,
		"email":  user.Email,
		"groups": user.Groups,
		"roles":  user.Roles,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(expiration).Unix(),
		"iss":    s.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken 验证 JWT
func (s *JWTService) ValidateToken(tokenString string) (*User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user := &User{
			ID:       claims["sub"].(string),
			Name:     claims["name"].(string),
			Email:    claims["email"].(string),
		}
		
		if groups, ok := claims["groups"].([]interface{}); ok {
			for _, g := range groups {
				user.Groups = append(user.Groups, g.(string))
			}
		}
		
		if roles, ok := claims["roles"].([]interface{}); ok {
			for _, r := range roles {
				user.Roles = append(user.Roles, r.(string))
			}
		}
		
		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken 刷新 JWT
func (s *JWTService) RefreshToken(tokenString string, expiration time.Duration) (string, error) {
	user, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return s.GenerateToken(user, expiration)
}

// ========== 辅助函数 ==========

func generateID() string {
	return fmt.Sprintf("_%d%s", time.Now().UnixNano(), randomString(16))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// ========== 权限检查 ==========

// Permission 权限
type Permission struct {
	Resource string
	Action   string
}

// HasPermission 检查权限
func (u *User) HasPermission(resource, action string) bool {
	for _, role := range u.Roles {
		if hasRolePermission(role, resource, action) {
			return true
		}
	}
	return false
}

func hasRolePermission(role, resource, action string) bool {
	permissions := map[string][]string{
		"admin":     {"*", "*"},
		"user":      {"chat:*", "memory:read", "session:*"},
		"readonly": {"chat:read", "memory:read"},
	}
	
	perms, ok := permissions[role]
	if !ok {
		return false
	}
	
	for _, p := range perms {
		if p == "*" || p == resource+"*" || p == resource+":"+action {
			return true
		}
	}
	return false
}

// HasGroup 检查用户组
func (u *User) HasGroup(group string) bool {
	for _, g := range u.Groups {
		if strings.Contains(g, group) {
			return true
		}
	}
	return false
}
