package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/openclaw/openclaw-saas-platform/internal/api/middleware"
	"github.com/openclaw/openclaw-saas-platform/internal/model"
	"github.com/openclaw/openclaw-saas-platform/internal/service"
	"github.com/openclaw/openclaw-saas-platform/pkg/config"
	"github.com/openclaw/openclaw-saas-platform/pkg/logger"
)

// InstanceHandler 实例管理 HTTP Handler
type InstanceHandler struct {
	db      *gorm.DB
	cfg     *config.Config
	log     *logger.Logger
	instSvc *service.InstanceService
}

// NewInstanceHandler 创建实例管理 Handler
func NewInstanceHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger, instSvc *service.InstanceService) *InstanceHandler {
	return &InstanceHandler{
		db:      db,
		cfg:     cfg,
		log:     log,
		instSvc: instSvc,
	}
}

// List 我的实例列表
// GET /api/v1/instances?page=1&size=10&status=running
func (h *InstanceHandler) List(c *gin.Context) {
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

	query := h.db.Model(&model.Instance{}).
		Preload("User").
		Where("user_id = ? AND deleted_at IS NULL", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var instances []model.Instance
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&instances).Error; err != nil {
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

// Get 实例详情
// GET /api/v1/instances/:id
func (h *InstanceHandler) Get(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	role := middleware.GetCurrentUserRole(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的实例 ID",
		})
		return
	}

	query := h.db.Preload("User").Where("id = ? AND deleted_at IS NULL", id)

	// 非管理员只能查看自己的实例
	if role != "admin" && role != "super_admin" {
		query = query.Where("user_id = ?", userID)
	}

	var instance model.Instance
	if err := query.First(&instance).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "实例不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    instance,
	})
}

// Delete 删除实例
// DELETE /api/v1/instances/:id
func (h *InstanceHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	role := middleware.GetCurrentUserRole(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的实例 ID",
		})
		return
	}

	query := h.db.Where("id = ? AND deleted_at IS NULL", id)

	// 非管理员只能删除自己的实例
	if role != "admin" && role != "super_admin" {
		query = query.Where("user_id = ?", userID)
	}

	var instance model.Instance
	if err := query.First(&instance).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "实例不存在",
		})
		return
	}

	// 设置审计上下文（后续由审计中间件统一落库）
	middleware.SetAuditContext(c, middleware.AuditContext{
		Action:       "delete",
		ResourceType: "instance",
		ResourceID:   strconv.Itoa(id),
		Extra: map[string]interface{}{
			"instance_name": instance.Name,
			"namespace":     instance.Namespace,
		},
	})

	// 若 InstanceService 可用（K8s 已配置），通过 Service 删除（含 K8s 资源清理）
	if h.instSvc != nil {
		if err := h.instSvc.DeleteInstance(&instance); err != nil {
			h.log.Error("删除实例失败: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "删除实例失败",
			})
			return
		}
	} else {
		// 未配置 K8s 时，仅做数据库软删除
		updates := map[string]interface{}{
			"deleted_at": gorm.Expr("NOW()"),
			"status":     "deleted",
		}
		if err := h.db.Model(&instance).Updates(updates).Error; err != nil {
			h.log.Error("删除实例失败: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "删除实例失败",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "实例删除中，K8s 资源将在后台清理",
	})
}
