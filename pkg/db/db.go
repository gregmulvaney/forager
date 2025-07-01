package db

import (
	"context"
	"database/sql"

	"github.com/gregmulvaney/forager/pkg/db/queries"
	"github.com/gregmulvaney/forager/sqlc"
	"go.uber.org/zap"
)

type DB struct {
	Conn *sql.DB
	Q    *queries.Queries
}

func Init(logger *zap.Logger) *DB {
	dbConn, err := sql.Open("sqlite3", "forager.sqlite3")
	if err != nil {
		logger.Panic("Failed to initialize sqlite database", zap.Error(err))
	}
	defer dbConn.Close()

	q := queries.New(dbConn)

	if _, err := dbConn.ExecContext(context.Background(), sqlc.DDL); err != nil {
		logger.Panic("Failed to migrate schema", zap.Error(err))
	}

	return &DB{
		Conn: dbConn,
		Q:    q,
	}
}
