package main

import (
	"log"
	"kk-solamon/internal/config"
	"kk-solamon/internal/port"
	"kk-solamon/internal/service"
	"kk-solamon/internal/solarman"
	"kk-solamon/internal/store"
	"time"
)

func main() {
	log.Println("=== Solarman → InfluxDB Pipeline เริ่มทำงาน ===")

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[ผิดพลาด] โหลด config: %v", err)
	}
	log.Printf("[สำเร็จ] config OK - ดึงข้อมูลทุก %d นาที", cfg.PollMinutes)

	// เชื่อม InfluxDB
	var stationStore port.StationStore
	influx := store.NewInfluxStore(
		cfg.InfluxURL, cfg.InfluxToken, cfg.InfluxOrg, cfg.InfluxBucket,
	)
	defer influx.Close()
	if err := influx.Ping(); err != nil {
		log.Printf("[เตือน] InfluxDB ไม่พร้อม - จะบันทึกเฉพาะ JSON: %v", err)
	} else {
		stationStore = influx // กำหนดเฉพาะตอน ping สำเร็จ (nil interface ที่แท้จริง)
	}

	// เชื่อม Solarman
	client := solarman.NewClient(cfg.BaseURL, cfg.APIID, cfg.APISecret, cfg.Email, cfg.Password)

	// เชื่อม MySQL (สำหรับ log API requests)
	var apiLogger port.APILogger
	if cfg.MySQLDSN != "" {
		if ms, err := store.NewMySQLStore(cfg.MySQLDSN); err != nil {
			log.Printf("[เตือน] MySQL ไม่พร้อม - ไม่บันทึก api_request_log: %v", err)
		} else {
			defer ms.Close()
			apiLogger = ms
		}
	}

	dataService := service.NewDataService(client, stationStore, apiLogger, cfg.OutputDir).
		SetAPIOpts(cfg.FetchStationList, cfg.FetchDeviceList, cfg.StationID)

	// วนรอบแรกทันที แล้ววนซ้ำตาม interval
	for {
		log.Printf("[Poll] เริ่มดึงข้อมูล %s", time.Now().Format("15:04:05"))
		if err := dataService.FetchAndStore(); err != nil {
			log.Printf("[ผิดพลาด] %v", err)
		}
		log.Printf("[Poll] สำเร็จ - รอ %d นาที...", cfg.PollMinutes)
		time.Sleep(time.Duration(cfg.PollMinutes) * time.Minute)
	}
}