tw:
	tailwindcss -i "./static/input.css" -o "./static/output.css" --watch

dev:
	go run main.go serve myapp