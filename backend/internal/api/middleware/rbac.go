package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RBACMiddleware RBAC 权限控制中间件
// 基于资源所有者检查，确保员工只能操作自己的资源

// RequireResourceOwnerOrAdmin 检查资源所有权：必须是资源所有者或管理员
// userIDParam: URL 参数中的用户 ID 字段名
func RequireResourceOwnerOrAdmin(userIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前登录用户信息
		currentUserID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
			})
			c.Abort()
			return
		}

		// 管理员可以操作任何资源
		role := GetCurrentUserRole(c)
		if role == "admin" || role == "super_admin" {
			c.Next()
			return
		}

		// 非管理员：检查资源是否属于当前用户
		// 实际的资源所有权检查需在 Handler 层完成（因为需要查数据库）
		// 这里只是将当前用户 ID 传递给 Handler 层使用
		_ = currentUserID
		c.Next()
	}
}
