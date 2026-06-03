package model

import "time"

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypeTransfer TransactionType = "transfer"
)

type Transaction struct {
	ID        uint64
	UserID    uint64
	Amount    int64 // сумма в копейках
	Status    TransactionStatus
	Type      TransactionType
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpdateTransaction struct {
	Status TransactionStatus
}

type TransactionEntryDirection string

const (
	TransactionEntryDirectionDebit  TransactionEntryDirection = "DEBIT"
	TransactionEntryDirectionCredit TransactionEntryDirection = "CREDIT"
)

type TransactionEntry struct {
	ID            uint64
	TransactionID uint64
	AccountID     uint64
	Direction     TransactionEntryDirection
	Amount        float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TransactionDetails struct {
	Transaction Transaction
	Entries     []TransactionEntry
}

type GetTransactionsParams struct {
	UserID   *uint64
	Type     *string
	Status   *string
	DateFrom *time.Time
	DateTo   *time.Time
	Limit    int
	Offset   int
}

type DepositParams struct {
	UserID uint64
	Amount float64
}

type WithdrawParams struct {
	AccountID uint64
	Amount    float64
}

type TransferParams struct {
	UserID    uint64
	Recipient uint64
	Amount    float64
}
