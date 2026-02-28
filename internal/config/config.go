package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config เก็บค่าตั้งค่าทั้งหมดของโปรแกรม
type Config struct {
	APIID     string // App ID จาก Solarman
	APISecret string // App Secret จาก Solarman
	BaseURL   string // URL หลักของ API
	Email     string // Email ที่ใช้ login
	Password  string // Password (plain text - จะ hash ใน code)
	LogLevel  string // ระดับ log (info, debug)
	OutputDir string // โฟลเดอร์สำหรับบันทึก JSON
}

// LoadConfig อ่านค่าจาก .env file แล้วคืนค่า Config
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("ไม่พบ .env file - ใช้ environment variables แทน")
	}

	cfg := &Config{
		APIID:     os.Getenv("API_ID"),
		APISecret: os.Getenv("API_SECRET"),
		BaseURL:   os.Getenv("BASE_URL"),
		Email:     os.Getenv("EMAIL"),
		Password:  os.Getenv("PASSWORD"),
		LogLevel:  os.Getenv("LOG_LEVEL"),
		OutputDir: os.Getenv("OUTPUT_DIR"),
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://globalapi.solarmanpv.com"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "./output"
	}

	if cfg.APIID == "" || cfg.APISecret == "" {
		return nil, fmt.Errorf("กรุณาตั้งค่า API_ID และ API_SECRET ใน .env file")
	}
	if cfg.Email == "" || cfg.Password == "" {
		return nil, fmt.Errorf("กรุณาตั้งค่า EMAIL และ PASSWORD ใน .env file")
	}

	return cfg, nil
}