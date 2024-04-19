package migrations

import (
	"context"
	"database/sql"

	"github.com/openfga/openfga/pkg/storage/migrate"
)

func up001(ctx context.Context, tx *sql.Tx) error {
	stmts := []string{
		`CREATE TABLE tuple (
			store NVARCHAR(26) NOT NULL,
			object_type NVARCHAR(128) NOT NULL,
			object_id NVARCHAR(128) NOT NULL,
			relation NVARCHAR(50) NOT NULL,
			_user NVARCHAR(256) NOT NULL,
			user_type NVARCHAR(7) NOT NULL,
			ulid NVARCHAR(26) NOT NULL,
			inserted_at DATETIME2 NOT NULL,
			PRIMARY KEY (store, object_type, object_id, relation, _user)
		);`,
		`CREATE UNIQUE INDEX idx_tuple_ulid ON tuple (ulid);`,
		`CREATE TABLE authorization_model (
			store NVARCHAR(26) NOT NULL,
			authorization_model_id NVARCHAR(26) NOT NULL,
			type NVARCHAR(256) NOT NULL,
			type_definition VARBINARY(MAX),
			PRIMARY KEY (store, authorization_model_id, type)
		);`,
		`CREATE TABLE store (
			id NVARCHAR(26) PRIMARY KEY,
			name NVARCHAR(64) NOT NULL,
			created_at DATETIME2 NOT NULL,
			updated_at DATETIME2,
			deleted_at DATETIME2
		);`,
		`CREATE TABLE assertion (
			store NVARCHAR(26) NOT NULL,
			authorization_model_id NVARCHAR(26) NOT NULL,
			assertions VARBINARY(MAX),
			PRIMARY KEY (store, authorization_model_id)
		);`,
		`CREATE TABLE changelog (
			store NVARCHAR(26) NOT NULL,
			object_type NVARCHAR(256) NOT NULL,
			object_id NVARCHAR(256) NOT NULL,
			relation NVARCHAR(50) NOT NULL,
			_user NVARCHAR(512) NOT NULL,
			operation BIGINT NOT NULL,
			ulid NVARCHAR(26) NOT NULL,
			inserted_at DATETIME2 NOT NULL,
			PRIMARY KEY (store, ulid, object_type)
		);`,
	}

	for _, stmt := range stmts {
		_, err := tx.ExecContext(ctx, stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func down001(ctx context.Context, tx *sql.Tx) error {
	stmts := []string{
		`DROP TABLE tuple;`,
		`DROP TABLE authorization_model;`,
		`DROP TABLE store;`,
		`DROP TABLE assertion;`,
		`DROP TABLE changelog;`,
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
			Version:  1,
			Forward:  up001,
			Backward: down001,
		},
	)
}
