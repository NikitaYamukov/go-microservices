package model

import "time"

type User struct {
	ID         uint64    `gorm:"column:id"`
	Login      string    `gorm:"column:login"`
	Email      string    `gorm:"column:email"`
	Phone      string    `gorm:"column:phone"`
	FirstName  string    `gorm:"column:first_name"`
	LastName   string    `gorm:"column:last_name"`
	MiddleName string    `gorm:"column:middle_name"`
	Age        uint32    `gorm:"column:age"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
