package migrations

import (
	"context"
	"database/sql"

	"github.com/openfga/openfga/pkg/storage/migrate"
)

func up002(ctx context.Context, tx *sql.Tx) error {
	stmt := `ALTER TABLE authorization_model ADD schema_version NVARCHAR(5) CONSTRAINT DF_schema_version DEFAULT '1.0' NOT NULL;`
	_, err := tx.ExecContext(ctx, stmt)
	return err
}

func down002(ctx context.Context, tx *sql.Tx) error {
	stmts := []string{
		`ALTER TABLE authorization_model DROP CONSTRAINT DF_schema_version;`,
		`ALTER TABLE authorization_model DROP COLUMN schema_version;`,
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
			Version:  2,
			Forward:  up002,
			Backward: down002,
		},
	)
}
