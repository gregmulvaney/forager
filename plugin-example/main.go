package main

import (
	"context"
	"database/sql"
	"fmt"
	"plugin-example/sqlc"
)

type service struct{}

func (s *service) Register(db *sql.DB) {
	if _, err := db.ExecContext(context.Background(), sqlc.DDL); err != nil {
		fmt.Printf("Errors and shit %s", err)
	}
}

var Service service
var ServiceName = "Example"
