package migrations

import (
	"context"
	"database/sql"
)

func upAddBalanceToUsers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE users
		ADD COLUMN balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
		ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
	`)
	return err
}

func downAddBalanceToUsers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE users
		DROP COLUMN balance,
		DROP COLUMN is_deleted;
	`)
	return err
}
