package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Seed     SeedConfig
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port string // 监听端口
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string // SQLite 文件路径
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string // 签名密钥
	ExpireHour int    // Token 过期小时数
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowedOrigins []string // 允许的前端地址
}

// SeedConfig 种子数据配置（首次启动时自动创建）
type SeedConfig struct {
	DefaultHQName        string // 默认总部校区名称
	DefaultAdminUsername string // 默认总部管理员用户名
	DefaultAdminPassword string // 默认总部管理员密码
	DefaultAdminPhone    string // 默认总部管理员手机号
}

// Load 加载配置（环境变量 > .env 文件 > 配置文件 > 默认值）
func Load() *Config {
	// 加载 .env 文件（静默处理：文件不存在时跳过）
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[配置] 未找到 .env 文件，使用系统环境变量和默认值")
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 环境变量绑定
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("database.path", "DB_PATH")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expire_hour", "JWT_EXPIRE_HOUR")
	viper.BindEnv("cors.allowed_origins", "CORS_ORIGINS")
	viper.BindEnv("seed.default_hq_name", "SEED_HQ_NAME")
	viper.BindEnv("seed.default_admin_username", "SEED_ADMIN_USERNAME")
	viper.BindEnv("seed.default_admin_password", "SEED_ADMIN_PASSWORD")
	viper.BindEnv("seed.default_admin_phone", "SEED_ADMIN_PHONE")

	// 默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.path", "./data/school.db")
	viper.SetDefault("jwt.secret", "school-system-secret-key")
	viper.SetDefault("jwt.expire_hour", 24)
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:5173"})
	viper.SetDefault("seed.default_hq_name", "总部")
	viper.SetDefault("seed.default_admin_username", "lfc")
	viper.SetDefault("seed.default_admin_password", "15061805300")
	viper.SetDefault("seed.default_admin_phone", "15061805300")

	viper.ReadInConfig()

	return &Config{
		Server: ServerConfig{
			Port: viper.GetString("server.port"),
		},
		Database: DatabaseConfig{
			Path: viper.GetString("database.path"),
		},
		JWT: JWTConfig{
			Secret:     viper.GetString("jwt.secret"),
			ExpireHour: viper.GetInt("jwt.expire_hour"),
		},
		CORS: CORSConfig{
			AllowedOrigins: viper.GetStringSlice("cors.allowed_origins"),
		},
		Seed: SeedConfig{
			DefaultHQName:        viper.GetString("seed.default_hq_name"),
			DefaultAdminUsername: viper.GetString("seed.default_admin_username"),
			DefaultAdminPassword: viper.GetString("seed.default_admin_password"),
			DefaultAdminPhone:    viper.GetString("seed.default_admin_phone"),
		},
	}
}
