run:
	go run cmd/forager/main.go

build-plugin:
	go build -buildmode=plugin -o tmp/plugin/example.so ./plugin-example/main.go

tailwind:
	tailwindcss -o ./web/static/style.css -i ./tailwind.css

.PHONY: sqlc
sqlc:
	sqlc generate -f ./sqlc/sqlc.yaml
