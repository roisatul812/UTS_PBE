package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	JWTSecret         string
	JWTExpiresIn      int
	RateLimitLoginMax int
	RateLimitLoginExp int
}

var AppConfig *Config

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or could not be loaded, using environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:isa123@localhost:5432/siakad_mini?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-jwt-secret-key-siakad-mini"
	}

	jwtExpiresIn, err := strconv.Atoi(os.Getenv("JWT_EXPIRES_IN"))
	if err != nil || jwtExpiresIn <= 0 {
		jwtExpiresIn = 86400 // 24 hours in seconds
	}

	rateLimitMax, err := strconv.Atoi(os.Getenv("RATE_LIMIT_LOGIN_MAX"))
	if err != nil || rateLimitMax <= 0 {
		rateLimitMax = 5
	}

	rateLimitExp, err := strconv.Atoi(os.Getenv("RATE_LIMIT_LOGIN_EXP"))
	if err != nil || rateLimitExp <= 0 {
		rateLimitExp = 60 // 60 seconds
	}

	AppConfig = &Config{
		Port:              port,
		DatabaseURL:       dbURL,
		JWTSecret:         jwtSecret,
		JWTExpiresIn:      jwtExpiresIn,
		RateLimitLoginMax: rateLimitMax,
		RateLimitLoginExp: rateLimitExp,
	}

	return AppConfig
}
