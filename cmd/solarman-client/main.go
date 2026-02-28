package main

import (
	"log"
	"solarman-go-client/internal/config"
	"solarman-go-client/internal/service"
	"solarman-go-client/internal/solarman"
	"solarman-go-client/internal/store"
	"time"
)

func main() {
	log.Println("=== Solarman → InfluxDB Pipeline เริ่มทำงาน ===")

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[ผิดพลาด] โหลด config: %v", err)
	}
	log.Printf("[สำเร็จ] config OK - ดึงข้อมูลทุก %d วินาที", cfg.PollSeconds)

	// เชื่อม InfluxDB
	influx := store.NewInfluxStore(
		cfg.InfluxURL, cfg.InfluxToken, cfg.InfluxOrg, cfg.InfluxBucket,
	)
	defer influx.Close()

	// Ping InfluxDB
	if err := influx.Ping(); err != nil {
		log.Printf("[เตือน] InfluxDB ไม่พร้อม - จะบันทึกเฉพาะ JSON: %v", err)
		influx = nil
	}

	// เชื่อม Solarman
	client := solarman.NewClient(cfg.BaseURL, cfg.APIID, cfg.APISecret, cfg.Email, cfg.Password)
	dataService := service.NewDataService(client, influx, cfg.OutputDir)

	// วนรอบแรกทันที แล้ววนซ้ำตาม interval
	for {
		log.Printf("[Poll] เริ่มดึงข้อมูล %s", time.Now().Format("15:04:05"))
		if err := dataService.FetchAndStore(); err != nil {
			log.Printf("[ผิดพลาด] %v", err)
		}
		log.Printf("[Poll] สำเร็จ - รอ %d วินาที...", cfg.PollSeconds)
		time.Sleep(time.Duration(cfg.PollSeconds) * time.Second)
	}
}