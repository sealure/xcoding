package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Log      LogConfig      `mapstructure:"log"`
	Queue    QueueConfig    `mapstructure:"queue"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type GRPCConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}
type HTTPConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}
type QueueConfig struct {
	URL   string `mapstructure:"url"`
	Queue string `mapstructure:"queue"`
}

func (c *Config) GRPCAddr() string               { return fmt.Sprintf("%s:%d", c.GRPC.Address, c.GRPC.Port) }
func (c *Config) HTTPAddr() string               { return fmt.Sprintf("%s:%d", c.HTTP.Address, c.HTTP.Port) }
func (c *Config) ShutdownTimeout() time.Duration { return 30 * time.Second }

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.BindEnv("database.host", "EXECUTOR_DB_HOST")
	viper.BindEnv("database.port", "EXECUTOR_DB_PORT")
	viper.BindEnv("database.user", "EXECUTOR_DB_USER")
	viper.BindEnv("database.password", "EXECUTOR_DB_PASSWORD")
	viper.BindEnv("database.db_name", "EXECUTOR_DB_NAME")
	viper.BindEnv("database.ssl_mode", "EXECUTOR_DB_SSLMODE")

	viper.BindEnv("grpc.address", "EXECUTOR_GRPC_ADDRESS")
	viper.BindEnv("grpc.port", "EXECUTOR_GRPC_PORT")
	viper.BindEnv("http.address", "EXECUTOR_HTTP_ADDRESS")
	viper.BindEnv("http.port", "EXECUTOR_HTTP_PORT")

	viper.BindEnv("queue.url", "RABBITMQ_URL")
	viper.BindEnv("queue.queue", "RABBITMQ_QUEUE")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse executor config: %w", err)
	}
	return &cfg, nil
}
