package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"kk-solamon/internal/port"
	"kk-solamon/internal/solarman"
	"time"
)

// DataService orchestrate การดึงและบันทึกข้อมูล
// ใช้ interfaces จาก port/ แทน concrete types — mock ได้ใน test
type DataService struct {
	client    port.SolarmanClient // interface — mock ได้
	store     port.StationStore   // interface — mock ได้ (InfluxDB implement นี้)
	logger    port.APILogger      // interface — mock ได้ (MySQL implement นี้)
	outputDir string
}

// NewDataService สร้าง DataService โดยรับ interface แทน concrete type
// ถ้าไม่มี store หรือ logger ให้ส่ง nil — จะข้ามการบันทึกส่วนนั้น
func NewDataService(client port.SolarmanClient, store port.StationStore, logger port.APILogger, outputDir string) *DataService {
	return &DataService{
		client:    client,
		store:     store,
		logger:    logger,
		outputDir: outputDir,
	}
}

// FetchAndStore ดึงและบันทึกข้อมูลทั้งหมด 1 รอบ
func (ds *DataService) FetchAndStore() error {
	if err := os.MkdirAll(ds.outputDir, 0755); err != nil {
		return fmt.Errorf("สร้างโฟลเดอร์ output ไม่สำเร็จ: %w", err)
	}

	stations, err := ds.fetchStations()
	if err != nil {
		return err
	}

	fmt.Printf("\n====== สรุปโรงไฟฟ้า ======\n")
	for i, s := range stations.StationList {
		fmt.Printf("%d. [ID: %d] %s | Power: %.0fW | Battery: %.1f%%\n",
			i+1, s.ID, s.Name, s.GenerationPower, s.BatterySoc)
		ds.processStation(s)
	}
	fmt.Println("========================")
	return nil
}

// fetchStations ดึงรายการโรงไฟฟ้าจาก API
func (ds *DataService) fetchStations() (*solarman.StationListResponse, error) {
	log.Println("[Service] กำลังดึงรายการโรงไฟฟ้า...")
	stations, err := ds.client.GetStationList(1, 20)
	if err != nil {
		ds.log("/station/v1.0/list", 0, "", "error", err.Error())
		return nil, fmt.Errorf("ดึงรายการโรงไฟฟ้าไม่สำเร็จ: %w", err)
	}
	ds.log("/station/v1.0/list", 0, "", "success", fmt.Sprintf("stations=%d", len(stations.StationList)))
	log.Printf("[Service] พบโรงไฟฟ้า %d แห่ง", len(stations.StationList))

	if err := ds.saveJSON("stations.json", stations); err != nil {
		log.Printf("[Service] เตือน: บันทึก JSON ไม่สำเร็จ: %v", err)
	}
	return stations, nil
}

// processStation เขียนข้อมูล station + ดึง device list + realtime
func (ds *DataService) processStation(s solarman.Station) {
	// เขียน current station data
	if ds.store != nil {
		if err := ds.store.WriteStation(s); err != nil {
			log.Printf("[Service] เตือน: เขียน station ไม่สำเร็จ: %v", err)
		}
	}

	// ดึงรายการอุปกรณ์
	devices, err := ds.client.GetDeviceList(s.ID)
	if err != nil {
		ds.log("/station/v1.0/device/list", s.ID, "", "error", err.Error())
		log.Printf("[Service] เตือน: ดึง device list ไม่สำเร็จ: %v", err)
		return
	}
	ds.log("/station/v1.0/device/list", s.ID, "", "success", fmt.Sprintf("devices=%d", len(devices.DeviceListItems)))

	for _, dev := range devices.DeviceListItems {
		ds.processDevice(s.ID, dev.DeviceSn)
	}
}

// processDevice ดึงและเขียน realtime data ของ device
func (ds *DataService) processDevice(stationID int64, deviceSn string) {
	realtimeResp, err := ds.client.GetDeviceRealtime(deviceSn)
	if err != nil {
		ds.log("/device/v1.0/currentData", stationID, "", "error", err.Error())
		log.Printf("[Service] เตือน: ดึง realtime %s ไม่สำเร็จ: %v", deviceSn, err)
		return
	}
	ds.log("/device/v1.0/currentData", stationID, "", "success", fmt.Sprintf("sn=%s", deviceSn))

	if ds.store != nil {
		if err := ds.store.WriteDeviceData(stationID, realtimeResp); err != nil {
			log.Printf("[Service] เตือน: เขียน device data ไม่สำเร็จ: %v", err)
		}
	}
}

// log บันทึก API request — ข้ามถ้าไม่มี logger
func (ds *DataService) log(endpoint string, stationID int64, date, status, note string) {
	if ds.logger != nil {
		ds.logger.LogAPIRequest(endpoint, stationID, date, status, note)
	}
}

func (ds *DataService) saveJSON(filename string, data any) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("แปลงเป็น JSON ไม่สำเร็จ: %w", err)
	}
	timestampedFile := filepath.Join(
		ds.outputDir,
		fmt.Sprintf("%s_%s", time.Now().Format("20060102_150405"), filename),
	)
	if err := os.WriteFile(timestampedFile, jsonBytes, 0644); err != nil {
		return fmt.Errorf("บันทึกไฟล์ ไม่สำเร็จ: %w", err)
	}
	log.Printf("[Service] บันทึกไฟล์แล้ว: %s", timestampedFile)
	return nil
}
