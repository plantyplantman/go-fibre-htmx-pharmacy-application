run:
	npx tailwindcss -i ./tailwind.css -o ./public/styles.css && go run cmd/server/main.go
build:
	npx tailwindcss -i ./tailwind.css -o ./public/styles.css && go build -o bin/server cmd/server/main.go
dev:
	npx tailwindcss -i ./tailwind.css -o ./public/styles.css && go build -o tmp/server.exe cmd/server/main.go