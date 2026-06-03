package migrations

import (
	"context"
	"database/sql"
)

func upAddAccountIndex(ctx context.Context, tx *sql.Tx) error {
	query := `CREATE INDEX ix_account_login
        ON users (login);`
	_, err := tx.Exec(query)
	return err
}

func downAddAccountIndex(ctx context.Context, tx *sql.Tx) error {
	query := `DROP INDEX ix_account_login ON users;`
	_, err := tx.Exec(query)
	return err
}
