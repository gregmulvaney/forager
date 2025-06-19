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

type Db struct {
	conn *sql.DB
	q    *queries.Queries
}

func Init(logger *zap.Logger) *Db {

	dbConn, err := sql.Open("sqlite3", "forager.sqlite3")
	if err != nil {
		logger.Panic("Unable to open connection to database", zap.Error(err))
	}

	if _, err := dbConn.ExecContext(context.Background(), sqlc.DDL); err != nil {
		logger.Panic("Failed to create schemas", zap.Error(err))
	}

	q := queries.New(dbConn)

	if err != nil {
		panic(err)
	}

	return &Db{
		conn: dbConn,
		q:    q,
	}
}
