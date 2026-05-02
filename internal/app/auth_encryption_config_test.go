package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthEncryptionConfig_ValidateLocalRequiresMasterKey(t *testing.T) {
	_ = os.Unsetenv("WASMFORGE_AUTH_MASTER_KEY")

	cfg := AuthEncryptionConfig{
		Provider:         "local",
		MasterKeyEnvName: "WASMFORGE_AUTH_MASTER_KEY",
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "WASMFORGE_AUTH_MASTER_KEY")
}

func TestAuthEncryptionConfig_ValidateLocalSuccess(t *testing.T) {
	if err := os.Setenv("WASMFORGE_AUTH_MASTER_KEY", "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("WASMFORGE_AUTH_MASTER_KEY") })

	cfg := AuthEncryptionConfig{
		Provider:         "local",
		MasterKeyEnvName: "WASMFORGE_AUTH_MASTER_KEY",
	}

	require.NoError(t, cfg.Validate())
}

func TestAuthEncryptionConfig_ValidateOnePasswordRequiresReference(t *testing.T) {
	cfg := AuthEncryptionConfig{
		Provider: "1password",
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secret reference")
}
