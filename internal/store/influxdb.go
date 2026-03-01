package store

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"solarman-go-client/internal/solarman"
)

// InfluxStore จัดการการเขียนข้อมูลลง InfluxDB
type InfluxStore struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	org      string
	bucket   string
}

// NewInfluxStore สร้าง InfluxStore ใหม่
func NewInfluxStore(url, token, org, bucket string) *InfluxStore {
	// สร้าง InfluxDB client
	client := influxdb2.NewClient(url, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)

	log.Printf("[InfluxDB] เชื่อมต่อกับ %s (org=%s, bucket=%s)", url, org, bucket)
	return &InfluxStore{
		client:   client,
		writeAPI: writeAPI,
		org:      org,
		bucket:   bucket,
	}
}

// Close ปิด connection
func (s *InfluxStore) Close() {
	s.client.Close()
}

// WriteStation บันทึกข้อมูลโรงไฟฟ้าลง InfluxDB
func (s *InfluxStore) WriteStation(station solarman.Station) error {
	// สร้าง data point
	p := influxdb2.NewPointWithMeasurement("solar_station").
		AddTag("station_id", fmt.Sprintf("%d", station.ID)).
		AddTag("station_name", station.Name).
		AddField("generation_power", station.GenerationPower).   // กำลังผลิตตอนนี้ (W)
		AddField("installed_capacity", station.InstalledCapacity). // กำลังติดตั้ง (W)
		AddField("battery_soc", station.BatterySoc).             // แบตเตอรี่ (%)
		SetTime(time.Now())

	if err := s.writeAPI.WritePoint(context.Background(), p); err != nil {
		return fmt.Errorf("เขียน station data ไม่สำเร็จ: %w", err)
	}

	log.Printf("[InfluxDB] บันทึก station: %s (power=%.2fW)", station.Name, station.GenerationPower)
	return nil
}

// WriteStationHistoryPoint บันทึก 1 data point จาก station history ลง InfluxDB
// ใช้ timestamp จาก dateTime ของ datapoint (ไม่ใช่ time.Now()) เพื่อให้ข้อมูลย้อนหลังถูกต้อง
func (s *InfluxStore) WriteStationHistoryPoint(stationID int64, stationName string, pt solarman.StationDataPoint) error {
	p := influxdb2.NewPointWithMeasurement("solar_station").
		AddTag("station_id", fmt.Sprintf("%d", stationID)).
		AddTag("station_name", stationName).
		// Power (W)
		AddField("generation_power", pt.GenerationPower).
		AddField("battery_power", pt.BatteryPower).
		AddField("battery_soc", pt.BatterySoc).
		AddField("charge_power", pt.ChargePower).
		AddField("discharge_power", pt.DischargePower).
		AddField("grid_power", pt.GridPower).
		AddField("wire_power", pt.WirePower).
		AddField("use_power", pt.UsePower).
		AddField("purchase_power", pt.PurchasePower).
		// Energy cumulative (kWh)
		AddField("generation_value", pt.GenerationValue).
		AddField("buy_value", pt.BuyValue).
		AddField("use_value", pt.UseValue).
		AddField("charge_value", pt.ChargeValue).
		AddField("discharge_value", pt.DischargeValue).
		AddField("grid_value", pt.GridValue).
		// Performance
		AddField("generation_ratio", pt.GenerationRatio).
		AddField("pr", pt.PR).
		AddField("cpr", pt.CPR).
		AddField("full_power_hours", pt.FullPowerHours).
		AddField("theoretical_generation", pt.TheoreticalGeneration).
		// Irradiation
		AddField("irradiate", pt.Irradiate).
		AddField("irradiate_intensity", pt.IrradiateIntensity).
		SetTime(time.Unix(pt.DateTime, 0))
	if err := s.writeAPI.WritePoint(context.Background(), p); err != nil {
		return fmt.Errorf("เขียน history point ไม่สำเร็จ: %w", err)
	}
	return nil
}

// WriteDeviceData บันทึกข้อมูล real-time อุปกรณ์ลง InfluxDB
func (s *InfluxStore) WriteDeviceData(stationID int64, data *solarman.RealtimeDataResponse) error {
	if data == nil || len(data.DataList) == 0 {
		return nil
	}

	// สร้าง point พร้อม tags
	p := influxdb2.NewPointWithMeasurement("solar_device").
		AddTag("device_sn", data.DeviceSn).
		AddTag("station_id", fmt.Sprintf("%d", stationID)).
		SetTime(time.Unix(data.CollectionTime, 0)) // ใช้เวลาจากอุปกรณ์

	// เพิ่ม fields จาก dataList ทุกตัว
	// เช่น Vpv1, Ipv1, Pac, Eday, Etotal ฯลฯ
	for _, attr := range data.DataList {
		if attr.Value != "" && attr.Value != "--" {
			p.AddField(attr.Key, attr.Value)
		}
	}

	if err := s.writeAPI.WritePoint(context.Background(), p); err != nil {
		return fmt.Errorf("เขียน device data ไม่สำเร็จ: %w", err)
	}

	log.Printf("[InfluxDB] บันทึก device: %s (%d attributes)", data.DeviceSn, len(data.DataList))
	return nil
}

// Ping ตรวจสอบว่า InfluxDB พร้อมใช้งาน
func (s *InfluxStore) Ping() error {
	ok, err := s.client.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("ping InfluxDB ไม่สำเร็จ: %w", err)
	}
	if !ok {
		return fmt.Errorf("InfluxDB ไม่ตอบสนอง")
	}
	log.Println("[InfluxDB] ping สำเร็จ - พร้อมใช้งาน")
	return nil
}
