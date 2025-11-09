.PHONY: migrate-up migrate-down tools lint

DB_URL="postgres://evaluation_user:evaluation_password@localhost:5432/evaluation_db?sslmode=disable"

migrate-up:
	migrate -database "$(DB_URL)" -path db/migrations up

migrate-down:
	migrate -database "$(DB_URL)" -path db/migrations down

GOLANGCI_LINT_VERSION ?= v2.6.0
GO ?= go

# Подтянуть линтер в отдельный modfile (как требует v2)
tools:
	@echo ">>> preparing golangci-lint $(GOLANGCI_LINT_VERSION)"
	@$(GO) mod init -modfile=golangci-lint.mod local/golangci-lint 2>/dev/null || true
	@$(GO) get -tool -modfile=golangci-lint.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# Запуск линтера той же версии
lint: tools
	@echo ">>> running golangci-lint $(GOLANGCI_LINT_VERSION)"
	@$(GO) tool -modfile=golangci-lint.mod golangci-lint run --modules-download-mode=mod --timeout=5m
