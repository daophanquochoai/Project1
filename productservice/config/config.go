package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2/log"
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
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name"`
	SSLMode      string `mapstructure:"sslmode"`
	TimeZone     string `mapstructure:"timezone"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

func getDefaultConfig() string {
	return "/config/config"
}

func Load() (*Config, error) {
	config := Config{}
	path := os.Getenv("cfgPath")
	if path == "" {
		path = getDefaultConfig()
	}
	fmt.Printf("config path:%s\n", path)

	v := viper.New()

	v.SetConfigName(path)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}

	err := v.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
		return nil, err
	}

	return &config, nil
}
