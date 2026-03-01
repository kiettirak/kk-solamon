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

	// MySQL
	MySQLDSN string

	// App
	LogLevel    string
	OutputDir   string
	PollMinutes int // interval ดึงข้อมูล (นาที)

	// Weather (Open-Meteo)
	WeatherLat       float64 // latitude ของสถานี (เช่น 13.7563)
	WeatherLon       float64 // longitude (เช่น 100.5018)
	WeatherBackfill  bool    // true = backfill ย้อนหลังอัตโนมัติถ้ายังไม่มีข้อมูล
	WeatherStartDate string  // วันที่เริ่ม backfill (YYYY-MM-DD)
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
		MySQLDSN:     os.Getenv("MYSQL_DSN"),
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
	// Weather coordinates ค่า default = เสกา, บึงกาน
	cfg.WeatherLat = 17.93046324892076
	cfg.WeatherLon = 103.94864700242748
	if v := os.Getenv("WEATHER_LAT"); v != "" {
		fmt.Sscanf(v, "%f", &cfg.WeatherLat)
	}
	if v := os.Getenv("WEATHER_LON"); v != "" {
		fmt.Sscanf(v, "%f", &cfg.WeatherLon)
	}
	cfg.WeatherBackfill = os.Getenv("WEATHER_BACKFILL") != "false"
	cfg.WeatherStartDate = os.Getenv("WEATHER_START_DATE")
	if cfg.WeatherStartDate == "" {
		cfg.WeatherStartDate = "2023-05-23" // วันที่ติดตั้งโซลาร์
	}

	cfg.PollMinutes = 5 // ดึงข้อมูลทุก 5 นาที (default)
	if v := os.Getenv("POLL_MINUTES"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.PollMinutes)
	} else if v := os.Getenv("POLL_SECONDS"); v != "" {
		// backward compat — แปลงวินาทีเป็นนาที
		var secs int
		fmt.Sscanf(v, "%d", &secs)
		if secs > 0 {
			cfg.PollMinutes = secs / 60
			if cfg.PollMinutes < 1 {
				cfg.PollMinutes = 1
			}
		}
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