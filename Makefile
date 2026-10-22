.PHONY: build run test test-unit test-integration migrate seed docker-up docker-down

build:
	go build -o bin/api ./cmd/api
	go build -o bin/seeder ./cmd/seeder

run:
	go run ./cmd/api

test:
	go test -v ./tests/...

test-unit:
	go test -v ./tests/unit/...

test-integration:
	go test -v ./tests/integration/...

seed:
	go run ./cmd/seeder

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v

