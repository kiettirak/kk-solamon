package main

import (
	"log"
	"solarman-go-client/internal/config"
	"solarman-go-client/internal/service"
	"solarman-go-client/internal/solarman"
)

func main() {
	log.Println("=== Solarman Data Fetcher เริ่มทำงาน ===")

	// Step 1: โหลด config จาก .env
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[ผิดพลาด] โหลด config ไม่สำเร็จ: %v", err)
	}
	log.Printf("[สำเร็จ] โหลด config - Base URL: %s, Email: %s", cfg.BaseURL, cfg.Email)

	// Step 2: สร้าง Solarman client (ส่ง email + password ด้วย)
	client := solarman.NewClient(cfg.BaseURL, cfg.APIID, cfg.APISecret, cfg.Email, cfg.Password)

	// Step 3: สร้าง data service
	dataService := service.NewDataService(client, cfg.OutputDir)

	// Step 4: ดึงข้อมูลและบันทึก
	if err := dataService.FetchAndSaveAll(); err != nil {
		log.Fatalf("[ผิดพลาด] ดึงข้อมูลไม่สำเร็จ: %v", err)
	}

	log.Println("=== ทำงานเสร็จสมบูรณ์ ===")
}