tw:
	tailwindcss -i "./static/input.css" -o "./static/output.css" --watch

dev:
	go run main.go serve ./tmp/myapp

init:
	go run main.go init ./tmp/myapp admin admin --force

kill:
	sudo lsof -t -i:8080 | xargs kill -9
