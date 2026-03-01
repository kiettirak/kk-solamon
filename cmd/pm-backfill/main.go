// cmd/pm-backfill/main.go
// เติม PM2.5 / PM10 ย้อนหลังสำหรับ rows ใน weather_history ที่ยังเป็น NULL
// ดึงจาก Open-Meteo Air Quality API (CAMS reanalysis) ทีละ 30 วัน

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"kk-solamon/internal/store"
	"kk-solamon/internal/weather"
)

func main() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("MYSQL_DSN not set")
	}

	lat := 17.9305
	lon := 103.9486
	if v := os.Getenv("WEATHER_LAT"); v != "" {
		fmt.Sscanf(v, "%f", &lat)
	}
	if v := os.Getenv("WEATHER_LON"); v != "" {
		fmt.Sscanf(v, "%f", &lon)
	}

	stationID := int64(60650830)
	if v := os.Getenv("STATION_ID"); v != "" {
		fmt.Sscanf(v, "%d", &stationID)
	}

	mysql, err := store.NewMySQLStore(dsn)
	if err != nil {
		log.Fatalf("mysql connect: %v", err)
	}
	log.Printf("✅ MySQL connected")

	wc := weather.NewClient(lat, lon)

	// ─── หาช่วงวันที่ต้องเติม PM ───────────────────
	earliest, latest, err := mysql.PMNullRange(stationID)
	if err != nil {
		log.Fatalf("query pm range: %v", err)
	}
	if earliest == "" {
		log.Println("✅ ไม่มี rows ที่ขาด PM data")
		return
	}
	log.Printf("📅 PM backfill range: %s → %s", earliest, latest)

	loc, _ := time.LoadLocation("Asia/Bangkok")
	from, _ := time.ParseInLocation("2006-01-02", earliest, loc)
	to, _ := time.ParseInLocation("2006-01-02", latest, loc)

	// ─── ดึง + อัปเดตทีละ 30 วัน ──────────────────
	totalUpdated := 0
	chunkDays := 30

	for cur := from; !cur.After(to); cur = cur.AddDate(0, 0, chunkDays) {
		end := cur.AddDate(0, 0, chunkDays-1)
		if end.After(to) {
			end = to
		}
		startStr := cur.Format("2006-01-02")
		endStr := end.Format("2006-01-02")

		log.Printf("🌫 กำลังดึง PM  %s → %s ...", startStr, endStr)
		pmMap, err := wc.FetchPMRange(startStr, endStr)
		if err != nil {
			log.Printf("⚠ AQ API error (%s~%s): %v — ข้าม chunk นี้", startStr, endStr, err)
			time.Sleep(2 * time.Second)
			continue
		}

		n, err := mysql.UpdatePMBatch(stationID, pmMap)
		if err != nil {
			log.Printf("⚠ UpdatePMBatch error (%s~%s): %v", startStr, endStr, err)
			continue
		}
		totalUpdated += n
		log.Printf("   ✅ อัปเดต %d rows (chunk %s~%s)", n, startStr, endStr)

		// ป้องกัน rate-limit
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("🎉 PM backfill เสร็จสิ้น — อัปเดตรวม %d rows", totalUpdated)
}
