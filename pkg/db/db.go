package db

import (
	"context"
	"database/sql"

	"github.com/gregmulvaney/forager/pkg/db/queries"
	"github.com/gregmulvaney/forager/sqlc"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// DB represents the database connection and query interface.
// It contains both the raw SQL connection and sqlc-generated query methods.
type DB struct {
	Conn *sql.DB          // Raw database connection for direct SQL operations
	Q    *queries.Queries // Generated query methods from sqlc
}

// Init initializes a new database connection and applies schema migrations.
// It opens a SQLite database, creates the queries interface, and runs DDL migrations.
// Returns a configured DB instance or panics on failure.
func Init(logger *zap.Logger) *DB {
	dbConn, err := sql.Open("sqlite3", "forager.sqlite3")
	if err != nil {
		logger.Panic("Failed to open sqlite database", zap.Error(err))
	}

	q := queries.New(dbConn)

	if _, err := dbConn.ExecContext(context.Background(), sqlc.DDL); err != nil {
		logger.Panic("Error migrating schema DDL", zap.Error(err))
	}

	return &DB{
		Conn: dbConn,
		Q:    q,
	}
}
