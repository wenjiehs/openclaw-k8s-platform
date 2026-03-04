package middleware

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/openclaw/openclaw-saas-platform/internal/model"
)

// AuditMiddleware 审计日志中间件
// 自动记录所有关键操作到数据库，满足合规要求
type AuditMiddleware struct {
	db *gorm.DB
}

// NewAuditMiddleware 创建审计日志中间件
func NewAuditMiddleware(db *gorm.DB) *AuditMiddleware {
	return &AuditMiddleware{db: db}
}

// AuditContext 审计上下文，用于在 Handler 中设置审计信息
type AuditContext struct {
	Action       string // 操作类型
	ResourceType string // 资源类型
	ResourceID   string // 资源 ID
	Extra        map[string]interface{} // 额外信息
}

// SetAuditContext 在 Gin Context 中设置审计信息
func SetAuditContext(c *gin.Context, ctx AuditContext) {
	c.Set("audit_context", ctx)
}

// AutoAudit 自动审计中间件（根据 HTTP 方法自动推断操作类型）
func (m *AuditMiddleware) AutoAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 执行后续处理
		c.Next()

		// 只记录非 GET 请求（写操作）
		if c.Request.Method == "GET" {
			return
		}

		// 获取当前用户信息
		userID, _ := GetCurrentUserID(c)
		if userID == 0 {
			return // 未认证的请求不记录
		}

		// 获取审计上下文（Handler 层可以设置更详细的信息）
		auditCtx, _ := c.Get("audit_context")
		var action, resourceType, resourceID string
		var extra string

		if ctx, ok := auditCtx.(AuditContext); ok {
			action = ctx.Action
			resourceType = ctx.ResourceType
			resourceID = ctx.ResourceID
			if ctx.Extra != nil {
				extraBytes, _ := json.Marshal(ctx.Extra)
				extra = string(extraBytes)
			}
		} else {
			// 根据 HTTP 方法自动推断操作类型
			switch c.Request.Method {
			case "POST":
				action = "create"
			case "PUT", "PATCH":
				action = "update"
			case "DELETE":
				action = "delete"
			default:
				action = "unknown"
			}
		}

		// 确定操作结果
		result := "success"
		if c.Writer.Status() >= 400 {
			result = "failed"
		}

		// 记录审计日志（异步，不阻塞请求）
		log := &model.AuditLog{
			UserID:       userID,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Result:       result,
			Extra:        extra,
		}

		// 使用 goroutine 异步写入，避免影响请求响应时间
		go func() {
			_ = startTime // 可用于记录操作耗时
			m.db.Create(log)
		}()
	}
}
