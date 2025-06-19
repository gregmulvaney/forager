NAME = "plugin"

build:
	go build -buildmode=plugin -o bin/${NAME}.so main.go

.PHONY = sqlc
sqlc:
	sqlc generate

