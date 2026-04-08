package config

import "fmt"

const scrubbed = "***"

type (
	Config struct {
		Server   ServerConfig   `mapstructure:"server"`
		Postgres PostgresConfig `mapstructure:"postgres"`
		Redis    RedisConfig    `mapstructure:"redis"`
		JWT      JWTConfig      `mapstructure:"jwt"`
	}

	ServerConfig struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
	}

	PostgresConfig struct {
		Name     string `mapstructure:"name"`
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"-"`
	}

	RedisConfig struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"-"`
		DB       int    `mapstructure:"db"`
	}

	JWTConfig struct {
		Secret string `mapstructure:"-"`
	}
)

func (c *Config) DSN() string {
	p := c.Postgres
	return fmt.Sprintf("dbname=%s host=%s port=%s user=%s password=%s sslmode=disable",
		p.Name, p.Host, p.Port, p.User, p.Password,
	)
}

func (c *Config) JWTSecret() string {
	return c.JWT.Secret
}

func (c *Config) String() string {
	return fmt.Sprintf("{Server:%s Postgres:%s Redis:%s JWT:%s}", c.Server, c.Postgres, c.Redis, c.JWT)
}

func (s ServerConfig) String() string {
	return fmt.Sprintf("{Host:%s Port:%s}", s.Host, s.Port)
}

func (p PostgresConfig) String() string {
	return fmt.Sprintf("{Name:%s Host:%s Port:%s User:%s Password:%s}", p.Name, p.Host, p.Port, p.User, scrubbed)
}

func (r RedisConfig) String() string {
	return fmt.Sprintf("{Addr:%s DB:%d Password:%s}", r.Addr, r.DB, scrubbed)
}

func (j JWTConfig) String() string {
	return fmt.Sprintf("{Secret:%s}", scrubbed)
}
