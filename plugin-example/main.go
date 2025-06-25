package main

import (
	"github.com/gregmulvaney/forager/pkg/plugins"
	"plugin-example/sqlc"
)

type service struct{}

var serverAPI plugins.ServerApiInterface

func Register(api plugins.ServerApiInterface) {
	serverAPI = api
	serverAPI.MigrateSchema(sqlc.DDL)
	serverAPI.RegisterRoute("/example", "Example", "<div>Example</div>")
}

var Service service
var ServiceName = "Example"
var ServiceDefaultPath = "/example"
