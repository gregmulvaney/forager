NAME=example

build:
	go build -buildmode=plugin -o ./out/${NAME}.so .

dev:
	go build -buildmode=plugin -o ../tmp/plugins/${NAME}.so .

.PHONY: sqlc
sqlc:
	sqlc generate -f sqlc/sqlc.yaml
