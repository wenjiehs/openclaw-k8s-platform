package repository

import (
	"gorm.io/gorm"

	"github.com/openclaw/openclaw-saas-platform/internal/model"
)

// ApplicationRepository 申请数据访问层
type ApplicationRepository struct {
	db *gorm.DB
}

// NewApplicationRepository 创建申请数据访问层
func NewApplicationRepository(db *gorm.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

// FindByID 根据 ID 查找申请
func (r *ApplicationRepository) FindByID(id uint) (*model.Application, error) {
	var application model.Application
	err := r.db.Preload("User").Preload("Approver").First(&application, id).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}

// FindByUserID 查找用户的所有申请
func (r *ApplicationRepository) FindByUserID(userID uint, status string, page, size int) ([]model.Application, int64, error) {
	var applications []model.Application
	var total int64

	query := r.db.Model(&model.Application{}).
		Preload("Approver").
		Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * size
	err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&applications).Error
	return applications, total, err
}

// FindPendingApplications 查找所有待审批申请
func (r *ApplicationRepository) FindPendingApplications(page, size int) ([]model.Application, int64, error) {
	var applications []model.Application
	var total int64

	query := r.db.Model(&model.Application{}).
		Preload("User").
		Where("status = 'pending'")

	query.Count(&total)

	offset := (page - 1) * size
	err := query.Offset(offset).Limit(size).Order("created_at ASC").Find(&applications).Error
	return applications, total, err
}

// Create 创建申请记录
func (r *ApplicationRepository) Create(application *model.Application) error {
	return r.db.Create(application).Error
}

// UpdateStatus 更新申请状态
func (r *ApplicationRepository) UpdateStatus(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.Application{}).Where("id = ?", id).Updates(updates).Error
}
