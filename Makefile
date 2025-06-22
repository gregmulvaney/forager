run:
	go run ./cmd/forager/main.go

.PHONY=tailwind
tailwind:
	tailwindcss -o ./web/static/style.css -i ./tailwind.css
