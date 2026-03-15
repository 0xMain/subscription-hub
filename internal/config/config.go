package config

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`

	Database struct {
		Name     string `mapstructure:"name"`
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
	} `mapstructure:"database"`

	JWT struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"jwt"`
}

func Load(ctx context.Context) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "8080")

	viper.SetDefault("database.name", "mail_queue")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5433")
	viper.SetDefault("database.user", "admin")
	viper.SetDefault("database.password", "admin")

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFound) {
			log.Printf("файл конфигурации не найден, используются значения по умолчанию")
		} else {
			return nil, fmt.Errorf("ошибка чтения конфигурации: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("ошибка преобразования конфигурации: %w", err)
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("jwt секрет не указан")
	}

	return &cfg, nil
}

func (config *Config) DSN() string {
	return fmt.Sprintf("dbname=%s host=%s port=%s user=%s password=%s sslmode=disable",
		config.Database.Name,
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
	)
}

func (config *Config) JWTSecret() string {
	return config.JWT.Secret
}
