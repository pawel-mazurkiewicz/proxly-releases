.PHONY: build run test clean migrate-up migrate-down docker-up docker-down

# Build the application
build:
	go build -o bin/proxly-license-server cmd/api/main.go

# Run the application
run:
	go run cmd/api/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run database migrations up
migrate-up:
	migrate -path internal/store/migrations -database "${DB_DSN}" up

# Run database migrations down
migrate-down:
	migrate -path internal/store/migrations -database "${DB_DSN}" down

# Start docker compose
docker-up:
	docker-compose -f docker/docker-compose.yml up -d

# Stop docker compose
docker-down:
	docker-compose -f docker/docker-compose.yml down

docker-build:
	docker-compose -f docker/docker-compose.yml build

# Install dependencies
deps:
	go mod tidy
	go mod download

# Generate seed data
seed:
	./scripts/dev_seed.sh
