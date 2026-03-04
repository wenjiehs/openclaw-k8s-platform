package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/openclaw/openclaw-saas-platform/internal/k8s"
	"github.com/openclaw/openclaw-saas-platform/internal/model"
	"github.com/openclaw/openclaw-saas-platform/pkg/config"
	"github.com/openclaw/openclaw-saas-platform/pkg/logger"
)

// InstanceService 实例管理业务逻辑服务
// 负责 OpenClaw 实例的创建、删除、状态同步等操作
type InstanceService struct {
	db        *gorm.DB
	cfg       *config.Config
	log       *logger.Logger
	operator  *k8s.Operator
	notifySvc *NotificationService
}

// NewInstanceService 创建实例服务
func NewInstanceService(db *gorm.DB, cfg *config.Config, log *logger.Logger, operator *k8s.Operator, notifySvc *NotificationService) *InstanceService {
	return &InstanceService{
		db:        db,
		cfg:       cfg,
		log:       log,
		operator:  operator,
		notifySvc: notifySvc,
	}
}

// CreateInstance 异步创建 OpenClaw 实例
// 在审批通过后调用，异步执行 K8s 资源创建
func (s *InstanceService) CreateInstance(instance *model.Instance, user *model.User) {
	// K8s Operator 未初始化时（本地开发），跳过创建
	if s.operator == nil {
		s.log.Warn(fmt.Sprintf("K8s Operator 未配置，跳过实例创建: %s", instance.Name))
		s.db.Model(instance).Update("status", "failed")
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		s.log.Info(fmt.Sprintf("开始异步创建实例: %s", instance.Name))

		// 配置部署参数
		cfg := k8s.InstanceDeployConfig{
			Username:   user.Username,
			Namespace:  instance.Namespace,
			Spec:       instance.Spec,
			Department: user.Department,
			Domain:     s.cfg.TKE.IngressDomain,
		}

		// 调用 K8s Operator 创建资源
		accessURL, err := s.operator.CreateInstance(ctx, cfg)
		if err != nil {
			s.log.Error(fmt.Sprintf("创建实例失败: %s, 错误: %v", instance.Name, err))
			// 更新实例状态为 failed
			s.db.Model(instance).Update("status", "failed")
			// 异步发送失败通知
			if s.notifySvc != nil {
				go func() {
					if notifyErr := s.notifySvc.SendInstanceFailedNotification(user.Username, instance.Name, err.Error()); notifyErr != nil {
						s.log.Warn(fmt.Sprintf("发送失败通知异常: %v", notifyErr))
					}
				}()
			}
			return
		}

		// 等待实例就绪
		if err := s.operator.WaitForInstanceReady(ctx, instance.Namespace, 10*time.Minute); err != nil {
			s.log.Error(fmt.Sprintf("等待实例就绪超时: %s, 错误: %v", instance.Name, err))
			s.db.Model(instance).Update("status", "failed")
			if s.notifySvc != nil {
				go func() {
					if notifyErr := s.notifySvc.SendInstanceFailedNotification(user.Username, instance.Name, "等待就绪超时"); notifyErr != nil {
						s.log.Warn(fmt.Sprintf("发送超时通知异常: %v", notifyErr))
					}
				}()
			}
			return
		}

		// 更新实例状态为 running，并保存访问地址
		updates := map[string]interface{}{
			"status":      "running",
			"ingress_url": accessURL,
		}
		if err := s.db.Model(instance).Updates(updates).Error; err != nil {
			s.log.Error(fmt.Sprintf("更新实例状态失败: %v", err))
		}

		s.log.Info(fmt.Sprintf("实例创建成功: %s, 访问地址: %s", instance.Name, accessURL))

		// 异步发送成功通知
		if s.notifySvc != nil {
			go func() {
				if notifyErr := s.notifySvc.SendInstanceReadyNotification(user.Username, instance.Name, accessURL); notifyErr != nil {
					s.log.Warn(fmt.Sprintf("发送就绪通知异常: %v", notifyErr))
				}
			}()
		}
	}()
}

// DeleteInstance 删除实例（包括 K8s 资源）
func (s *InstanceService) DeleteInstance(instance *model.Instance) error {
	// K8s Operator 未初始化时，仅做软删除
	if s.operator != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := s.operator.DeleteInstance(ctx, instance.Namespace); err != nil {
			return fmt.Errorf("删除 K8s 资源失败: %w", err)
		}
	}

	// 软删除数据库记录
	now := time.Now()
	updates := map[string]interface{}{
		"deleted_at": now,
		"status":     "deleted",
	}
	return s.db.Model(instance).Updates(updates).Error
}

// SyncInstanceStatus 同步实例状态（从 K8s 更新到数据库）
// 定期调用，保持数据库状态与 K8s 实际状态一致
func (s *InstanceService) SyncInstanceStatus(instance *model.Instance) error {
	if s.operator == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	k8sStatus, err := s.operator.GetInstanceMetrics(ctx, instance.Namespace)
	if err != nil {
		return fmt.Errorf("获取实例状态失败: %w", err)
	}

	// 映射 K8s Pod 状态到业务状态
	var businessStatus string
	switch k8sStatus.Status {
	case "Running":
		businessStatus = "running"
	case "Pending":
		businessStatus = "creating"
	case "Failed", "Unknown":
		businessStatus = "failed"
	default:
		return nil // 状态未知，不更新
	}

	// 只在状态变化时更新
	if instance.Status != businessStatus {
		return s.db.Model(instance).Update("status", businessStatus).Error
	}

	return nil
}
