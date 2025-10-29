package testutil

import (
	"crypto/rsa"
	"log"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// CreateToken создает подписанный JWT для тестов.
func CreateToken(privateKey *rsa.PrivateKey, keyID, issuer, audience string, scopes []string, ttl time.Duration) string {
	// Собираем claims
	token, err := jwt.NewBuilder().
		Issuer(issuer).
		Audience([]string{audience}).
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(ttl)).
		Claim("scp", scopes).
		Build()
	if err != nil {
		log.Fatalf("failed to build token: %s", err)
	}

	// Преобразуем приватный ключ в jwk.Key и задаём kid.
	jwkKey, err := jwk.FromRaw(privateKey)
	if err != nil {
		log.Fatalf("failed to create jwk from private key: %s", err)
	}
	if err := jwkKey.Set(jwk.KeyIDKey, keyID); err != nil {
		log.Fatalf("failed to set kid on jwk: %s", err)
	}

	// Подписываем токен ключом с установленным kid — без дополнительных suboptions.
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, jwkKey))
	if err != nil {
		log.Fatalf("failed to sign token: %s", err)
	}

	return string(signed)
}
