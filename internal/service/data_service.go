package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"solarman-go-client/internal/solarman"
	"solarman-go-client/internal/store"
	"time"
)

type DataService struct {
	client      *solarman.Client
	influx      *store.InfluxStore
	outputDir   string
}

func NewDataService(client *solarman.Client, influx *store.InfluxStore, outputDir string) *DataService {
	return &DataService{
		client:    client,
		influx:    influx,
		outputDir: outputDir,
	}
}

// FetchAndStore ดึงข้อมูลทั้งหมด บันทึกลง InfluxDB + JSON
func (ds *DataService) FetchAndStore() error {
	if err := os.MkdirAll(ds.outputDir, 0755); err != nil {
		return fmt.Errorf("สร้างโฟลเดอร์ output ไม่สำเร็จ: %w", err)
	}

	// ---- Step 1: ดึงรายการโรงไฟฟ้า ----
	log.Println("[Service] กำลังดึงรายการโรงไฟฟ้า...")
	stations, err := ds.client.GetStationList(1, 20)
	if err != nil {
		return fmt.Errorf("ดึงรายการโรงไฟฟ้าไม่สำเร็จ: %w", err)
	}
	log.Printf("[Service] พบโรงไฟฟ้า %d แห่ง", len(stations.StationList))

	if err := ds.saveJSON("stations.json", stations); err != nil {
		log.Printf("[Service] เตือน: บันทึก JSON ไม่สำเร็จ: %v", err)
	}

	fmt.Printf("\n====== สรุปโรงไฟฟ้า ======\n")
	for i, s := range stations.StationList {
		fmt.Printf("%d. [ID: %d] %s | Power: %.0fW | Battery: %.1f%%\n",
			i+1, s.ID, s.Name, s.GenerationPower, s.BatterySoc)

		// เขียน station data ลง InfluxDB
		if ds.influx != nil {
			if err := ds.influx.WriteStation(s); err != nil {
				log.Printf("[Service] เตือน: เขียน station ไม่สำเร็จ: %v", err)
			}
		}

		// ---- Step 2: ดึงรายการอุปกรณ์ของแต่ละโรงไฟฟ้า ----
		devices, err := ds.client.GetDeviceList(s.ID)
		if err != nil {
			log.Printf("[Service] เตือน: ดึง device list ไม่สำเร็จ: %v", err)
			continue
		}

		for _, dev := range devices.DeviceListItems {
			// ---- Step 3: ดึงข้อมูล real-time ของแต่ละอุปกรณ์ ----
			realtimeResp, err := ds.client.GetDeviceRealtime(dev.DeviceSn)
			if err != nil {
				log.Printf("[Service] เตือน: ดึง realtime %s ไม่สำเร็จ: %v", dev.DeviceSn, err)
				continue
			}

			// เขียน device data ลง InfluxDB
			if ds.influx != nil {
				if err := ds.influx.WriteDeviceData(s.ID, realtimeResp); err != nil {
					log.Printf("[Service] เตือน: เขียน device data ไม่สำเร็จ: %v", err)
				}
			}
		}
	}
	fmt.Println("========================")
	return nil
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