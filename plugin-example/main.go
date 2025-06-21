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

//go:embed schema.sql
var DDL string

func (s *service) Register(db *sql.DB, logger *zap.Logger) {
	logger.Debug("Attemping to wrie schema to DB", zap.String("Plugin name", ServiceName))
	if _, err := db.ExecContext(context.Background(), DDL); err != nil {
		logger.Info("Failed to migrate schema")
	}

	_ = queries.New(db)
}

var Service service
var ServiceName = "Example"
