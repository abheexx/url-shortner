package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Security SecurityConfig `mapstructure:"security"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	MaxRetries   int           `mapstructure:"max_retries"`
	TTL          time.Duration `mapstructure:"ttl"`
	NegativeTTL  time.Duration `mapstructure:"negative_ttl"`
}

type RateLimitConfig struct {
	GlobalRPS    int           `mapstructure:"global_rps"`
	PerIPRPS     int           `mapstructure:"per_ip_rps"`
	BurstSize    int           `mapstructure:"burst_size"`
	WindowSize   time.Duration `mapstructure:"window_size"`
}

type SecurityConfig struct {
	AdminSecret string   `mapstructure:"admin_secret"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedHosts []string `mapstructure:"allowed_hosts"`
	BlockedDomains []string `mapstructure:"blocked_domains"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load() (*Config, error) {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.shutdown_timeout", "30s")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.ttl", "24h")
	viper.SetDefault("redis.negative_ttl", "5m")

	viper.SetDefault("rate_limit.global_rps", 100)
	viper.SetDefault("rate_limit.per_ip_rps", 10)
	viper.SetDefault("rate_limit.burst_size", 20)
	viper.SetDefault("rate_limit.window_size", "1s")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Environment variables
	viper.SetEnvPrefix("URLSHORTENER")
	viper.AutomaticEnv()

	// Config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password,
		c.Database.DBName, c.Database.SSLMode)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}
