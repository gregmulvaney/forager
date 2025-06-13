run:
	go run cmd/forager/main.go

tailwind:
	bunx @tailwindcss/cli -i ./tailwind.css -o web/static/index.css 

templ:
	templ generate && make tailwind
