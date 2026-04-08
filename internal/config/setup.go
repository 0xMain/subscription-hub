package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/0xMain/subscription-hub/internal/vault"
	"github.com/spf13/viper"
)

func Load(ctx context.Context) (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := bindServer(v); err != nil {
		return nil, fmt.Errorf("настройка сервера: %w", err)
	}
	if err := bindPostgres(v); err != nil {
		return nil, fmt.Errorf("настройка базы данных: %w", err)
	}
	if err := bindRedis(v); err != nil {
		return nil, fmt.Errorf("настройка кэша: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFound) {
			log.Printf("файл конфигурации не найден, используются значения по умолчанию")
		} else {
			return nil, fmt.Errorf("чтение конфигурации: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("преобразование конфигурации: %w", err)
	}

	vs, err := vault.GetSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("получение секретов из vault: %w", err)
	}

	cfg.Postgres.Password = vs[vault.KeyPostgresPass]
	cfg.Redis.Password = vs[vault.KeyRedisPass]
	cfg.JWT.Secret = vs[vault.KeyJwtSecret]

	if cfg.JWT.Secret == "" {
		return nil, errors.New("jwt секрет не заполнен")
	}

	return cfg, nil
}

func bindServer(v *viper.Viper) error {
	err := bind(v, []struct{ key, env string }{
		{"server.host", "SERVER_HOST"},
		{"server.port", "SERVER_PORT"},
	})
	if err != nil {
		return err
	}

	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", "8080")

	return nil
}

func bindPostgres(v *viper.Viper) error {
	err := bind(v, []struct{ key, env string }{
		{"postgres.name", "POSTGRES_DB"},
		{"postgres.user", "POSTGRES_USER"},
		{"postgres.host", "POSTGRES_HOST"},
		{"postgres.port", "POSTGRES_PORT"},
	})
	if err != nil {
		return err
	}

	v.SetDefault("postgres.name", "subscription_hub")
	v.SetDefault("postgres.user", "admin")
	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", "5432")

	return nil
}

func bindRedis(v *viper.Viper) error {
	err := bind(v, []struct{ key, env string }{
		{"redis.addr", "REDIS_ADDR"},
		{"redis.db", "REDIS_DB"},
	})
	if err != nil {
		return err
	}

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)

	return nil
}

func bind(v *viper.Viper, binds []struct{ key, env string }) error {
	for _, b := range binds {
		if err := v.BindEnv(b.key, b.env); err != nil {
			return fmt.Errorf("привязка переменной окружения %s: %w", b.key, err)
		}
	}

	return nil
}
