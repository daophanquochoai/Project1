package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Redis    RedisConfig
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"pass"`
	DB       int    `mapstructure:"db"`
}

type ServerConfig struct {
	HTTPPort string `mapstructure:"http_port"`
	GRPCPort string `mapstructure:"grpc_port"`
	Env      string `mapstructure:"env"`
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

type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.user-service")

	viper.SetEnvPrefix("USER_SERVICE")
	viper.AutomaticEnv()

	// Bind environment variables
	viper.BindEnv("server.http_port", "HTTP_PORT")
	viper.BindEnv("server.grpc_port", "GRPC_PORT")
	viper.BindEnv("server.env", "ENV")
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
	viper.BindEnv("database.timezone", "DB_TIMEZONE")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.access_expiry", "JWT_ACCESS_EXPIRY")
	viper.BindEnv("jwt.refresh_expiry", "JWT_REFRESH_EXPIRY")
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

	// Validate required fields
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

//
//func setDefaults() {
//	// Server defaults
//	viper.SetDefault("server.http_port", "8005")
//	viper.SetDefault("server.grpc_port", "9005")
//	viper.SetDefault("server.env", "development")
//
//	// Database defaults
//	viper.SetDefault("database.host", "localhost")
//	viper.SetDefault("database.port", "5432")
//	viper.SetDefault("database.user", "user_service")
//	viper.SetDefault("database.password", "user_service_pwd")
//	viper.SetDefault("database.name", "user_service")
//	viper.SetDefault("database.sslmode", "disable")
//	viper.SetDefault("database.timezone", "UTC")
//
//	// JWT defaults
//	viper.SetDefault("jwt.secret", "default-secret-change-in-production")
//	viper.SetDefault("jwt.access_expiry", "15m")
//	viper.SetDefault("jwt.refresh_expiry", "168h") // 7 days
//
//	// Redis defaults
//	viper.SetDefault("redis.addr", "localhost:6379")
//	viper.SetDefault("redis.password", "")
//	viper.SetDefault("redis.db", 0)
//}

func validateConfig(cfg *Config) error {
	if cfg.JWT.Secret == "default-secret-change-in-production" && cfg.Server.Env == "production" {
		return fmt.Errorf("JWT_SECRET must be changed in production")
	}

	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	if cfg.JWT.AccessExpiry == 0 {
		return fmt.Errorf("JWT access expiry is required")
	}

	if cfg.JWT.RefreshExpiry == 0 {
		return fmt.Errorf("JWT refresh expiry is required")
	}

	return nil
}
