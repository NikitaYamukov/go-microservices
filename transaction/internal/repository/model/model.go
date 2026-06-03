package model

import "time"

type Transaction struct {
	ID        uint64    `gorm:"column:id"`
	UserID    uint64    `gorm:"column:user_id"`
	Amount    int64     `gorm:"column:amount"`
	Status    string    `gorm:"column:status"`
	Type      string    `gorm:"column:type"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Transaction) TableName() string {
	return "transactions"
}

type TransactionEntry struct {
	ID            uint64    `gorm:"column:id"`
	TransactionID uint64    `gorm:"column:transaction_id"`
	AccountID     uint64    `gorm:"column:account_id"`
	Direction     string    `gorm:"column:direction"`
	Amount        float64   `gorm:"column:amount"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (TransactionEntry) TableName() string {
	return "transaction_entries"
}
