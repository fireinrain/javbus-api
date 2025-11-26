package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// ==========================================
// 配置结构体（按 TOML 分组）
// ==========================================

type ServerConfig struct {
	DebugLevel string `mapstructure:"debug_level"`
	ServerPort int    `mapstructure:"server_port"`
}

type ProxyConfig struct {
	HttpProxy   string `mapstructure:"http_proxy"`
	HttpsProxy  string `mapstructure:"https_proxy"`
	Socks5Proxy string `mapstructure:"socks5_proxy"`
}

type AdminConfig struct {
	AdminUsername string `mapstructure:"admin_username"`
	AdminPassword string `mapstructure:"admin_password"`
}

type AuthConfig struct {
	JavbusJwtToken      string `mapstructure:"javbus_jwt_token"`
	JavbusSessionSecret string `mapstructure:"javbus_session_secret"`
}

type DatabaseConfig struct {
	DBType       string `mapstructure:"db_type"`
	DBServerPath string `mapstructure:"db_server_path"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Proxy    ProxyConfig    `mapstructure:"proxy"`
	Admin    AdminConfig    `mapstructure:"admin"`
	Auth     AuthConfig     `mapstructure:"auth"`
	DATABASE DatabaseConfig `mapstructure:"database"`
}

var GlobalConfig *Config

var (
	proxyRegex = regexp.MustCompile(`^(https?|socks5?):\/\/`)
)

// ==========================================
// 加载配置 (支持 TOML)
// ==========================================
func InitConfig() (*Config, error) {
	// 1. 加载 .env
	_ = godotenv.Load()

	v := viper.New()

	// 设置默认值（对应 server)
	v.SetDefault("server.debug_level", "debug")
	v.SetDefault("server.server_port", 3000)

	// 使用 TOML
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(".")

	// 读取 config.toml
	if err := v.ReadInConfig(); err != nil {
		var pathError *os.PathError
		if errors.As(err, &pathError) {
			fmt.Println("注意: 未找到 config.toml，将仅使用环境变量")
		} else {
			return nil, err
		}
	}

	// 环境变量覆盖
	v.AutomaticEnv()

	//映射配置
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("配置解析失败: %w", err)
	}

	// 校验
	if err := validate(&config); err != nil {
		return nil, err
	}
	GlobalConfig = &config

	return &config, nil
}

// ==========================================
// 校验逻辑
// ==========================================
func validate(c *Config) error {
	// Proxy 格式检查
	if c.Proxy.HttpProxy != "" && !proxyRegex.MatchString(c.Proxy.HttpProxy) {
		return fmt.Errorf("HTTP_PROXY 格式错误: 必须以 http://, https://, socks:// 或 socks5:// 开头")
	}

	if c.Proxy.HttpsProxy != "" && !proxyRegex.MatchString(c.Proxy.HttpsProxy) {
		return fmt.Errorf("HTTPS_PROXY 格式错误: 必须以 http://, https://, socks:// 或 socks5:// 开头")
	}

	if c.Proxy.Socks5Proxy != "" && !proxyRegex.MatchString(c.Proxy.Socks5Proxy) {
		return fmt.Errorf("SOCKS5_PROXY 格式错误: 必须以 socks:// 或 socks5:// 开头")
	}

	return nil
}
