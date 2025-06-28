package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"plugin-example/sqlc"
	"plugin-example/web"

	"github.com/a-h/templ"
)

type ServerAPIInterface interface {
	MigrateSchema(ddl string) *sql.DB
	RegisterViewRoute(path string, title string, content string)
}

type service struct {
	serverAPI ServerAPIInterface
}

func (s *service) Register(serverAPI any) {
	s.serverAPI = serverAPI.(ServerAPIInterface)

	// Migrate our plugins schema to the apps database
	_ = s.serverAPI.MigrateSchema(sqlc.DDL)

	indexContent, err := templToString(web.Index())
	if err != nil {
		fmt.Println(err)
	}

	s.serverAPI.RegisterViewRoute("/example", "Example", indexContent)
}

func templToString(component templ.Component) (string, error) {
	buf := new(bytes.Buffer)
	err := component.Render(context.Background(), buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

var Service service
var ServiceName = "Example"
