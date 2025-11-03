.PHONY: migrate-up migrate-down tools lint

DB_URL="postgres://evaluation_user:evaluation_password@localhost:5432/evaluation_db?sslmode=disable"

migrate-up:
	migrate -database "$(DB_URL)" -path db/migrations up

migrate-down:
	migrate -database "$(DB_URL)" -path db/migrations down

GOLANGCI_LINT_VERSION ?= v2.6.0
GOLANGCI_LINT_BIN ?= $(HOME)/go/bin/golangci-lint

# установка линтера локально
tools:
	@echo ">>> installing golangci-lint $(GOLANGCI_LINT_VERSION)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# запуск линтера
lint: $(GOLANGCI_LINT_BIN)
	@echo ">>> running golangci-lint"
	@$(GOLANGCI_LINT_BIN) run --modules-download-mode=mod --timeout=5m

# если бинаря нет — сначала ставим
$(GOLANGCI_LINT_BIN):
	$(MAKE) tools
