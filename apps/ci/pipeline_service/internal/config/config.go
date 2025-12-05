package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig       `mapstructure:"database"`
	GRPC     GRPCConfig           `mapstructure:"grpc"`
	HTTP     HTTPConfig           `mapstructure:"http"`
	Log      LogConfig            `mapstructure:"log"`
	Metrics  MetricsConfig        `mapstructure:"metrics"`
	Health   HealthConfig         `mapstructure:"health"`
	Auth     AuthConfig           `mapstructure:"auth"`
	App      AppConfig            `mapstructure:"app"`
	Logging  LoggingConfig        `mapstructure:"logging"`
	Project  ProjectClientConfig  `mapstructure:"project"`
	Executor ExecutorClientConfig `mapstructure:"executor"`
	Queue    QueueConfig          `mapstructure:"queue"`
}

type DatabaseConfig struct {
	URL      string `mapstructure:"url"`
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

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

type HealthConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	Path          string `mapstructure:"path"`
	LivenessPath  string `mapstructure:"liveness_path"`
	ReadinessPath string `mapstructure:"readiness_path"`
	StartupPath   string `mapstructure:"startup_path"`
	Address       string `mapstructure:"address"`
}

type AuthConfig struct {
	PublicPaths []string `mapstructure:"public_paths"`
	Timeout     int      `mapstructure:"timeout"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

type LoggingConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Level   string `mapstructure:"level"`
	Format  string `mapstructure:"format"`
}

// Project 客户端配置（用于拨号外部 Project gRPC 服务）
type ProjectClientConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

// Executor 客户端配置（用于拨号执行器 gRPC 服务）
type ExecutorClientConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

// 队列配置（RabbitMQ等）
type QueueConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
	Queue   string `mapstructure:"queue"`
}

func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.db_name", "DB_NAME")
	viper.BindEnv("database.ssl_mode", "DB_SSLMODE")

	viper.BindEnv("grpc.address", "GRPC_ADDRESS")
	viper.BindEnv("grpc.port", "GRPC_PORT")
	viper.BindEnv("http.address", "HTTP_ADDRESS")
	viper.BindEnv("http.port", "HTTP_PORT")

	viper.BindEnv("project.address", "PROJECT_GRPC_ADDRESS")
	viper.BindEnv("project.port", "PROJECT_GRPC_PORT")
	viper.BindEnv("executor.address", "EXECUTOR_GRPC_ADDRESS")
	viper.BindEnv("executor.port", "EXECUTOR_GRPC_PORT")

	// 队列（RabbitMQ）配置
	viper.BindEnv("queue.enabled", "RABBITMQ_ENABLED")
	viper.BindEnv("queue.url", "RABBITMQ_URL")
	viper.BindEnv("queue.queue", "RABBITMQ_QUEUE")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	return &config, nil
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}
