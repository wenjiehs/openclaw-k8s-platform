package model

import (
	"time"
)

// Application 申请记录表
// 对应数据库 applications 表
// 员工申请 OpenClaw 实例的流程记录
type Application struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	User         *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	InstanceName string     `gorm:"type:varchar(50);not null" json:"instance_name"` // 申请的实例名称
	Spec         string     `gorm:"type:varchar(20);not null" json:"spec"`          // 申请的规格
	// 使用时长类型：long（长期）/ temporary（临时）
	DurationType string     `gorm:"type:varchar(20);default:'long'" json:"duration_type"`
	// 临时实例使用天数（仅 temporary 类型有效）
	DurationDays int        `gorm:"default:0" json:"duration_days"`
	Reason       string     `gorm:"type:text" json:"reason"`                        // 申请理由
	// 状态：pending（待审批）/ approved（已批准）/ rejected（已拒绝）/ cancelled（已撤销）
	Status       string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ApproverID   *uint      `gorm:"index" json:"approver_id,omitempty"`
	Approver     *User      `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	ApproveNote  string     `gorm:"type:text" json:"approve_note"`                  // 审批备注
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Application) TableName() string {
	return "applications"
}

// IsPending 检查申请是否待审批
func (a *Application) IsPending() bool {
	return a.Status == "pending"
}

// IsApproved 检查申请是否已批准
func (a *Application) IsApproved() bool {
	return a.Status == "approved"
}

// CanCancel 检查申请是否可以撤销（只有 pending 状态可以撤销）
func (a *Application) CanCancel() bool {
	return a.Status == "pending"
}
