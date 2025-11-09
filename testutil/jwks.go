package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"log"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// GenerateRsaKeyAndJwks создает приватный RSA ключ и соответствующий ему публичный JWKS JSON.
func GenerateRsaKeyAndJwks(keyID string) (*rsa.PrivateKey, string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	publicKey, err := jwk.FromRaw(privateKey.PublicKey)
	if err != nil {
		log.Fatalf("failed to create public key: %s", err)
	}
	_ = publicKey.Set(jwk.KeyIDKey, keyID)
	_ = publicKey.Set(jwk.AlgorithmKey, "RS256")
	_ = publicKey.Set(jwk.KeyUsageKey, jwk.ForSignature)

	keySet := jwk.NewSet()
	_ = keySet.AddKey(publicKey)

	jsonBytes, err := json.MarshalIndent(keySet, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal key set: %s", err)
	}

	return privateKey, string(jsonBytes)
}
