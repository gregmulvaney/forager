package main

type ServerAPIInterface interface {
	MigrateSchema(ddl string)
	RegisterRoute(path string, method string)
}

type service struct {
	serverAPI ServerAPIInterface
}

func (s service) Register(serverAPI any) {
	s.serverAPI = serverAPI.(ServerAPIInterface)
	s.serverAPI.RegisterRoute("path", "method")
}

var Service service
