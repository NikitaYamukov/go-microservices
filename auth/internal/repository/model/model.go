package model

import "time"

type User struct {
	ID           uint64    `gorm:"column:id"`
	Login        string    `gorm:"column:login"`
	Email        string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}

type RefreshToken struct {
	ID        uint64     `gorm:"column:id"`
	UserID    uint64     `gorm:"column:user_id"`
	Token     string     `gorm:"column:token"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
