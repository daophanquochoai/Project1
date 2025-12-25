package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"pass"`
	DB       int    `mapstructure:"db"`
}

type ServerConfig struct {
	HTTPPort string `mapstructure:"http_port"`
	Grpc     GRPC   `mapstructure:"grpc"`
	Env      string `mapstructure:"env"`
}

type GRPC struct {
	Auth AuthGrpc `mapstructure:"auth"`
}

type AuthGrpc struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	TimeZone string `mapstructure:"timezone"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.productservice")

	viper.SetEnvPrefix("PRODUCT_SERVICE")
	viper.AutomaticEnv()

	// Bind environment variables
	viper.BindEnv("server.http_port", "HTTP_PORT")
	viper.BindEnv("server.env", "ENV")
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
	viper.BindEnv("database.timezone", "DB_TIMEZONE")
	viper.BindEnv("redis.addr", "REDIS_ADDR")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	//setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Config file not found, using defaults and environment variables: %v\n", err)
	} else {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	var cfg Config

	// Unmarshal the config into struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &cfg, nil
}
