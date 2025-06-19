package db

import (
	"context"
	"database/sql"

	_ "embed"

	"github.com/gregmulvaney/forager/pkg/db/queries"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// go:embed ../../sqlc/queries.sql
var ddl string

type Db struct {
	conn *sql.DB
	q    *queries.Queries
}

func Init(logger *zap.Logger) *Db {

	dbConn, err := sql.Open("sqlite3", "forager.sqlite3")
	if err != nil {
		logger.Panic("Unable to open connection to database", zap.Error(err))
	}

	if _, err := dbConn.ExecContext(context.Background(), ddl); err != nil {
		logger.Panic("Failed to create schemas", zap.Error(err))
	}

	queries := queries.New(dbConn)

	return &Db{
		conn: dbConn,
		q:    queries,
	}
}
