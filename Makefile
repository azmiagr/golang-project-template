init:
	cp .env.example .env
	go mod tidy

run:
	go run cmd/app/main.go

migrate:
	go run cmd/app/main.go migrate

test:
	go test ./...