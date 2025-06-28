run:
	go run ./cmd/forager/main.go

build:
	go build -o ./bin/forager ./cmd/forager/main.go

tailwind:
	tailwindcss -o ./web/static/style.css -i ./tailwind.css --minify

.PHONY: sqlc
sqlc:
	sqlc generate -f ./sqlc/sqlc.yaml

.PHONY: templ
templ:
	templ generate
