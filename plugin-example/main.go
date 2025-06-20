package main

import (
	"context"
	"database/sql"
	_ "embed"
	"plugin-example/queries"

	"go.uber.org/zap"
)

type service struct {
	db *sql.DB
}

// go:embed queries.sql
var DDL string

func (s *service) Register(db *sql.DB, logger *zap.Logger) {
	if _, err := db.ExecContext(context.Background(), DDL); err != nil {
		logger.Error("Failed to migrate schema", zap.Error(err))
	}

	_ = queries.New(db)
}

var Service service
var ServiceName = "Example"
