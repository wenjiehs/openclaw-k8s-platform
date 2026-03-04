package model

import (
	"time"
)

// Instance 实例表
// 对应数据库 instances 表
// 每个 Instance 对应一个 K8s Namespace 中的 OpenClaw 富容器部署
type Instance struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"` // 实例名称，如 openclaw-zhangsan
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	User         *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	// 规格：basic（基础版）/ standard（标准版）/ enterprise（企业版）
	Spec         string     `gorm:"type:varchar(20);not null;default:'standard'" json:"spec"`
	// 状态：pending（等待创建）/ creating（创建中）/ running（运行中）/ stopped（已停止）/ failed（创建失败）/ deleted（已删除）
	Status       string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Namespace    string     `gorm:"type:varchar(100)" json:"namespace"`         // K8s Namespace，如 openclaw-zhangsan
	IngressURL   string     `gorm:"type:varchar(200)" json:"ingress_url"`       // 访问地址，如 https://openclaw-zhangsan.tke-cloud.com
	// 使用时长类型：long（长期）/ temporary（临时）
	DurationType string     `gorm:"type:varchar(20);default:'long'" json:"duration_type"`
	ExpireAt     *time.Time `gorm:"index" json:"expire_at,omitempty"`           // 临时实例到期时间
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at,omitempty"`          // 软删除时间
}

// TableName 指定表名
func (Instance) TableName() string {
	return "instances"
}

// IsRunning 检查实例是否在运行
func (i *Instance) IsRunning() bool {
	return i.Status == "running"
}

// IsDeleted 检查实例是否已被删除
func (i *Instance) IsDeleted() bool {
	return i.DeletedAt != nil
}

// GetNamespace 返回 K8s Namespace 名称
// 如果未设置，根据实例名称自动生成
func (i *Instance) GetNamespace() string {
	if i.Namespace != "" {
		return i.Namespace
	}
	return "openclaw-" + i.Name
}

// SpecResourceConfig 规格对应的资源配置
type SpecResourceConfig struct {
	CPU    string // CPU 请求，如 "1"、"2"、"4"
	Memory string // 内存请求，如 "2Gi"、"4Gi"、"8Gi"
}

// GetResourceConfig 根据规格获取资源配置
func GetResourceConfig(spec string) SpecResourceConfig {
	switch spec {
	case "basic":
		return SpecResourceConfig{CPU: "1", Memory: "2Gi"}
	case "enterprise":
		return SpecResourceConfig{CPU: "4", Memory: "8Gi"}
	default: // standard
		return SpecResourceConfig{CPU: "2", Memory: "4Gi"}
	}
}
