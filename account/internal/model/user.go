package model

import "time"

type User struct {
	ID         uint64
	Login      string
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
	Age        uint32
	Balance    float32
	IsDeleted  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreateUser struct {
	Login      string
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
	Age        uint32
	Balance    float32
}

type UpdateUser struct {
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
	Age        uint32
	Balance    float32
}

type BalanceOperationType string

const (
	BalanceOperationDeposit BalanceOperationType = "deposit"
	BalanceOperationCredit  BalanceOperationType = "credit"
)

type GetBalanceResponse struct {
	UserID  uint64
	Balance float32
}

type UpdateBalanceRequest struct {
	UserID uint64
	Amount float32
	Type   BalanceOperationType
}

type UpdateBalanceResponse struct {
	UserID     uint64
	OldBalance float32
	NewBalance float32
	Amount     float32
	Operation  BalanceOperationType
}
