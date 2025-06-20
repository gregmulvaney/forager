package db

import (
	"context"
	"database/sql"

	_ "embed"

	"github.com/gregmulvaney/forager/pkg/db/queries"
	"github.com/gregmulvaney/forager/sqlc"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

const currentSchemaVersion = 1

type Db struct {
	Conn *sql.DB
	Q    *queries.Queries
}

func Init(logger *zap.Logger) *Db {

	dbConn, err := sql.Open("sqlite3", "forager.sqlite3")
	if err != nil {
		logger.Panic("Unable to open connection to database", zap.Error(err))
	}

	q := queries.New(dbConn)
	db := &Db{
		Conn: dbConn,
		Q:    q,
	}

	if err := db.ensureSchema(logger); err != nil {
		logger.Panic("Failed to ensure database schema", zap.Error(err))
	}

	return db
}

func (db *Db) ensureSchema(logger *zap.Logger) error {
	ctx := context.Background()

	// Check current schema version
	currentVersion, err := db.Q.GetSchemaVersion(ctx)
	if err != nil {
		// Schema version table doesn't exist, run initial schema
		logger.Info("No schema version found, initializing database")
		if _, err := db.Conn.ExecContext(ctx, sqlc.DDL); err != nil {
			return err
		}

		// Set initial schema version
		if err := db.Q.SetSchemaVersion(ctx, currentSchemaVersion); err != nil {
			return err
		}

		logger.Info("Database schema initialized", zap.Int64("version", currentSchemaVersion))
		return nil
	}

	// Check if schema needs updating
	if currentVersion < currentSchemaVersion {
		logger.Info("Updating database schema",
			zap.Int64("from_version", currentVersion),
			zap.Int64("to_version", currentSchemaVersion))

		// Run migrations here if needed
		// For now, just update the version
		if err := db.Q.SetSchemaVersion(ctx, currentSchemaVersion); err != nil {
			return err
		}

		logger.Info("Database schema updated", zap.Int64("version", currentSchemaVersion))
	} else {
		logger.Debug("Database schema is up to date", zap.Int64("version", currentVersion))
	}

	return nil
}
