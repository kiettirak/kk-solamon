package store

import (
	"database/sql"
	"fmt"
	"log"

	"solarman-go-client/internal/weather"
)

// WriteWeatherPoint เขียน 1 ชั่วโมง weather เข้า MySQL (UPSERT)
// ถ้า station_id + observed_at ซ้ำ → update ข้อมูลใหม่ทับ
func (s *MySQLStore) WriteWeatherPoint(stationID int64, w weather.HourlyWeather) error {
	query := `
INSERT INTO weather_history (
    station_id, observed_at, latitude, longitude,
    ghi_wm2, direct_rad_wm2, diffuse_rad_wm2,
    cloud_cover_pct, cloud_low_pct, cloud_mid_pct, cloud_high_pct,
    weather_code, weather_desc,
    temperature_c, precipitation_mm
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    ghi_wm2          = VALUES(ghi_wm2),
    direct_rad_wm2   = VALUES(direct_rad_wm2),
    diffuse_rad_wm2  = VALUES(diffuse_rad_wm2),
    cloud_cover_pct  = VALUES(cloud_cover_pct),
    cloud_low_pct    = VALUES(cloud_low_pct),
    cloud_mid_pct    = VALUES(cloud_mid_pct),
    cloud_high_pct   = VALUES(cloud_high_pct),
    weather_code     = VALUES(weather_code),
    weather_desc     = VALUES(weather_desc),
    temperature_c    = VALUES(temperature_c),
    precipitation_mm = VALUES(precipitation_mm),
    fetched_at       = CURRENT_TIMESTAMP`

	_, err := s.db.Exec(query,
		stationID, w.ObservedAt, w.Latitude, w.Longitude,
		w.GHIWm2, w.DirectRadWm2, w.DiffuseRadWm2,
		w.CloudCoverPct, w.CloudLowPct, w.CloudMidPct, w.CloudHighPct,
		w.WeatherCode, w.WeatherDesc,
		w.TemperatureC, w.PrecipitationMm,
	)
	return err
}

// WriteWeatherBatch เขียนหลาย records ใน transaction เดียว
func (s *MySQLStore) WriteWeatherBatch(stationID int64, records []weather.HourlyWeather) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("weather batch tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`
INSERT INTO weather_history (
    station_id, observed_at, latitude, longitude,
    ghi_wm2, direct_rad_wm2, diffuse_rad_wm2,
    cloud_cover_pct, cloud_low_pct, cloud_mid_pct, cloud_high_pct,
    weather_code, weather_desc, temperature_c, precipitation_mm
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    ghi_wm2          = VALUES(ghi_wm2),
    direct_rad_wm2   = VALUES(direct_rad_wm2),
    diffuse_rad_wm2  = VALUES(diffuse_rad_wm2),
    cloud_cover_pct  = VALUES(cloud_cover_pct),
    cloud_low_pct    = VALUES(cloud_low_pct),
    cloud_mid_pct    = VALUES(cloud_mid_pct),
    cloud_high_pct   = VALUES(cloud_high_pct),
    weather_code     = VALUES(weather_code),
    weather_desc     = VALUES(weather_desc),
    temperature_c    = VALUES(temperature_c),
    precipitation_mm = VALUES(precipitation_mm),
    fetched_at       = CURRENT_TIMESTAMP`)
	if err != nil {
		return 0, fmt.Errorf("weather batch prepare: %w", err)
	}
	defer stmt.Close()

	count := 0
	for _, w := range records {
		_, err = stmt.Exec(
			stationID, w.ObservedAt, w.Latitude, w.Longitude,
			w.GHIWm2, w.DirectRadWm2, w.DiffuseRadWm2,
			w.CloudCoverPct, w.CloudLowPct, w.CloudMidPct, w.CloudHighPct,
			w.WeatherCode, w.WeatherDesc,
			w.TemperatureC, w.PrecipitationMm,
		)
		if err != nil {
			return count, fmt.Errorf("weather insert %s: %w", w.ObservedAt, err)
		}
		count++
	}

	err = tx.Commit()
	return count, err
}

// LatestWeatherDate คืน date ล่าสุดที่มีข้อมูล weather (สำหรับ backfill gap detection)
func (s *MySQLStore) LatestWeatherDate(stationID int64) (string, error) {
	var d sql.NullString
	err := s.db.QueryRow(
		`SELECT DATE(MAX(observed_at)) FROM weather_history WHERE station_id = ?`,
		stationID,
	).Scan(&d)
	if err != nil {
		return "", err
	}
	if !d.Valid {
		return "", nil // ยังไม่มีข้อมูลเลย
	}
	return d.String, nil
}

// WeatherRecordCount จำนวน records ทั้งหมด (สำหรับ log)
func (s *MySQLStore) WeatherRecordCount(stationID int64) int64 {
	var n int64
	_ = s.db.QueryRow(
		`SELECT COUNT(*) FROM weather_history WHERE station_id = ?`,
		stationID,
	).Scan(&n)
	log.Printf("[weather] total records in DB: %d", n)
	return n
}
