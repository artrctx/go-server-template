# Build the application
all: build test

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

run-auth:
	@cd ./cmd/auth && pnpm dev

playground:
	@go run cmd/playground/main.go

# Create DB container
docker-run:
	@if docker compose -f docker-compose.app.yml up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose -f docker-compose.app.yml down --volumes --remove-orphans 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

docker-nuke:
	docker system prune --all --force --volumes
	
# Test the application
test:
	export SKIP_ENV_INIT="true" && \
	echo "Testing..." && \
	go test ./... -v && \
	export SKIP_ENV_INIT="false"
	
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Remove unused dep
tidy:
	@go mod tidy
db-up:
	docker compose -f docker-compose.db.yml up -d
db-down:
	docker compose -f docker-compose.db.yml down --volumes
db-auth-sync:
	@pg_dump --schema-only -n auth -f migrations/auth-schema.sql postgres://postgres:password@localhost:5432/postgres
db-migrate:
	migrate create -ext sql -dir migrations -seq ${name}
db-sync:
	migrate -database postgres://postgres:password@localhost:5432/postgres?sslmode=disable -path migrations up
db-desync:
	migrate -database postgres://postgres:password@localhost:5432/postgres?sslmode=disable -path migrations down 
db-generate:
	sqlc generate