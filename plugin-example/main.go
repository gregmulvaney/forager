package main

import (
	"database/sql"
	_ "embed"
)

type service struct{}

// go:embed queries.sql
var DDL string

func (s *service) Register(*sql.DB) {
	// TODO: Register routes
	// register sql

}

var Service service
