package model

import (
	"time"
)

// User 用户表
// 对应数据库 users 表
type User struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username   string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email      string    `gorm:"type:varchar(100);not null" json:"email"`
	Department string    `gorm:"type:varchar(100)" json:"department"`
	// 角色：employee（员工）/ admin（管理员）/ super_admin（超级管理员）
	Role       string    `gorm:"type:varchar(20);not null;default:'employee'" json:"role"`
	// 密码哈希（使用 bcrypt）
	PasswordHash string  `gorm:"type:varchar(255);not null" json:"-"` // 不序列化到 JSON
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsAdmin 检查用户是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == "admin" || u.Role == "super_admin"
}

// IsSuperAdmin 检查用户是否是超级管理员
func (u *User) IsSuperAdmin() bool {
	return u.Role == "super_admin"
}
