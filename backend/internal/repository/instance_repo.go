package repository

import (
	"gorm.io/gorm"

	"github.com/openclaw/openclaw-saas-platform/internal/model"
)

// InstanceRepository 实例数据访问层
type InstanceRepository struct {
	db *gorm.DB
}

// NewInstanceRepository 创建实例数据访问层
func NewInstanceRepository(db *gorm.DB) *InstanceRepository {
	return &InstanceRepository{db: db}
}

// FindByID 根据 ID 查找实例
func (r *InstanceRepository) FindByID(id uint) (*model.Instance, error) {
	var instance model.Instance
	err := r.db.Preload("User").Where("id = ? AND deleted_at IS NULL", id).First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// FindByUserID 查找用户的所有实例
func (r *InstanceRepository) FindByUserID(userID uint) ([]model.Instance, error) {
	var instances []model.Instance
	err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

// FindByNamespace 根据 Namespace 查找实例
func (r *InstanceRepository) FindByNamespace(namespace string) (*model.Instance, error) {
	var instance model.Instance
	err := r.db.Where("namespace = ? AND deleted_at IS NULL", namespace).First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// Create 创建实例记录
func (r *InstanceRepository) Create(instance *model.Instance) error {
	return r.db.Create(instance).Error
}

// UpdateStatus 更新实例状态
func (r *InstanceRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Instance{}).Where("id = ?", id).
		Update("status", status).Error
}

// UpdateIngressURL 更新实例访问地址
func (r *InstanceRepository) UpdateIngressURL(id uint, url string) error {
	return r.db.Model(&model.Instance{}).Where("id = ?", id).
		Update("ingress_url", url).Error
}

// CountByStatus 统计各状态实例数量
func (r *InstanceRepository) CountByStatus() (map[string]int64, error) {
	type StatusCount struct {
		Status string
		Count  int64
	}

	var results []StatusCount
	err := r.db.Model(&model.Instance{}).
		Select("status, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}
