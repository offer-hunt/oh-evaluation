package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/offer-hunt/oh-evaluation/internal/config"
)

type contextKey string

const TokenContextKey contextKey = "jwt"

// Authenticator содержит все необходимое для валидации токенов.
type Authenticator struct {
	keySet   jwk.Set
	issuer   string
	audience string
}

// NewAuthenticator создает новый экземпляр Authenticator.
func NewAuthenticator(ctx context.Context, cfg *config.Config) (*Authenticator, error) {
	// --- ИЗМЕНЕНИЯ ЗДЕСЬ ---

	// Шаг 1: Создаем объект кэша.
	cache := jwk.NewCache(ctx)

	// Шаг 2: Регистрируем наш JWKS URL в кэше с опцией авто-обновления.
	err := cache.Register(cfg.AuthJwksURL, jwk.WithRefreshInterval(15*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("failed to register JWKS URL for auto-refresh: %w", err)
	}

	// Шаг 3 (Важно для "fail-fast"): Выполняем первую принудительную загрузку ключей.
	// Это гарантирует, что JWKS URL доступен при старте приложения.
	_, err = cache.Get(ctx, cfg.AuthJwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch initial JWKS: %w", err)
	}

	// Шаг 4: Создаем jwk.Set, используя наш настроенный кэш.
	// Эта функция НЕ возвращает ошибку.
	keySet := jwk.NewCachedSet(cache, cfg.AuthJwksURL)

	// --- КОНЕЦ ИЗМЕНЕНИЙ ---

	return &Authenticator{
		keySet:   keySet,
		issuer:   cfg.AuthIssuer,
		audience: cfg.AuthAudience,
	}, nil
}

// Authenticator - это middleware для проверки JWT.
func (a *Authenticator) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// Парсим и валидируем токен
		token, err := jwt.Parse([]byte(tokenString),
			jwt.WithKeySet(a.keySet),
			jwt.WithValidate(true),
			jwt.WithIssuer(a.issuer),
			jwt.WithAudience(a.audience),
		)
		if err != nil {
			log.Printf("Token validation error: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Сохраняем токен в контексте для дальнейшего использования
		ctx := context.WithValue(r.Context(), TokenContextKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireScope - это middleware для проверки наличия нужных прав (scope) в токене.
func RequireScope(requiredScope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := r.Context().Value(TokenContextKey).(jwt.Token)
			if !ok {
				// Этого не должно случиться, если Authenticator отработал корректно
				http.Error(w, "Could not retrieve token from context", http.StatusInternalServerError)
				return
			}

			scpClaim, found := token.Get("scp")
			if !found {
				http.Error(w, "Forbidden: scp claim is missing", http.StatusForbidden)
				return
			}

			scopes, ok := scpClaim.([]interface{})
			if !ok {
				http.Error(w, "Forbidden: scp claim has invalid format", http.StatusForbidden)
				return
			}

			for _, scope := range scopes {
				if s, ok := scope.(string); ok && s == requiredScope {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Forbidden: insufficient scope", http.StatusForbidden)
		})
	}
}
