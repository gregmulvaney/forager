package db

import (
	"context"
	"database/sql"

	"github.com/gregmulvaney/forager/pkg/db/queries"
	"github.com/gregmulvaney/forager/sqlc"
	"go.uber.org/zap"
)

type Db struct {
	Conn *sql.DB
	Q    *queries.Queries
}

func Init(logger *zap.Logger) *Db {
	dbConn, err := sql.Open("sqlite3", "forager.sqlite3")
	if err != nil {
		logger.Panic("Failed to open sqlite database", zap.Error(err))
	}

	q := queries.New(dbConn)

	if _, err := dbConn.ExecContext(context.Background(), sqlc.DDL); err != nil {
		logger.Debug("Faile to migrate database table", zap.Error(err))
	}

	return &Db{
		Conn: dbConn,
		Q:    q,
	}
}
