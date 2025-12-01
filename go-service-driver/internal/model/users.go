package model

import "time"

type User struct {
	ID           uint64     `json:"id" gorm:"primaryKey;autoIncrement;comment:用户ID"`
	Username     string     `json:"username" gorm:"unique;not null;comment:用户名"`
	Email        string     `json:"email" gorm:"unique;not null;comment:邮箱"`
	PasswordHash string     `json:"-" gorm:"not null;comment:加密后的密码"`
	DisplayName  string     `json:"display_name,omitempty" gorm:"comment:显示名称"`
	AvatarURL    string     `json:"avatar_url,omitempty" gorm:"comment:头像URL"`
	CreatedAt    time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
	LastLoginAt  *time.Time `json:"last_login_at" gorm:"comment:最后登录时间"`
	Status       string     `json:"status" gorm:"default:'active';comment:用户状态"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"comment:软删除时间"`
}

func (User) TableName() string {
	return "users"
}
