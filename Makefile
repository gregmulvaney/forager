run:
	go run ./cmd/forager/main.go

build:
	go build -o ./bin/forager ./cmd/forager/main.go

.PHONY: sqlc
sqlc:
	sqlc generate -f ./sqlc/sqlc.yaml

.PHONY: templ
templ:
	templ generate 

.PHONY: tailwind
tailwind:
	tailwindcss -i ./tailwind.css -o ./web/static/style.css --minify
