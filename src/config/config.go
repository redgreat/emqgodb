package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	EMQX       EMQXConfig       `mapstructure:"emqx"`
	PostgreSQL PostgreSQLConfig `mapstructure:"postgresql"`
}

// EMQXConfig EMQX连接配置
type EMQXConfig struct {
	Broker   string    `mapstructure:"broker"`
	ClientID string    `mapstructure:"client_id"`
	Username string    `mapstructure:"username"`
	Password string    `mapstructure:"password"`
	Topic    string    `mapstructure:"topic"`
	QoS      int       `mapstructure:"qos"`
	SSL      SSLConfig `mapstructure:"ssl"`
}

// SSLConfig SSL配置
type SSLConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// PostgreSQLConfig PostgreSQL连接配置
type PostgreSQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
	Table    string `mapstructure:"table"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// 支持环境变量覆盖
	v.SetEnvPrefix("EMQGODB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	if cfg.EMQX.Broker == "" {
		return fmt.Errorf("EMQX broker地址不能为空")
	}
	if cfg.EMQX.Topic == "" {
		return fmt.Errorf("EMQX topic不能为空")
	}

	if cfg.PostgreSQL.Host == "" {
		return fmt.Errorf("PostgreSQL主机不能为空")
	}
	if cfg.PostgreSQL.Database == "" {
		return fmt.Errorf("PostgreSQL数据库名不能为空")
	}
	return nil
}

// DSN 返回PostgreSQL连接字符串
func (p *PostgreSQLConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.Username, p.Password, p.Database, p.SSLMode)
}
