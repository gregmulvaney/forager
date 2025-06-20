NAME = example

build:
	go build -buildmode=plugin -o bin/${NAME}.so main.go

.PHONY = sqlc
sqlc:
	sqlc generate

dev:
	go build -buildmode=plugin -o ../tmp/plugins/${NAME}.so

