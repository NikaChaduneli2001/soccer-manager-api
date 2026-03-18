.PHONY: run build test tidy docker-up docker-down docker-build clean

# Run the application locally (requires DB or will start without DB)
run:
	go run ./cmd/api

# Build the binary (CGO_ENABLED=0 for portable binary)
build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api

# Run tests
test:
	go test ./...

# Download and tidy dependencies
tidy:
	go mod tidy
	go mod download

# Docker Compose: start all services
docker-up:
	docker-compose up -d

# Docker Compose: stop all services
docker-down:
	docker-compose down

# Docker Compose: build and start (foreground, for development)
docker-up-build:
	docker-compose up --build

# Build Docker image only
docker-build:
	docker-compose build

# Remove build artifacts
clean:
	rm -rf bin/

# Regenerate Swagger docs from controller annotations (requires: go install github.com/swaggo/swag/cmd/swag@latest)
swagger-docs:
	swag init -g cmd/api/main.go --parseDependency --parseInternal

# Run with hot reload (optional: requires air - go install github.com/cosmtrek/air@latest)
# dev:
# 	air
