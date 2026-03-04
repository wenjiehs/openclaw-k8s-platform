package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/openclaw/openclaw-saas-platform/internal/api/middleware"
	"github.com/openclaw/openclaw-saas-platform/internal/model"
	"github.com/openclaw/openclaw-saas-platform/pkg/config"
	"github.com/openclaw/openclaw-saas-platform/pkg/logger"
)

// ApplicationHandler 申请管理 HTTP Handler
type ApplicationHandler struct {
	db  *gorm.DB
	cfg *config.Config
	log *logger.Logger
}

// NewApplicationHandler 创建申请管理 Handler
func NewApplicationHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *ApplicationHandler {
	return &ApplicationHandler{db: db, cfg: cfg, log: log}
}

// CreateApplicationRequest 创建申请请求体
type CreateApplicationRequest struct {
	InstanceName string `json:"instance_name" binding:"required,min=3,max=50"`
	Spec         string `json:"spec" binding:"required,oneof=basic standard enterprise"`
	DurationType string `json:"duration_type" binding:"required,oneof=long temporary"`
	DurationDays int    `json:"duration_days"` // 仅 temporary 类型需要
	Reason       string `json:"reason" binding:"required,min=10"`
}

// Create 提交申请
// POST /api/v1/applications
func (h *ApplicationHandler) Create(c *gin.Context) {
	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误：" + err.Error(),
		})
		return
	}

	// 临时实例必须填写使用天数
	if req.DurationType == "temporary" && req.DurationDays <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "临时实例必须填写使用天数",
		})
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	// 检查是否有相同实例名的待审批申请
	var count int64
	h.db.Model(&model.Application{}).
		Where("instance_name = ? AND status = 'pending'", req.InstanceName).
		Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": "该实例名称已有待审批的申请",
		})
		return
	}

	// 创建申请记录
	application := &model.Application{
		UserID:       userID,
		InstanceName: req.InstanceName,
		Spec:         req.Spec,
		DurationType: req.DurationType,
		DurationDays: req.DurationDays,
		Reason:       req.Reason,
		Status:       "pending",
	}

	if err := h.db.Create(application).Error; err != nil {
		h.log.Error("创建申请失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建申请失败",
		})
		return
	}

	// 设置审计上下文（后续由审计中间件统一落库）
	middleware.SetAuditContext(c, middleware.AuditContext{
		Action:       "create",
		ResourceType: "application",
		ResourceID:   strconv.FormatUint(uint64(application.ID), 10),
		Extra: map[string]interface{}{
			"instance_name": application.InstanceName,
			"spec":          application.Spec,
			"duration_type": application.DurationType,
		},
	})

	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "申请提交成功",
		"data":    application,
	})
}

// List 我的申请列表
// GET /api/v1/applications?page=1&size=10&status=pending
func (h *ApplicationHandler) List(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := h.db.Model(&model.Application{}).
		Preload("Approver").
		Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var applications []model.Application
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&applications).Error; err != nil {
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

// Get 申请详情
// GET /api/v1/applications/:id
func (h *ApplicationHandler) Get(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的申请 ID",
		})
		return
	}

	var application model.Application
	if err := h.db.Preload("User").Preload("Approver").
		Where("id = ? AND user_id = ?", id, userID).
		First(&application).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "申请不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    application,
	})
}

// Cancel 撤销申请
// DELETE /api/v1/applications/:id
func (h *ApplicationHandler) Cancel(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的申请 ID",
		})
		return
	}

	var application model.Application
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).
		First(&application).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "申请不存在",
		})
		return
	}

	// 只有 pending 状态的申请可以撤销
	if !application.CanCancel() {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只有待审批的申请可以撤销",
		})
		return
	}

	// 设置审计上下文（后续由审计中间件统一落库）
	middleware.SetAuditContext(c, middleware.AuditContext{
		Action:       "cancel",
		ResourceType: "application",
		ResourceID:   strconv.Itoa(id),
		Extra: map[string]interface{}{
			"instance_name": application.InstanceName,
		},
	})

	if err := h.db.Model(&application).Update("status", "cancelled").Error; err != nil {
		h.log.Error("撤销申请失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "撤销失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "申请已撤销",
	})
}
