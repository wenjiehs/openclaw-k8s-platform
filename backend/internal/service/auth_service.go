package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/openclaw/openclaw-saas-platform/internal/api/middleware"
	"github.com/openclaw/openclaw-saas-platform/internal/model"
	"github.com/openclaw/openclaw-saas-platform/pkg/config"
)

// AuthService 认证业务逻辑服务
type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

// TokenResult 登录结果
type TokenResult struct {
	Token    string
	ExpireAt time.Time
	User     *model.User
}

// Login 用户登录，验证用户名密码，返回 JWT Token
func (s *AuthService) Login(username, password string) (*TokenResult, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 生成 JWT Token
	expireAt := time.Now().Add(time.Duration(s.cfg.JWT.ExpireHours) * time.Hour)
	claims := middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "openclaw-saas",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("生成 Token 失败: %w", err)
	}

	return &TokenResult{
		Token:    tokenStr,
		ExpireAt: expireAt,
		User:     &user,
	}, nil
}

// HashPassword 对密码进行 bcrypt 哈希
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
