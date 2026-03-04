package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/openclaw/openclaw-saas-platform/internal/api/middleware"
	"github.com/openclaw/openclaw-saas-platform/internal/model"
	"github.com/openclaw/openclaw-saas-platform/internal/service"
	"github.com/openclaw/openclaw-saas-platform/pkg/config"
	"github.com/openclaw/openclaw-saas-platform/pkg/logger"
)

// AdminHandler 管理员操作 HTTP Handler
type AdminHandler struct {
	db        *gorm.DB
	cfg       *config.Config
	log       *logger.Logger
	instSvc   *service.InstanceService
	notifySvc *service.NotificationService
}

// NewAdminHandler 创建管理员 Handler
func NewAdminHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger, instSvc *service.InstanceService, notifySvc *service.NotificationService) *AdminHandler {
	return &AdminHandler{
		db:        db,
		cfg:       cfg,
		log:       log,
		instSvc:   instSvc,
		notifySvc: notifySvc,
	}
}

// ApproveRequest 审批请求体
type ApproveRequest struct {
	Note string `json:"note"` // 审批备注（可选）
}

// RejectRequest 拒绝请求体
type RejectRequest struct {
	Note string `json:"note" binding:"required"` // 拒绝原因（必填）
}

// ListApplications 待审批申请列表
// GET /api/v1/admin/applications?status=pending&page=1&size=10
func (h *AdminHandler) ListApplications(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status := c.DefaultQuery("status", "pending") // 默认只看待审批

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := h.db.Model(&model.Application{}).
		Preload("User").
		Preload("Approver")

	if status != "all" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var applications []model.Application
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("created_at ASC").Find(&applications).Error; err != nil {
		h.log.Error("查询申请列表失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  applications,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// ApproveApplication 批准申请
// POST /api/v1/admin/applications/:id/approve
func (h *AdminHandler) ApproveApplication(c *gin.Context) {
	adminID, _ := middleware.GetCurrentUserID(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的申请 ID",
		})
		return
	}

	var req ApproveRequest
	c.ShouldBindJSON(&req) // 备注是可选的，不强制绑定

	// 查找申请（预加载 User，后续创建实例需要用户信息）
	var application model.Application
	if err := h.db.Preload("User").First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "申请不存在",
		})
		return
	}

	// 只能审批 pending 状态的申请
	if !application.IsPending() {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该申请已处理，无法重复审批",
		})
		return
	}

	// 事务提交后需要用到的实例数据（在事务外声明，事务内赋值）
	var createdInstance *model.Instance

	// 开启事务：更新申请状态 + 创建实例记录
	err = h.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// 1. 更新申请状态为 approved
		updates := map[string]interface{}{
			"status":       "approved",
			"approver_id":  adminID,
			"approve_note": req.Note,
			"approved_at":  now,
		}
		if err := tx.Model(&application).Updates(updates).Error; err != nil {
			return err
		}

		// 2. 创建实例记录（状态为 creating，等待 K8s 资源创建完成）
		namespace := "openclaw-" + application.InstanceName
		instance := &model.Instance{
			Name:         application.InstanceName,
			UserID:       application.UserID,
			Spec:         application.Spec,
			Status:       "creating",
			Namespace:    namespace,
			DurationType: application.DurationType,
		}

		// 如果是临时实例，设置到期时间
		if application.DurationType == "temporary" && application.DurationDays > 0 {
			expireAt := now.Add(time.Duration(application.DurationDays) * 24 * time.Hour)
			instance.ExpireAt = &expireAt
		}

		if err := tx.Create(instance).Error; err != nil {
			return err
		}

		// 将实例数据传出事务（事务成功后触发异步创建）
		createdInstance = instance
		return nil
	})

	if err != nil {
		h.log.Error("审批申请失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "审批操作失败",
		})
		return
	}

	// 事务提交成功后，异步触发 K8s 资源创建
	// 注意：必须在事务外启动，否则 goroutine 里查不到刚提交的数据
	if h.instSvc != nil && createdInstance != nil && application.User != nil {
		h.instSvc.CreateInstance(createdInstance, application.User)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "申请已批准，实例创建中（约 5 分钟）",
	})
}

// RejectApplication 拒绝申请
// POST /api/v1/admin/applications/:id/reject
func (h *AdminHandler) RejectApplication(c *gin.Context) {
	adminID, _ := middleware.GetCurrentUserID(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的申请 ID",
		})
		return
	}

	var req RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请填写拒绝原因",
		})
		return
	}

	// 查找申请（预加载 User，用于发送通知）
	var application model.Application
	if err := h.db.Preload("User").First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "申请不存在",
		})
		return
	}

	if !application.IsPending() {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该申请已处理，无法重复审批",
		})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       "rejected",
		"approver_id":  adminID,
		"approve_note": req.Note,
		"approved_at":  now,
	}

	if err := h.db.Model(&application).Updates(updates).Error; err != nil {
		h.log.Error("拒绝申请失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "拒绝操作失败",
		})
		return
	}

	// 异步发送企业微信拒绝通知给申请人
	if h.notifySvc != nil && application.User != nil {
		notifySvc := h.notifySvc
		username := application.User.Username
		instanceName := application.InstanceName
		note := req.Note
		go func() {
			if notifyErr := notifySvc.SendRejectNotification(username, instanceName, note); notifyErr != nil {
				h.log.Warn("发送拒绝通知失败: " + notifyErr.Error())
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "申请已拒绝",
	})
}

// ListInstances 管理员查看所有实例
// GET /api/v1/admin/instances?page=1&size=10&status=running
func (h *AdminHandler) ListInstances(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status := c.Query("status")
	department := c.Query("department")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := h.db.Model(&model.Instance{}).
		Preload("User").
		Where("deleted_at IS NULL")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 按部门过滤（需要 JOIN users 表）
	if department != "" {
		query = query.Joins("JOIN users ON users.id = instances.user_id").
			Where("users.department = ?", department)
	}

	var total int64
	query.Count(&total)

	var instances []model.Instance
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("instances.created_at DESC").Find(&instances).Error; err != nil {
		h.log.Error("查询实例列表失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  instances,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// MetricsSummary 监控汇总数据
// GET /api/v1/admin/metrics/summary
func (h *AdminHandler) MetricsSummary(c *gin.Context) {
	// 统计实例数量
	var totalInstances, runningInstances, creatingInstances, failedInstances int64

	h.db.Model(&model.Instance{}).Where("deleted_at IS NULL").Count(&totalInstances)
	h.db.Model(&model.Instance{}).Where("deleted_at IS NULL AND status = 'running'").Count(&runningInstances)
	h.db.Model(&model.Instance{}).Where("deleted_at IS NULL AND status = 'creating'").Count(&creatingInstances)
	h.db.Model(&model.Instance{}).Where("deleted_at IS NULL AND status = 'failed'").Count(&failedInstances)

	// 统计申请数量
	var pendingApplications, totalApplicationsToday int64
	h.db.Model(&model.Application{}).Where("status = 'pending'").Count(&pendingApplications)
	h.db.Model(&model.Application{}).
		Where("DATE(created_at) = CURRENT_DATE").
		Count(&totalApplicationsToday)

	// 统计用户数量
	var totalUsers int64
	h.db.Model(&model.User{}).Count(&totalUsers)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"instances": gin.H{
				"total":    totalInstances,
				"running":  runningInstances,
				"creating": creatingInstances,
				"failed":   failedInstances,
			},
			"applications": gin.H{
				"pending": pendingApplications,
				"today":   totalApplicationsToday,
			},
			"users": gin.H{
				"total": totalUsers,
			},
		},
	})
}

// ListAuditLogs 审计日志列表
// GET /api/v1/admin/audit-logs?page=1&size=20&action=approve&result=success
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	query := h.db.Model(&model.AuditLog{}).Preload("User")

	// 按操作类型过滤
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}
	// 按结果过滤（success/failed）
	if result := c.Query("result"); result != "" {
		query = query.Where("result = ?", result)
	}
	// 按用户 ID 过滤
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			query = query.Where("user_id = ?", userID)
		}
	}
	// 按时间范围过滤
	if startTime := c.Query("start_time"); startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime := c.Query("end_time"); endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	var total int64
	query.Count(&total)

	var logs []model.AuditLog
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&logs).Error; err != nil {
		h.log.Error("查询审计日志失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  logs,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}
