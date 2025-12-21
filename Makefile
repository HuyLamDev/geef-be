.PHONY: build run lint format clean install-lint swagger

# Generate Swagger documentation
swagger:
	~/go/bin/swag init -g cmd/main.go

# Build the application
build:
	go build -o geef-be ./cmd

# Run the application (development mode - faster)
run: swagger
	go run cmd/main.go

# Lint the code (requires golangci-lint)
lint:
	golangci-lint run

# Format the code
format:
	gofmt -w .
	goimports -w .

# Clean build artifacts
clean:
	rm -f geef-be

# Install golangci-lint if not present
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install goimports if not present
install-tools: install-lint
	go install golang.org/x/tools/cmd/goimports@latest

# Tidy dependencies
tidy:
	go mod tidy

# Default target
all: build