.PHONY: migrate-up migrate-down

DB_URL="postgres://evaluation_user:evaluation_password@localhost:5432/evaluation_db?sslmode=disable"

migrate-up:
	migrate -database "$(DB_URL)" -path db/migrations up

migrate-down:
	migrate -database "$(DB_URL)" -path db/migrations down