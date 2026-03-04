package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/openclaw/openclaw-saas-platform/internal/model"
	"github.com/openclaw/openclaw-saas-platform/pkg/config"
	"github.com/openclaw/openclaw-saas-platform/pkg/logger"
)

// ApplicationService 申请管理业务逻辑服务
type ApplicationService struct {
	db  *gorm.DB
	cfg *config.Config
	log *logger.Logger
}

// NewApplicationService 创建申请服务
func NewApplicationService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *ApplicationService {
	return &ApplicationService{db: db, cfg: cfg, log: log}
}

// CreateApplication 创建申请
func (s *ApplicationService) CreateApplication(app *model.Application) error {
	// 检查是否有相同实例名的待审批申请
	var count int64
	s.db.Model(&model.Application{}).
		Where("instance_name = ? AND status = 'pending'", app.InstanceName).
		Count(&count)

	if count > 0 {
		return fmt.Errorf("该实例名称已有待审批的申请")
	}

	// 检查实例名是否已存在
	var instanceCount int64
	s.db.Model(&model.Instance{}).
		Where("name = ? AND deleted_at IS NULL", app.InstanceName).
		Count(&instanceCount)

	if instanceCount > 0 {
		return fmt.Errorf("实例名称 %s 已被使用", app.InstanceName)
	}

	return s.db.Create(app).Error
}

// ApproveApplication 审批通过申请
// adminID: 审批管理员的 ID
// note: 审批备注
func (s *ApplicationService) ApproveApplication(applicationID uint, adminID uint, note string) (*model.Application, error) {
	var application model.Application
	if err := s.db.Preload("User").First(&application, applicationID).Error; err != nil {
		return nil, fmt.Errorf("申请不存在")
	}

	if !application.IsPending() {
		return nil, fmt.Errorf("该申请已处理")
	}

	// 开启事务：更新申请状态 + 创建实例记录
	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// 更新申请状态
		updates := map[string]interface{}{
			"status":       "approved",
			"approver_id":  adminID,
			"approve_note": note,
			"approved_at":  now,
		}
		if err := tx.Model(&application).Updates(updates).Error; err != nil {
			return err
		}

		// 创建实例记录
		namespace := "openclaw-" + application.InstanceName
		instance := &model.Instance{
			Name:         application.InstanceName,
			UserID:       application.UserID,
			Spec:         application.Spec,
			Status:       "creating",
			Namespace:    namespace,
			DurationType: application.DurationType,
		}

		// 设置临时实例到期时间
		if application.DurationType == "temporary" && application.DurationDays > 0 {
			expireAt := now.Add(time.Duration(application.DurationDays) * 24 * time.Hour)
			instance.ExpireAt = &expireAt
		}

		return tx.Create(instance).Error
	})

	if err != nil {
		return nil, fmt.Errorf("审批操作失败: %w", err)
	}

	return &application, nil
}

// RejectApplication 审批拒绝申请
func (s *ApplicationService) RejectApplication(applicationID uint, adminID uint, reason string) error {
	var application model.Application
	if err := s.db.First(&application, applicationID).Error; err != nil {
		return fmt.Errorf("申请不存在")
	}

	if !application.IsPending() {
		return fmt.Errorf("该申请已处理")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       "rejected",
		"approver_id":  adminID,
		"approve_note": reason,
		"approved_at":  now,
	}

	return s.db.Model(&application).Updates(updates).Error
}
