package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	authmocks "github.com/mikhail5545/wasmforge/internal/database/auth/mocks"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestValidator_ValidateToken_EnvBackend(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	_ = privatePEM

	if err := os.Setenv("WF_ENV_PUBLIC", publicPEM); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("WF_ENV_PUBLIC") })

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := authmocks.NewMockKeyMaterialRepository(ctrl)
	validator := NewTokenValidator(keyRepo, nil, zap.NewNop())

	metadata := `{"env_public_key_var":"WF_ENV_PUBLIC","env_key_id":"env-kid"}`
	authConfig := &configmodel.AuthConfig{
		ID:              uuid.New(),
		KeyBackendType:  configmodel.KeyBackendTypeEnv,
		Metadata:        metadata,
		TokenAudience:   "env-aud",
		TokenIssuer:     "env-iss",
		TokenTTLSeconds: 3600,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "env-user",
		"iss": "env-iss",
		"aud": jwt.ClaimStrings{"env-aud"},
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	})
	token.Header["kid"] = "env-kid"

	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	validated, err := validator.ValidateToken(context.Background(), tokenString, authConfig)
	require.NoError(t, err)
	assert.Equal(t, "env-user", validated.Subject)
	assert.Equal(t, "env-kid", validated.KeyID)
}

func TestValidator_ValidateToken_JWKSBackend(t *testing.T) {
	privateKey, _, _ := testGenerateRSAKeyPairPEM(t)
	n := base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"keys":[{"kty":"RSA","kid":"jwks-kid","n":"` + n + `","e":"` + e + `"}]}`))
	}))
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := authmocks.NewMockKeyMaterialRepository(ctrl)
	validator := NewTokenValidator(keyRepo, nil, zap.NewNop())

	cacheTTL := 60
	authConfig := &configmodel.AuthConfig{
		ID:                  uuid.New(),
		KeyBackendType:      configmodel.KeyBackendTypeJWKS,
		JWKSUrl:             server.URL,
		JWKSCacheTTLSeconds: &cacheTTL,
		TokenAudience:       "jwks-aud",
		TokenIssuer:         "jwks-iss",
		TokenTTLSeconds:     3600,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "jwks-user",
		"iss": "jwks-iss",
		"aud": jwt.ClaimStrings{"jwks-aud"},
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	})
	token.Header["kid"] = "jwks-kid"

	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	validated, err := validator.ValidateToken(context.Background(), tokenString, authConfig)
	require.NoError(t, err)
	assert.Equal(t, "jwks-user", validated.Subject)
}

func TestIssuer_IssueToken_EnvBackend(t *testing.T) {
	privateKey, privatePEM, _ := testGenerateRSAKeyPairPEM(t)
	if err := os.Setenv("WF_ENV_PRIVATE", privatePEM); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("WF_ENV_PRIVATE") })

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := authmocks.NewMockKeyMaterialRepository(ctrl)
	issuer := NewTokenIssuer(keyRepo, nil, zap.NewNop())

	authConfig := &configmodel.AuthConfig{
		ID:              uuid.New(),
		KeyBackendType:  configmodel.KeyBackendTypeEnv,
		Metadata:        `{"env_private_key_var":"WF_ENV_PRIVATE","env_key_id":"env-issue"}`,
		TokenAudience:   "issue-aud",
		TokenIssuer:     "issue-iss",
		TokenTTLSeconds: 3600,
	}

	tokenString, err := issuer.IssueToken(context.Background(), map[string]interface{}{"sub": "issued-user"}, authConfig)
	require.NoError(t, err)

	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)
}

func TestIssuer_IssueToken_JWKSBackendRejected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := authmocks.NewMockKeyMaterialRepository(ctrl)
	issuer := NewTokenIssuer(keyRepo, nil, zap.NewNop())

	authConfig := &configmodel.AuthConfig{
		ID:             uuid.New(),
		KeyBackendType: configmodel.KeyBackendTypeJWKS,
		JWKSUrl:        "https://example.invalid/jwks",
	}

	_, err := issuer.IssueToken(context.Background(), map[string]interface{}{"sub": "issued-user"}, authConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwks backend")
}

func TestParseJWKRSAPublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	jwk := map[string]interface{}{
		"kty": "RSA",
		"kid": "test",
		"n":   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
	}
	publicKey, err := parseJWKRSAPublicKey(jwk)
	require.NoError(t, err)
	assert.Equal(t, privateKey.PublicKey.N, publicKey.N)
	assert.Equal(t, privateKey.PublicKey.E, publicKey.E)
}
