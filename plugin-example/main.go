package main

import (
	"bytes"
	"context"
	"fmt"
	"plugin-example/queries"
	"plugin-example/sqlc"
	"plugin-example/web/pages"

	"github.com/gregmulvaney/forager/pkg/plugins"
)

type service struct {
	Q *queries.Queries
}

func (s *service) Register(api plugins.ServerApiInterface) {
	dbConn, err := api.MigrateSchema(sqlc.DDL)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	err = pages.Index().Render(context.Background(), buf)
	if err != nil {
		fmt.Println("Failed to render component")
	}

	api.RegisterRoute("/example", "Example", buf.String())

	q := queries.New(dbConn)
	s.Q = q
}

var Service service
var ServiceName = "Example"
var ServiceDefaultPath = "/example"
var DomainRegex = "^https?://([a-zA-Z0-9-]+\\.)*example\\.(com|org|net)(/.*)?$"
var Version = "0.0.1"
