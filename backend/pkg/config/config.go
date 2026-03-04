package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用全局配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	TKE      TKEConfig      `mapstructure:"tke"`
	WeChat   WeChatConfig   `mapstructure:"wechat"`
}

// ServerConfig HTTP 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug / release
}

// DatabaseConfig PostgreSQL 数据库配置
type DatabaseConfig struct {
	URL          string `mapstructure:"url"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// RedisConfig Redis 缓存配置
type RedisConfig struct {
	URL string `mapstructure:"url"`
}

// JWTConfig JWT 认证配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// TKEConfig 腾讯云 TKE 集群配置
type TKEConfig struct {
	ClusterID   string `mapstructure:"cluster_id"`
	Region      string `mapstructure:"region"`
	SecretID    string `mapstructure:"secret_id"`
	SecretKey   string `mapstructure:"secret_key"`
	IngressDomain string `mapstructure:"ingress_domain"` // 实例访问域名后缀，如 tke-cloud.com
}

// WeChatConfig 企业微信通知配置
type WeChatConfig struct {
	CorpID  string `mapstructure:"corp_id"`
	AgentID string `mapstructure:"agent_id"`
	Secret  string `mapstructure:"secret"`
}

// Load 从环境变量和配置文件加载配置
func Load() (*Config, error) {
	v := viper.New()

	// 设置环境变量前缀（自动大写）
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 绑定环境变量到配置项
	// 服务器配置
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.mode", "SERVER_MODE")

	// 数据库配置
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	v.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")

	// Redis 配置
	v.BindEnv("redis.url", "REDIS_URL")

	// JWT 配置
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("jwt.expire_hours", "JWT_EXPIRE_HOURS")

	// TKE 配置
	v.BindEnv("tke.cluster_id", "TKE_CLUSTER_ID")
	v.BindEnv("tke.region", "TKE_REGION")
	v.BindEnv("tke.secret_id", "QCLOUD_SECRET_ID")
	v.BindEnv("tke.secret_key", "QCLOUD_SECRET_KEY")
	v.BindEnv("tke.ingress_domain", "TKE_INGRESS_DOMAIN")

	// 企业微信配置
	v.BindEnv("wechat.corp_id", "WECHAT_CORP_ID")
	v.BindEnv("wechat.agent_id", "WECHAT_AGENT_ID")
	v.BindEnv("wechat.secret", "WECHAT_SECRET")

	// 设置默认值
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("database.max_open_conns", 10)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("jwt.expire_hours", 24)
	v.SetDefault("tke.region", "ap-guangzhou")
	v.SetDefault("tke.ingress_domain", "openclaw.example.com")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证必填配置
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL 不能为空")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET 不能为空")
	}

	return &cfg, nil
}
