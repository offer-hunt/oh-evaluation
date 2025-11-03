//go:build integration

package integration

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// isDockerAvailable возвращает nil, если можно говорить с Docker daemon.
func isDockerAvailable(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer func() {
		if cerr := cli.Close(); cerr != nil {
			log.Printf("failed to close docker client: %v", cerr)
		}
	}()

	_, err = cli.Ping(ctx)
	return err
}

func TestDbSmoke(t *testing.T) {
	ctx := context.Background()

	// Если Docker демона нет — пропускаем интеграционный тест
	if err := isDockerAvailable(ctx); err != nil {
		t.Skipf("skipping DB smoke test: docker daemon is not available: %v", err)
	}

	// Поднимаем контейнер с ТЕМ ЖЕ именем БД и пользователем,
	// под которые написаны миграции и docker-compose:
	//   DB:   evaluation_db
	//   USER: evaluation_user
	//   PASS: evaluation_password
	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("evaluation_db"),
		postgres.WithUsername("evaluation_user"),
		postgres.WithPassword("evaluation_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	// останавливаем контейнер по завершении
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// получаем DSN к ЭТОЙ БД
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// прогоняем миграции из ./db/migrations в поднятую БД
	m, err := migrate.New("file://./db/migrations", connStr)
	require.NoError(t, err)
	err = m.Up()
	require.NoError(t, err, "не удалось применить миграции")

	// подключаемся через pgxpool с нужным search_path
	poolConfig, err := pgxpool.ParseConfig(connStr + "&search_path=evaluation")
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	require.NoError(t, err)
	defer pool.Close()

	t.Run("проверка создания схемы и таблицы", func(t *testing.T) {
		var count int
		query := `
			SELECT count(*)
			FROM information_schema.tables
			WHERE table_schema = 'evaluation' AND table_name = '__migration_probe'
		`
		err := pool.QueryRow(ctx, query).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Таблица __migration_probe должна существовать в схеме evaluation")
	})

	t.Run("проверка вставки данных в таблицу", func(t *testing.T) {
		var id int64
		query := `INSERT INTO evaluation.__migration_probe DEFAULT VALUES RETURNING id`
		err := pool.QueryRow(ctx, query).Scan(&id)
		require.NoError(t, err)
		assert.Positive(t, id, "ID должен быть положительным числом")

		fmt.Printf("DB Smoke Test: Successfully inserted row with id = %d\n", id)
	})
}
