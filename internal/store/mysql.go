package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"solarman-go-client/internal/solarman"
)

// MySQLStore จัดการการเขียนข้อมูลลง MySQL
type MySQLStore struct {
	db *sql.DB
}

// NewMySQLStore สร้าง connection ไปยัง MySQL
func NewMySQLStore(dsn string) (*MySQLStore, error) {
	// dsn เช่น "solar:solar1234@tcp(localhost:3306)/solardata?parseTime=true&loc=Asia%2FBangkok"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("เปิด MySQL connection ไม่สำเร็จ: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping MySQL ไม่สำเร็จ: %w", err)
	}
	log.Println("[MySQL] เชื่อมต่อสำเร็จ")
	return &MySQLStore{db: db}, nil
}

// Close ปิด connection
func (m *MySQLStore) Close() {
	m.db.Close()
}

// WriteStationHistoryPoint เขียน 1 data point ลง MySQL
// ใช้ INSERT ... ON DUPLICATE KEY UPDATE เพื่อรองรับ re-run backfill
func (m *MySQLStore) WriteStationHistoryPoint(stationID int64, stationName string, pt solarman.StationDataPoint) error {
	if pt.DateTime == 0 {
		return nil
	}
	recordedAt := time.Unix(pt.DateTime, 0).In(bangkokTZ)

	_, err := m.db.Exec(`
		INSERT INTO solar_history (
			station_id, station_name, recorded_at,
			generation_power, battery_power, battery_soc,
			charge_power, discharge_power, grid_power,
			wire_power, use_power, purchase_power,
			generation_value, buy_value, use_value,
			charge_value, discharge_value, grid_value,
			generation_ratio, pr, cpr,
			full_power_hours, theoretical_gen,
			irradiate, irradiate_intensity
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
			generation_power    = VALUES(generation_power),
			battery_power       = VALUES(battery_power),
			battery_soc         = VALUES(battery_soc),
			charge_power        = VALUES(charge_power),
			discharge_power     = VALUES(discharge_power),
			grid_power          = VALUES(grid_power),
			wire_power          = VALUES(wire_power),
			use_power           = VALUES(use_power),
			purchase_power      = VALUES(purchase_power),
			generation_value    = VALUES(generation_value),
			buy_value           = VALUES(buy_value),
			use_value           = VALUES(use_value),
			charge_value        = VALUES(charge_value),
			discharge_value     = VALUES(discharge_value),
			grid_value          = VALUES(grid_value),
			generation_ratio    = VALUES(generation_ratio),
			pr                  = VALUES(pr),
			cpr                 = VALUES(cpr),
			full_power_hours    = VALUES(full_power_hours),
			theoretical_gen     = VALUES(theoretical_gen),
			irradiate           = VALUES(irradiate),
			irradiate_intensity = VALUES(irradiate_intensity)
	`,
		stationID, stationName, recordedAt,
		pt.GenerationPower, pt.BatteryPower, pt.BatterySoc,
		pt.ChargePower, pt.DischargePower, pt.GridPower,
		pt.WirePower, pt.UsePower, pt.PurchasePower,
		pt.GenerationValue, pt.BuyValue, pt.UseValue,
		pt.ChargeValue, pt.DischargeValue, pt.GridValue,
		pt.GenerationRatio, pt.PR, pt.CPR,
		pt.FullPowerHours, pt.TheoreticalGeneration,
		pt.Irradiate, pt.IrradiateIntensity,
	)
	if err != nil {
		return fmt.Errorf("INSERT solar_history ไม่สำเร็จ: %w", err)
	}
	return nil
}

// LogAPIRequest บันทึก 1 API request ลงตาราง api_request_log
// stationID=0 หรือ dateParam="" หมายถึงไม่มีค่า → เก็บเป็น NULL
func (m *MySQLStore) LogAPIRequest(endpoint string, stationID int64, dateParam, status, note string) {
	var sid interface{}
	if stationID != 0 {
		sid = stationID
	}
	var dp interface{}
	if dateParam != "" {
		dp = dateParam
	}
	_, err := m.db.Exec(
		`INSERT INTO api_request_log (endpoint, station_id, date_param, status, note) VALUES (?, ?, ?, ?, ?)`,
		endpoint, sid, dp, status, note,
	)
	if err != nil {
		log.Printf("[MySQL] บันทึก api_request_log ไม่สำเร็จ: %v", err)
	}
}

// bangkokTZ คือ timezone Asia/Bangkok — โหลดครั้งเดียวตอน startup
var bangkokTZ = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return time.UTC
	}
	return loc
}()
