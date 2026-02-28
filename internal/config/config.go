package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Solarman API
	APIID     string
	APISecret string
	BaseURL   string
	Email     string
	Password  string

	// InfluxDB
	InfluxURL    string
	InfluxToken  string
	InfluxOrg    string
	InfluxBucket string

	// App
	LogLevel    string
	OutputDir   string
	PollSeconds int // interval ดึงข้อมูล (วินาที)
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("ไม่พบ .env file - ใช้ environment variables แทน")
	}

	cfg := &Config{
		APIID:        os.Getenv("API_ID"),
		APISecret:    os.Getenv("API_SECRET"),
		BaseURL:      os.Getenv("BASE_URL"),
		Email:        os.Getenv("EMAIL"),
		Password:     os.Getenv("PASSWORD"),
		InfluxURL:    os.Getenv("INFLUX_URL"),
		InfluxToken:  os.Getenv("INFLUX_TOKEN"),
		InfluxOrg:    os.Getenv("INFLUX_ORG"),
		InfluxBucket: os.Getenv("INFLUX_BUCKET"),
		LogLevel:     os.Getenv("LOG_LEVEL"),
		OutputDir:    os.Getenv("OUTPUT_DIR"),
	}

	// defaults
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://globalapi.solarmanpv.com"
	}
	if cfg.InfluxURL == "" {
		cfg.InfluxURL = "http://localhost:8086"
	}
	if cfg.InfluxOrg == "" {
		cfg.InfluxOrg = "solar-org"
	}
	if cfg.InfluxBucket == "" {
		cfg.InfluxBucket = "solar-data"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "./output"
	}
	cfg.PollSeconds = 300 // ดึงข้อมูลทุก 5 นาที
	if v := os.Getenv("POLL_SECONDS"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.PollSeconds)
	}

	if cfg.APIID == "" || cfg.APISecret == "" {
		return nil, fmt.Errorf("กรุณาตั้งค่า API_ID และ API_SECRET ใน .env file")
	}
	if cfg.Email == "" || cfg.Password == "" {
		return nil, fmt.Errorf("กรุณาตั้งค่า EMAIL และ PASSWORD ใน .env file")
	}
	if cfg.InfluxToken == "" {
		return nil, fmt.Errorf("กรุณาตั้งค่า INFLUX_TOKEN ใน .env file")
	}

	return cfg, nil
}