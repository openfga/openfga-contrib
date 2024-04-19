package migrations

import (
	"context"
	"database/sql"

	"github.com/openfga/openfga/pkg/storage/migrate"
)

func up005(ctx context.Context, tx *sql.Tx) error {
	stmts := []string{
		`ALTER TABLE tuple ADD condition_name NVARCHAR(256);`,
		`ALTER TABLE tuple ADD condition_context VARBINARY(MAX);`,
		`ALTER TABLE changelog ADD condition_name NVARCHAR(256);`,
		`ALTER TABLE changelog ADD condition_context VARBINARY(MAX);`,
	}

	for _, stmt := range stmts {
		_, err := tx.ExecContext(ctx, stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func down005(ctx context.Context, tx *sql.Tx) error {
	stmts := []string{
		`ALTER TABLE tuple DROP COLUMN condition_name;`,
		`ALTER TABLE tuple DROP COLUMN condition_context;`,
		`ALTER TABLE changelog DROP COLUMN condition_name;`,
		`ALTER TABLE changelog DROP COLUMN condition_context;`,
	}

	for _, stmt := range stmts {
		_, err := tx.ExecContext(ctx, stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	Migrations.MustRegister(
		&migrate.Migration{
			Version:  5,
			Forward:  up005,
			Backward: down005,
		},
	)
}
