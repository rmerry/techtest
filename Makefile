.PHONY: build
build:
	CGO_ENABLED=0 go build -o ./bin/server ./cmd/main.go

.PHONY: swag
swag:
	swag init --output ./docs/swagger -g ./internal/api/handlers.go

.PHONY: swag-fmt
swag-fmt:
	swag fmt
