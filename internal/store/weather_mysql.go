package store

import (
	"database/sql"
	"fmt"
	"log"

	"kk-solamon/internal/weather"
)

// WriteWeatherPoint เขียน 1 ชั่วโมง weather เข้า MySQL (UPSERT)
// ถ้า station_id + observed_at ซ้ำ → update ข้อมูลใหม่ทับ
// หาก source เดิมเป็น "era5" แต่ใหม่เป็น "forecast" → ไม่ทับ source (era5 ดีกว่า)
func (s *MySQLStore) WriteWeatherPoint(stationID int64, w weather.HourlyWeather) error {
	src := w.Source
	if src == "" {
		src = "open-meteo"
	}
	query := `
INSERT INTO weather_history (
    station_id, observed_at, latitude, longitude,
    ghi_wm2, direct_rad_wm2, diffuse_rad_wm2,
    cloud_cover_pct, cloud_low_pct, cloud_mid_pct, cloud_high_pct,
    weather_code, weather_desc,
    temperature_c, precipitation_mm,
    pm25_ugm3, pm10_ugm3, source
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
    pm25_ugm3        = VALUES(pm25_ugm3),
    pm10_ugm3        = VALUES(pm10_ugm3),
    source           = IF(source = 'era5', 'era5', VALUES(source)),
    fetched_at       = CURRENT_TIMESTAMP`

	_, err := s.db.Exec(query,
		stationID, w.ObservedAt, w.Latitude, w.Longitude,
		w.GHIWm2, w.DirectRadWm2, w.DiffuseRadWm2,
		w.CloudCoverPct, w.CloudLowPct, w.CloudMidPct, w.CloudHighPct,
		w.WeatherCode, w.WeatherDesc,
		w.TemperatureC, w.PrecipitationMm,
		w.PM25ugm3, w.PM10ugm3, src,
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
    weather_code, weather_desc, temperature_c, precipitation_mm,
    pm25_ugm3, pm10_ugm3, source
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
    pm25_ugm3        = VALUES(pm25_ugm3),
    pm10_ugm3        = VALUES(pm10_ugm3),
    source           = IF(source = 'era5', 'era5', VALUES(source)),
    fetched_at       = CURRENT_TIMESTAMP`)
	if err != nil {
		return 0, fmt.Errorf("weather batch prepare: %w", err)
	}
	defer stmt.Close()

	count := 0
	for _, w := range records {
		src := w.Source
		if src == "" {
			src = "open-meteo"
		}
		_, err = stmt.Exec(
			stationID, w.ObservedAt, w.Latitude, w.Longitude,
			w.GHIWm2, w.DirectRadWm2, w.DiffuseRadWm2,
			w.CloudCoverPct, w.CloudLowPct, w.CloudMidPct, w.CloudHighPct,
			w.WeatherCode, w.WeatherDesc,
			w.TemperatureC, w.PrecipitationMm,
			w.PM25ugm3, w.PM10ugm3, src,
		)
		if err != nil {
			return count, fmt.Errorf("weather insert %s: %w", w.ObservedAt, err)
		}
		count++
	}

	err = tx.Commit()
	return count, err
}

// UpdatePMBatch อัปเดต pm25_ugm3 / pm10_ugm3 สำหรับ rows ที่มีอยู่แล้ว
// pmMap: key = "2023-05-23 08:00:00", value = [pm25, pm10]
// คืน จำนวน rows ที่ UPDATE สำเร็จ
func (s *MySQLStore) UpdatePMBatch(stationID int64, pmMap map[string][2]float64) (int, error) {
	if len(pmMap) == 0 {
		return 0, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("pm batch tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`
UPDATE weather_history
   SET pm25_ugm3 = ?, pm10_ugm3 = ?
 WHERE station_id = ? AND observed_at = ?`)
	if err != nil {
		return 0, fmt.Errorf("pm batch prepare: %w", err)
	}
	defer stmt.Close()

	count := 0
	for ts, pm := range pmMap {
		res, err2 := stmt.Exec(pm[0], pm[1], stationID, ts)
		if err2 != nil {
			err = err2
			return count, fmt.Errorf("pm update %s: %w", ts, err2)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			count++
		}
	}

	err = tx.Commit()
	return count, err
}

// PMNullRange คืน earliest/latest DATE ที่ pm25_ugm3 หรือ pm10_ugm3 ยังเป็น NULL
// ใช้สำหรับ pm-backfill tool
func (s *MySQLStore) PMNullRange(stationID int64) (earliest, latest string, err error) {
	var e, l sql.NullString
	err = s.db.QueryRow(`
		SELECT DATE_FORMAT(MIN(observed_at), '%Y-%m-%d'),
		       DATE_FORMAT(MAX(observed_at), '%Y-%m-%d')
		  FROM weather_history
		 WHERE station_id = ?
		   AND (pm25_ugm3 IS NULL OR pm10_ugm3 IS NULL)`,
		stationID,
	).Scan(&e, &l)
	if err != nil {
		return "", "", err
	}
	if !e.Valid {
		return "", "", nil
	}
	return e.String, l.String, nil
}

// LatestWeatherDate คืน date ล่าสุดที่มีข้อมูล weather (สำหรับ backfill gap detection)
// ใช้ DATE_FORMAT เพื่อบังคับให้ MySQL ส่งกลับเป็น string "YYYY-MM-DD"
// (หลีกเลี่ยงปัญหา parseTime=true ใน DSN ที่ทำให้ Go แปลง DATE เป็น time.Time)
func (s *MySQLStore) LatestWeatherDate(stationID int64) (string, error) {
	var d sql.NullString
	err := s.db.QueryRow(
		`SELECT DATE_FORMAT(MAX(observed_at), '%Y-%m-%d') FROM weather_history WHERE station_id = ?`,
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
