// cmd/weather-collector/main.go
// Service ดึงข้อมูลสภาพอากาศจาก Open-Meteo API (ฟรี, ไม่ต้อง API key)
// แล้วเขียนลง MySQL ทุก POLL_MINUTES
//
// ฟีเจอร์:
//   - Backfill อัตโนมัติ: ถ้า DB ว่าง → ดึงย้อนหลังตั้งแต่ WEATHER_START_DATE
//   - Poll ปกติ: ดึงทุก POLL_MINUTES (default 10 นาที)
//     → open-meteo อัปเดตทุกชั่วโมง แต่ poll บ่อยกว่าไม่เสียหาย (UPSERT)
//
// Env:
//   MYSQL_DSN        - required
//   POLL_MINUTES     - interval (default 10)
//   WEATHER_LAT      - latitude  (default 13.7563 = Bangkok)
//   WEATHER_LON      - longitude (default 100.5018)
//   WEATHER_START_DATE - วันที่เริ่ม backfill (default 2023-05-23)
//   WEATHER_BACKFILL   - "false" เพื่อปิด backfill (default true)

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"solarman-go-client/internal/config"
	"solarman-go-client/internal/store"
	"solarman-go-client/internal/weather"
)

// stationID สำหรับ KK-Home สถานีเสกา, บึงกาน (กำหนดตรง ๆ หรืออ่านจาก env)
const defaultStationID = int64(60650830)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("config:", err)
	}

	// ─── MySQL store ───────────────────────────────────────────────────────────
	mysql, err := store.NewMySQLStore(cfg.MySQLDSN)
	if err != nil {
		log.Fatal("mysql connect:", err)
	}
	log.Printf("✅ MySQL connected")

	// ─── Open-Meteo client ─────────────────────────────────────────────────────
	wClient := weather.NewClient(cfg.WeatherLat, cfg.WeatherLon)
	log.Printf("🌤 Weather client: lat=%.4f lon=%.4f", cfg.WeatherLat, cfg.WeatherLon)

	stationID := defaultStationID
	if v := os.Getenv("STATION_ID"); v != "" {
		fmt.Sscanf(v, "%d", &stationID)
	}

	// ─── Backfill: เติมข้อมูลย้อนหลังที่ขาด ──────────────────────────────────
	if cfg.WeatherBackfill {
		if err := runBackfill(mysql, wClient, stationID, cfg.WeatherStartDate); err != nil {
			log.Printf("⚠ backfill warning: %v", err)
		}
	}

	// ─── Poll loop ─────────────────────────────────────────────────────────────
	pollInterval := time.Duration(cfg.PollMinutes) * time.Minute
	log.Printf("⏱ Starting poll every %v", pollInterval)

	for {
		if err := pollCurrent(mysql, wClient, stationID); err != nil {
			log.Printf("❌ poll error: %v", err)
		}
		time.Sleep(pollInterval)
	}
}

// pollCurrent ดึงข้อมูลปัจจุบัน + 1 วันย้อนหลัง แล้วเขียน DB
func pollCurrent(mysql *store.MySQLStore, wc *weather.Client, stationID int64) error {
	records, err := wc.FetchForecast(1) // past_days=1
	if err != nil {
		return fmt.Errorf("fetch forecast: %w", err)
	}

	// ตัดเฉพาะชั่วโมงที่ไม่เกินตอนนี้
	now := time.Now()
	filtered := make([]weather.HourlyWeather, 0, len(records))
	for _, r := range records {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", r.ObservedAt, bangkokLoc())
		if err != nil || t.After(now) {
			continue
		}
		filtered = append(filtered, r)
	}

	n, err := mysql.WriteWeatherBatch(stationID, filtered)
	if err != nil {
		return fmt.Errorf("write weather: %w", err)
	}
	log.Printf("🌤 poll: wrote %d hourly records (total filtered=%d)", n, len(filtered))
	return nil
}

// runBackfill ดึงข้อมูลย้อนหลังทีละ 30 วัน จากวันที่ขาดจนถึงเมื่อวาน
func runBackfill(mysql *store.MySQLStore, wc *weather.Client, stationID int64, startDate string) error {
	latestDate, err := mysql.LatestWeatherDate(stationID)
	if err != nil {
		return fmt.Errorf("latest weather date: %w", err)
	}

	var from time.Time
	loc := bangkokLoc()

	if latestDate == "" {
		// ยังไม่มีข้อมูลเลย → เริ่มจาก startDate
		from, err = time.ParseInLocation("2006-01-02", startDate, loc)
		if err != nil {
			return fmt.Errorf("parse start date %q: %w", startDate, err)
		}
	} else {
		// มีข้อมูลแล้ว → เริ่มจากวันถัดไป
		latest, err := time.ParseInLocation("2006-01-02", latestDate, loc)
		if err != nil {
			return fmt.Errorf("parse latest date %q: %w", latestDate, err)
		}
		from = latest.AddDate(0, 0, 1)
	}

	yesterday := time.Now().In(loc).AddDate(0, 0, -1)
	if !from.Before(yesterday) {
		log.Printf("✅ Weather data up to date (latest: %s)", latestDate)
		return nil
	}

	log.Printf("📅 Backfilling weather from %s to %s ...",
		from.Format("2006-01-02"), yesterday.Format("2006-01-02"))

	total := 0
	cur := from
	for cur.Before(yesterday) {
		end := cur.AddDate(0, 0, 29) // ดึงทีละ 30 วัน
		if end.After(yesterday) {
			end = yesterday
		}

		startStr := cur.Format("2006-01-02")
		endStr := end.Format("2006-01-02")

		records, err := wc.FetchHistorical(startStr, endStr)
		if err != nil {
			log.Printf("  ⚠ backfill %s→%s: %v", startStr, endStr, err)
			cur = end.AddDate(0, 0, 1)
			continue
		}

		n, err := mysql.WriteWeatherBatch(stationID, records)
		if err != nil {
			log.Printf("  ⚠ write %s→%s: %v", startStr, endStr, err)
		}
		total += n
		log.Printf("  📥 %s → %s: %d records", startStr, endStr, n)

		cur = end.AddDate(0, 0, 1)
		time.Sleep(200 * time.Millisecond) // ไม่ spam API
	}

	log.Printf("✅ Backfill complete: %d records written", total)
	return nil
}

func bangkokLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return time.UTC
	}
	return loc
}
