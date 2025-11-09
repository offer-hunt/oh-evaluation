package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/offer-hunt/oh-evaluation/internal/config"
	"github.com/offer-hunt/oh-evaluation/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurity(t *testing.T) {
	// 1. Настройка Mock JWKS сервера
	privateKey, jwksJSON := testutil.GenerateRsaKeyAndJwks("test-kid")

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(jwksJSON)); err != nil {
			t.Fatalf("failed to write JWKS response: %v", err)
		}
	}))
	defer jwksServer.Close()

	// 2. Настройка конфигурации для тестов
	cfg := &config.Config{
		AuthIssuer:   "test-issuer",
		AuthAudience: "test-audience",
		AuthJwksURL:  jwksServer.URL,
	}

	auth, err := NewAuthenticator(context.Background(), cfg)
	require.NoError(t, err)

	// 3. Создаем тестовый хендлер, до которого дойдет запрос, если все middleware пройдут
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			t.Fatalf("failed to write test handler response: %v", err)
		}
	})

	// Собираем полный обработчик с middleware
	fullHandler := auth.Authenticator(RequireScope("evaluation.read")(testHandler))

	t.Run("Ответ 200 при валидном токене с нужным scope", func(t *testing.T) {
		token := testutil.CreateToken(privateKey, "test-kid", cfg.AuthIssuer, cfg.AuthAudience, []string{"evaluation.read"}, 1*time.Hour)
		req := httptest.NewRequest("GET", "/secure", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		fullHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Ответ 401 при отсутствии токена", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/secure", nil)
		rr := httptest.NewRecorder()
		fullHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Authorization header is required")
	})

	t.Run("Ответ 401 при невалидном токене (неверная подпись)", func(t *testing.T) {
		otherKey, _ := testutil.GenerateRsaKeyAndJwks("other-kid")
		token := testutil.CreateToken(otherKey, "other-kid", cfg.AuthIssuer, cfg.AuthAudience, []string{"evaluation.read"}, 1*time.Hour)
		req := httptest.NewRequest("GET", "/secure", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		fullHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid token")
	})

	t.Run("Ответ 403 при нехватке прав (scope)", func(t *testing.T) {
		token := testutil.CreateToken(privateKey, "test-kid", cfg.AuthIssuer, cfg.AuthAudience, []string{"other.scope"}, 1*time.Hour)
		req := httptest.NewRequest("GET", "/secure", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		fullHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "insufficient scope")
	})
}
