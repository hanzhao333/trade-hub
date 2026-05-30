package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env       string
	HTTPAddr  string
	DBDSN     string
	RedisAddr string
	JWTSecret string
}

func Load() Config {
	// 本地开发：从项目根目录的 .env 加载；文件不存在则沿用系统环境变量或默认值
	_ = godotenv.Load()
	return Config{
		Env:       getEnv("APP_ENV", "dev"),
		HTTPAddr:  getEnv("HTTP_ADDR", ":8080"),
		DBDSN:     getEnv("DB_DSN", "trade:trade_pass@tcp(127.0.0.1:3306)/trade_hub?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr: getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
