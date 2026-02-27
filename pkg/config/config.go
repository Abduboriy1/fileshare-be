package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	RefreshSecret   string
	AWSRegion       string
	AWSAccessKey    string
	AWSSecretKey    string
	S3Bucket        string
	S3Endpoint      string
	FrontendOrigin  string
	AppEnv          string
	AppKey          string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		RefreshSecret:  os.Getenv("REFRESH_SECRET"),
		AWSRegion:      os.Getenv("AWS_REGION"),
		AWSAccessKey:   os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretKey:   os.Getenv("AWS_SECRET_ACCESS_KEY"),
		S3Bucket:       os.Getenv("S3_BUCKET"),
		S3Endpoint:     os.Getenv("S3_ENDPOINT"),
		FrontendOrigin: os.Getenv("FRONTEND_ORIGIN"),
		AppEnv:         getEnv("APP_ENV", "production"),
		AppKey:         os.Getenv("APP_KEY"),
	}

	required := map[string]string{
		"DATABASE_URL":   cfg.DatabaseURL,
		"JWT_SECRET":     cfg.JWTSecret,
		"REFRESH_SECRET": cfg.RefreshSecret,
		"AWS_REGION":     cfg.AWSRegion,
		"S3_BUCKET":      cfg.S3Bucket,
		"APP_KEY":        cfg.AppKey,
	}

	for name, val := range required {
		if val == "" {
			return nil, fmt.Errorf("required environment variable %s is not set", name)
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
