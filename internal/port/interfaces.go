// Package port กำหนด interfaces (ports) ที่แยก business logic
// ออกจาก implementation จริง — ทำให้ test และเปลี่ยน implementation ได้ง่าย
package port

import (
	"kk-solamon/internal/solarman"
	weatherpkg "kk-solamon/internal/weather"
)

// SolarmanClient interface สำหรับ Solarman API
// ทำให้ mock ได้ใน unit test โดยไม่ต้องเรียก API จริง
type SolarmanClient interface {
	GetStationList(page, size int) (*solarman.StationListResponse, error)
	GetDeviceList(stationID int64) (*solarman.DeviceListResponse, error)
	GetDeviceRealtime(deviceSn string) (*solarman.RealtimeDataResponse, error)
	GetStationHistory(appID string, stationID int64, startDate, endDate string, timeType int) (*solarman.StationHistoryResponse, error)
}

// StationWriter interface สำหรับเขียน station current data
type StationWriter interface {
	WriteStation(station solarman.Station) error
}

// HistoryWriter interface สำหรับเขียน historical data points
type HistoryWriter interface {
	WriteStationHistoryPoint(stationID int64, stationName string, pt solarman.StationDataPoint) error
}

// DeviceWriter interface สำหรับเขียน device realtime data
type DeviceWriter interface {
	WriteDeviceData(stationID int64, data *solarman.RealtimeDataResponse) error
}

// StationStore รวม StationWriter + HistoryWriter + DeviceWriter
// ใช้เมื่อ store รองรับทั้ง 3 อย่าง (เช่น InfluxDB)
type StationStore interface {
	StationWriter
	HistoryWriter
	DeviceWriter
}

// APILogger interface สำหรับบันทึก API request log
type APILogger interface {
	LogAPIRequest(endpoint string, stationID int64, dateParam, status, note string)
}

// WeatherWriter interface สำหรับเขียนข้อมูลสภาพอากาศลง DB
type WeatherWriter interface {
	WriteWeatherBatch(stationID int64, records []weatherpkg.HourlyWeather) (int, error)
	LatestWeatherDate(stationID int64) (string, error)
}
