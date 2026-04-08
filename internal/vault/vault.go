package vault

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
)

const (
	defaultAddr  = "http://localhost:8200"
	defaultToken = "admin-token"
	mountPath    = "secret"
	secretPath   = "subscription-hub"

	KeyPostgresPass = "postgres_password"
	KeyRedisPass    = "redis_password"
	KeyJwtSecret    = "jwt_secret"
)

type Secrets map[string]string

func GetSecrets(ctx context.Context) (Secrets, error) {
	config := api.DefaultConfig()

	config.Address = getEnv("VAULT_ADDR", defaultAddr)

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("создание клиента vault: %w", err)
	}

	client.SetToken(getEnv("VAULT_TOKEN", defaultToken))

	secret, err := client.KVv2(mountPath).Get(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("получение секрета %s: %w", secretPath, err)
	}

	res := make(Secrets)
	for k, v := range secret.Data {
		if s, ok := v.(string); ok {
			res[k] = s
		}
	}

	return res, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
