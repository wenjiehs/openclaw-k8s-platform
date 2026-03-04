package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/openclaw/openclaw-saas-platform/internal/api/middleware"
	"github.com/openclaw/openclaw-saas-platform/internal/model"
	"github.com/openclaw/openclaw-saas-platform/pkg/config"
	"github.com/openclaw/openclaw-saas-platform/pkg/logger"
)

// AuthHandler 认证相关 HTTP Handler
type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
	log *logger.Logger
}

// NewAuthHandler 创建认证 Handler
func NewAuthHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg, log: log}
}

// LoginRequest 登录请求体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应体
type LoginResponse struct {
	Token    string     `json:"token"`
	ExpireAt time.Time  `json:"expire_at"`
	User     *UserInfo  `json:"user"`
}

// UserInfo 用户信息（用于返回给前端，不含密码）
type UserInfo struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Role       string `json:"role"`
}

// Login 处理用户登录请求
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误：" + err.Error(),
		})
		return
	}

	// 查找用户
	var user model.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 验证密码（bcrypt 哈希比较）
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 生成 JWT Token
	expireAt := time.Now().Add(time.Duration(h.cfg.JWT.ExpireHours) * time.Hour)
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
	tokenStr, err := token.SignedString([]byte(h.cfg.JWT.Secret))
	if err != nil {
		h.log.Error("生成 JWT Token 失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": LoginResponse{
			Token:    tokenStr,
			ExpireAt: expireAt,
			User: &UserInfo{
				ID:         user.ID,
				Username:   user.Username,
				Email:      user.Email,
				Department: user.Department,
				Role:       user.Role,
			},
		},
	})
}

// Me 获取当前登录用户的信息
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": UserInfo{
			ID:         user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Department: user.Department,
			Role:       user.Role,
		},
	})
}
