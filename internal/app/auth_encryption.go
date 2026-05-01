package app

import (
	"context"
	"os"

	"github.com/mikhail5545/wasmforge/internal/services/auth/encryption"
)

func (a *App) buildAuthEncryption(ctx context.Context) (encryption.KeyEncryptionProvider, *encryption.KeyEncryptionRegistry, error) {
	if err := a.cfg.AuthEncryption.Validate(); err != nil {
		return nil, nil, err
	}

	switch a.cfg.AuthEncryption.Provider {
	case "", "local":
		envName := a.cfg.AuthEncryption.MasterKeyEnvName
		if envName == "" {
			envName = "WASMFORGE_AUTH_MASTER_KEY"
		}
		provider, err := encryption.NewLocalKeyEncryptionProviderFromEnv(envName, a.logger)
		if err != nil {
			return nil, nil, err
		}
		return provider, encryption.NewKeyEncryptionRegistry(provider), nil
	case "1password":
		tokenEnv := a.cfg.AuthEncryption.OnePasswordTokenEnv
		if tokenEnv == "" {
			tokenEnv = "OP_SERVICE_ACCOUNT_TOKEN"
		}
		token := os.Getenv(tokenEnv)
		provider, err := encryption.NewOnePasswordKeyEncryptionProvider(
			ctx,
			a.cfg.AuthEncryption.OnePasswordReference,
			token,
			a.cfg.AuthEncryption.OnePasswordIntegration,
			"",
			a.logger,
		)
		if err != nil {
			return nil, nil, err
		}
		return provider, encryption.NewKeyEncryptionRegistry(provider), nil
	case "aws-kms":
		provider, err := encryption.NewAwsKmsKeyEncryptionProvider(
			ctx,
			a.cfg.AuthEncryption.AWSKMSRegion,
			a.cfg.AuthEncryption.AWSKMSKeyID,
			a.logger,
		)
		if err != nil {
			return nil, nil, err
		}
		return provider, encryption.NewKeyEncryptionRegistry(provider), nil
	default:
		return nil, nil, &ConfigError{Message: "unsupported auth encryption provider: " + a.cfg.AuthEncryption.Provider}
	}
}
