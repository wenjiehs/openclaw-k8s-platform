package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openclaw/openclaw-saas-platform/internal/api/handler"
	"github.com/openclaw/openclaw-saas-platform/internal/api/middleware"
	"github.com/openclaw/openclaw-saas-platform/internal/k8s"
	"github.com/openclaw/openclaw-saas-platform/internal/service"
	"github.com/openclaw/openclaw-saas-platform/pkg/config"
	"github.com/openclaw/openclaw-saas-platform/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewRouter 创建并配置 Gin 路由器
func NewRouter(cfg *config.Config, log *logger.Logger) (*gin.Engine, error) {
	// 根据运行模式设置 Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎（使用默认中间件：Logger + Recovery）
	r := gin.New()

	// 添加请求日志中间件（记录每个 HTTP 请求）
	r.Use(gin.Logger())

	// 添加恢复中间件（防止 panic 导致服务崩溃）
	r.Use(gin.Recovery())

	// 添加 CORS 中间件
	r.Use(corsMiddleware())

	// 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// ====== 初始化 K8s Client（本地无集群时优雅降级） ======
	var operator *k8s.Operator
	k8sClient, k8sErr := k8s.NewClient(cfg, log)
	if k8sErr != nil {
		log.Warn(fmt.Sprintf("K8s 客户端初始化失败（本地开发模式，跳过 K8s 功能）: %v", k8sErr))
	} else {
		operator = k8s.NewOperator(k8sClient)
		log.Info("K8s 客户端初始化成功")
	}

	// ====== 初始化 Service 层 ======
	notifySvc := service.NewNotificationService(cfg, log)
	instSvc := service.NewInstanceService(db, cfg, log, operator, notifySvc)

	// ====== 初始化各个 Handler ======
	authHandler := handler.NewAuthHandler(db, cfg, log)
	applicationHandler := handler.NewApplicationHandler(db, cfg, log)
	instanceHandler := handler.NewInstanceHandler(db, cfg, log, instSvc)
	adminHandler := handler.NewAdminHandler(db, cfg, log, instSvc, notifySvc)

	// JWT 鉴权中间件
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	// 审计日志中间件（记录关键写操作）
	auditMiddleware := middleware.NewAuditMiddleware(db)

	// ====== 路由配置 ======
	// 健康检查（不需要认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": "1.0.0",
		})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// ====== 认证接口（不需要 JWT） ======
		auth := v1.Group("/auth")
		{
			// POST /api/v1/auth/login - 用户登录
			auth.POST("/login", authHandler.Login)
		}

		// ====== 需要 JWT 认证的接口 ======
		authenticated := v1.Group("")
		authenticated.Use(authMiddleware.Authenticate(), auditMiddleware.AutoAudit())
		{
			// GET /api/v1/auth/me - 获取当前用户信息
			authenticated.GET("/auth/me", authHandler.Me)

			// ====== 申请管理接口（员工） ======
			applications := authenticated.Group("/applications")
			{
				// POST /api/v1/applications - 提交申请
				applications.POST("", applicationHandler.Create)
				// GET /api/v1/applications - 我的申请列表
				applications.GET("", applicationHandler.List)
				// GET /api/v1/applications/:id - 申请详情
				applications.GET("/:id", applicationHandler.Get)
				// DELETE /api/v1/applications/:id - 撤销申请
				applications.DELETE("/:id", applicationHandler.Cancel)
			}

			// ====== 实例管理接口（员工） ======
			instances := authenticated.Group("/instances")
			{
				// GET /api/v1/instances - 我的实例列表
				instances.GET("", instanceHandler.List)
				// GET /api/v1/instances/:id - 实例详情
				instances.GET("/:id", instanceHandler.Get)
				// DELETE /api/v1/instances/:id - 删除实例
				instances.DELETE("/:id", instanceHandler.Delete)
			}

			// ====== 管理员接口（需要 admin 角色） ======
			admin := authenticated.Group("/admin")
			admin.Use(middleware.RequireAdmin())
			{
				// 申请审批管理
				// GET /api/v1/admin/applications - 待审批申请列表
				admin.GET("/applications", adminHandler.ListApplications)
				// POST /api/v1/admin/applications/:id/approve - 批准申请
				admin.POST("/applications/:id/approve", adminHandler.ApproveApplication)
				// POST /api/v1/admin/applications/:id/reject - 拒绝申请
				admin.POST("/applications/:id/reject", adminHandler.RejectApplication)

				// 实例管理
				// GET /api/v1/admin/instances - 所有实例列表
				admin.GET("/instances", adminHandler.ListInstances)

				// 监控汇总
				// GET /api/v1/admin/metrics/summary - 监控汇总数据
				admin.GET("/metrics/summary", adminHandler.MetricsSummary)

				// 审计日志
				// GET /api/v1/admin/audit-logs - 操作审计日志
				admin.GET("/audit-logs", adminHandler.ListAuditLogs)
			}
		}
	}

	// 404 处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "接口不存在",
		})
	})

	return r, nil
}

// corsMiddleware CORS 跨域中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
