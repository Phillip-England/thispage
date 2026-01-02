tw:
	tailwindcss -i "./static/input.css" -o "./static/output.css" --watch

dev:
	go run main.go serve myapp

init:
	go run main.go init myapp --force

kill:
	sudo lsof -t -i:8080 | xargs kill -9
