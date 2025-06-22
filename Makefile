run:
	go run ./cmd/forager/main.go

.PHONY=tailwind
tailwind:
	tailwindcss -o ./web/static/style.css -i ./tailwind.css --minify

.PHONY=sqlc
sqlc:
	sqlc generate -f sqlc/sqlc.yaml

