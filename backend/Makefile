.PHONY: build run test lint clean migrate

# Build the server binary
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/server ./cmd/server/main.go

# Run the server locally
run:
	go run ./cmd/server/main.go

# Run all tests
test:
	go test -v -race ./...

# Run tests with coverage
test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run ./...

# Run go vet
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Run database migrations (requires psql and DATABASE_URL)
migrate:
	psql $(DATABASE_URL) -f migrations/001_initial_schema.sql

# Docker build
docker-build:
	docker build -t ledgerly-api .

# Docker run
docker-run:
	docker run -p 8080:8080 --env-file .env ledgerly-api

# Tidy dependencies
tidy:
	go mod tidy

# Download dependencies
deps:
	go mod download
