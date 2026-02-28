package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"solarman-go-client/internal/solarman"
	"time"
)

// DataService จัดการเรียก API และบันทึกข้อมูล
type DataService struct {
	client    *solarman.Client
	outputDir string
}

// NewDataService สร้าง DataService ใหม่
func NewDataService(client *solarman.Client, outputDir string) *DataService {
	return &DataService{
		client:    client,
		outputDir: outputDir,
	}
}

// FetchAndSaveAll ดึงข้อมูลทั้งหมดแล้วบันทึกเป็น JSON
func (ds *DataService) FetchAndSaveAll() error {
	// สร้างโฟลเดอร์ output ถ้ายังไม่มี
	// os.MkdirAll เหมือน mkdir -p
	if err := os.MkdirAll(ds.outputDir, 0755); err != nil {
		return fmt.Errorf("สร้างโฟลเดอร์ output ไม่สำเร็จ: %w", err)
	}

	// ---- Step 1: ดึงรายการโรงไฟฟ้า ----
	log.Println("[Service] กำลังดึงรายการโรงไฟฟ้า...")
	stations, err := ds.client.GetStationList(1, 20)
	if err != nil {
		return fmt.Errorf("ดึงรายการโรงไฟฟ้าไม่สำเร็จ: %w", err)
	}
	log.Printf("[Service] พบโรงไฟฟ้าทั้งหมด %d แห่ง", len(stations.StationList))

	// บันทึก stations เป็น JSON file
	if err := ds.saveJSON("stations.json", stations); err != nil {
		return err
	}

	// สรุปให้ดูใน console
	fmt.Printf("\n========== สรุปโรงไฟฟ้า ==========\n")
	for i, s := range stations.StationList {
		fmt.Printf("%d. [ID: %d] %s (%.2f kW)\n", i+1, s.ID, s.Name, s.InstalledPower)
	}
	fmt.Println("================================\n")

	return nil
}

// saveJSON บันทึกข้อมูลเป็น JSON file
func (ds *DataService) saveJSON(filename string, data any) error {
	// สร้าง path เต็ม เช่น ./output/stations.json
	filePath := filepath.Join(ds.outputDir, filename)

	// json.MarshalIndent ทำให้ JSON อ่านง่าย (pretty print)
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("แปลงเป็น JSON ไม่สำเร็จ: %w", err)
	}

	// os.WriteFile เขียนไฟล์ ชื่อไฟล์ timestamp เพื่อไม่ทับกัน
	timestampedFile := filepath.Join(
		ds.outputDir,
		fmt.Sprintf("%s_%s", time.Now().Format("20060102_150405"), filename),
	)
	if err := os.WriteFile(timestampedFile, jsonBytes, 0644); err != nil {
		return fmt.Errorf("บันทึกไฟล์ %s ไม่สำเร็จ: %w", filePath, err)
	}

	log.Printf("[Service] บันทึกไฟล์แล้ว: %s", timestampedFile)
	return nil
}