package model

import (
	"time"
)

// AuditLog 审计日志表
// 对应数据库 audit_logs 表
// 记录所有关键操作，满足合规要求，保留 180 天
type AuditLog struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint      `gorm:"index" json:"user_id"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	// 操作类型：login/logout/create/delete/approve/reject/view 等
	Action       string    `gorm:"type:varchar(50);not null;index" json:"action"`
	// 资源类型：instance/application/user 等
	ResourceType string    `gorm:"type:varchar(50);not null;index" json:"resource_type"`
	// 资源 ID（字符串格式，兼容不同类型的 ID）
	ResourceID   string    `gorm:"type:varchar(100)" json:"resource_id"`
	IP           string    `gorm:"type:varchar(50)" json:"ip"`
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	// 操作结果：success（成功）/ failed（失败）
	Result       string    `gorm:"type:varchar(20)" json:"result"`
	// 额外信息（JSON 格式，存储操作的详细参数）
	Extra        string    `gorm:"type:text" json:"extra,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}
