.PHONY: up down migrate seed run tidy

up:
	docker compose up -d postgres

down:
	docker compose down

migrate:
	go run ./cmd/library

seed:
	go run ./cmd/seed -mode=reset -authors=1000 -books=10000 -readers=5000 -loans=20000

run:
	go run ./cmd/library

tidy:
	go mod tidy
