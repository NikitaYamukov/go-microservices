package mapper

import (
	"transaction/internal/model"
	repomodel "transaction/internal/repository/model"
)

func TransactionToRepoTransaction(transaction model.Transaction) repomodel.Transaction {
	return repomodel.Transaction{
		ID:        transaction.ID,
		UserID:    transaction.UserID,
		Amount:    transaction.Amount,
		Status:    string(transaction.Status),
		Type:      string(transaction.Type),
		CreatedAt: transaction.CreatedAt,
		UpdatedAt: transaction.UpdatedAt,
	}
}

func RepoTransactionToTransaction(transaction repomodel.Transaction) model.Transaction {
	return model.Transaction{
		ID:        transaction.ID,
		UserID:    transaction.UserID,
		Amount:    transaction.Amount,
		Status:    model.TransactionStatus(transaction.Status),
		Type:      model.TransactionType(transaction.Type),
		CreatedAt: transaction.CreatedAt,
		UpdatedAt: transaction.UpdatedAt,
	}
}

func RepoTransactionsToTransactions(transactions []repomodel.Transaction) []model.Transaction {
	res := make([]model.Transaction, len(transactions))
	for i, transaction := range transactions {
		res[i] = RepoTransactionToTransaction(transaction)
	}

	return res
}

func TransactionEntryToRepoTransactionEntry(entry model.TransactionEntry) repomodel.TransactionEntry {
	return repomodel.TransactionEntry{
		ID:            entry.ID,
		TransactionID: entry.TransactionID,
		AccountID:     entry.AccountID,
		Direction:     string(entry.Direction),
		Amount:        entry.Amount,
		CreatedAt:     entry.CreatedAt,
		UpdatedAt:     entry.UpdatedAt,
	}
}

func RepoTransactionEntryToTransactionEntry(entry repomodel.TransactionEntry) model.TransactionEntry {
	return model.TransactionEntry{
		ID:            entry.ID,
		TransactionID: entry.TransactionID,
		AccountID:     entry.AccountID,
		Direction:     model.TransactionEntryDirection(entry.Direction),
		Amount:        entry.Amount,
		CreatedAt:     entry.CreatedAt,
		UpdatedAt:     entry.UpdatedAt,
	}
}

func RepoTransactionEntriesToTransactionEntries(entries []repomodel.TransactionEntry) []model.TransactionEntry {
	res := make([]model.TransactionEntry, len(entries))
	for i, entry := range entries {
		res[i] = RepoTransactionEntryToTransactionEntry(entry)
	}

	return res
}

func UpdateTransactionToRepoTransaction(transaction model.UpdateTransaction) repomodel.Transaction {
	return repomodel.Transaction{
		Status: string(transaction.Status),
	}
}
