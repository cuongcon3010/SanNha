.PHONY: run build test test-unit lint fmt db-up db-down
run:
	go run ./cmd/api
build:
	go build -o bin/api ./cmd/api
	go build -o bin/admin ./cmd/admin
test:
	@test -n "$$TEST_DATABASE_URL" || (echo "Set TEST_DATABASE_URL to a dedicated PostgreSQL test database"; exit 1)
	go test -race ./...
test-unit:
	TEST_DATABASE_URL= go test ./...
lint:
	go vet ./...
fmt:
	gofmt -w cmd internal
db-up:
	docker compose up -d --wait postgres
db-down:
	docker compose down
