package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"solarman-go-client/internal/config"
	"solarman-go-client/internal/solarman"
	"solarman-go-client/internal/store"
)

func main() {
	// รับ flag จาก command line
	startFlag := flag.String("start", "2023-05-23", "วันเริ่มต้น backfill (YYYY-MM-DD) — วันติดตั้ง KK-Home")
	endFlag := flag.String("end", time.Now().Format("2006-01-02"), "วันสิ้นสุด backfill (YYYY-MM-DD)")
	dryRun := flag.Bool("dry-run", false, "ถ้า true จะแสดงผลแต่ไม่เขียนลง InfluxDB")
	flag.Parse()

	startDate, err := time.Parse("2006-01-02", *startFlag)
	if err != nil {
		log.Fatalf("วันเริ่มต้นไม่ถูกต้อง: %v", err)
	}
	endDate, err := time.Parse("2006-01-02", *endFlag)
	if err != nil {
		log.Fatalf("วันสิ้นสุดไม่ถูกต้อง: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	client := solarman.NewClient(cfg.BaseURL, cfg.APIID, cfg.APISecret, cfg.Email, cfg.Password)

	var influxStore *store.InfluxStore
	var mysqlStore *store.MySQLStore
	if !*dryRun {
		influxStore = store.NewInfluxStore(cfg.InfluxURL, cfg.InfluxToken, cfg.InfluxOrg, cfg.InfluxBucket)
		defer influxStore.Close()
		if err := influxStore.Ping(); err != nil {
			log.Printf("[InfluxDB] ไม่พร้อม: %v — ข้ามการเขียน InfluxDB", err)
			influxStore = nil
		}
		if cfg.MySQLDSN != "" {
			var myErr error
			mysqlStore, myErr = store.NewMySQLStore(cfg.MySQLDSN)
			if myErr != nil {
				log.Printf("[MySQL] ไม่พร้อม: %v — ข้ามการเขียน MySQL", myErr)
				mysqlStore = nil
			} else {
				defer mysqlStore.Close()
			}
		}
	}

	stationID := int64(60650830)
	stationName := "KK-Home"

	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	log.Printf("=== Backfill: %s → %s (%d วัน) ===", *startFlag, *endFlag, totalDays)
	if *dryRun {
		log.Println("[Dry-run] ไม่เขียนลง InfluxDB")
	}

	totalPoints := 0
	current := startDate
	day := 0

	for !current.After(endDate) {
		day++
		dateStr := current.Format("2006-01-02")
		log.Printf("[%d/%d] ดึงข้อมูลวัน %s ...", day, totalDays, dateStr)

		resp, err := client.GetStationHistory(cfg.APIID, stationID, dateStr, dateStr, 1)
		if err != nil {
			log.Printf("  ⚠ ดึงไม่สำเร็จ: %v — ข้ามไป", err)
			if mysqlStore != nil {
				mysqlStore.LogAPIRequest("/station/v1.0/history", stationID, dateStr, "error", err.Error())
			}
			current = current.AddDate(0, 0, 1)
			continue
		}
		if mysqlStore != nil {
			mysqlStore.LogAPIRequest("/station/v1.0/history", stationID, dateStr, "success",
				fmt.Sprintf("records=%d", len(resp.StationDataItems)))
		}

		log.Printf("  พบ %d records", len(resp.StationDataItems))

		if !*dryRun {
			written := 0
			for _, pt := range resp.StationDataItems {
				if pt.DateTime == 0 {
					continue
				}
				if influxStore != nil {
					if err := influxStore.WriteStationHistoryPoint(stationID, stationName, pt); err != nil {
						log.Printf("  ⚠ Influx: %v", err)
					}
				}
				if mysqlStore != nil {
					if err := mysqlStore.WriteStationHistoryPoint(stationID, stationName, pt); err != nil {
						log.Printf("  ⚠ MySQL: %v", err)
					}
				}
				written++
			}
			log.Printf("  ✅ เขียน: %d points (Influx+MySQL)", written)
			totalPoints += written
		}

		// throttle: รอ 300ms ระหว่างแต่ละวัน เพื่อไม่ให้ API rate-limit
		time.Sleep(300 * time.Millisecond)
		current = current.AddDate(0, 0, 1)
	}

	fmt.Printf("\n=== Backfill เสร็จสิ้น ===\n")
	fmt.Printf("  วัน: %d\n", totalDays)
	if !*dryRun {
		fmt.Printf("  Points เขียนลง InfluxDB: %d\n", totalPoints)
	}
	fmt.Printf("========================\n")
}

