package service_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"kk-solamon/internal/service"
	"kk-solamon/internal/solarman"
)

// ==================== Mock: SolarmanClient ====================

type mockClient struct {
	stations     *solarman.StationListResponse
	stationsErr  error
	devices      *solarman.DeviceListResponse
	devicesErr   error
	realtime     *solarman.RealtimeDataResponse
	realtimeErr  error
}

func (m *mockClient) GetStationList(page, size int) (*solarman.StationListResponse, error) {
	return m.stations, m.stationsErr
}
func (m *mockClient) GetDeviceList(stationID int64) (*solarman.DeviceListResponse, error) {
	return m.devices, m.devicesErr
}
func (m *mockClient) GetDeviceRealtime(deviceSn string) (*solarman.RealtimeDataResponse, error) {
	return m.realtime, m.realtimeErr
}
func (m *mockClient) GetStationHistory(appID string, stationID int64, startDate, endDate string, timeType int) (*solarman.StationHistoryResponse, error) {
	return nil, nil
}

// ==================== Mock: StationStore ====================

type mockStore struct {
	stationCalls    int
	historyCalls    int
	deviceCalls     int
	writeDeviceErr  error
}

func (m *mockStore) WriteStation(s solarman.Station) error {
	m.stationCalls++
	return nil
}
func (m *mockStore) WriteStationHistoryPoint(stationID int64, name string, pt solarman.StationDataPoint) error {
	m.historyCalls++
	return nil
}
func (m *mockStore) WriteDeviceData(stationID int64, data *solarman.RealtimeDataResponse) error {
	m.deviceCalls++
	return m.writeDeviceErr
}

// ==================== Mock: APILogger ====================

type mockLogger struct {
	logs []logEntry
}
type logEntry struct {
	Endpoint  string
	StationID int64
	Status    string
	Note      string
}

func (m *mockLogger) LogAPIRequest(endpoint string, stationID int64, dateParam, status, note string) {
	m.logs = append(m.logs, logEntry{
		Endpoint:  endpoint,
		StationID: stationID,
		Status:    status,
		Note:      note,
	})
}

// ==================== Tests ====================

func TestFetchAndStore_Success(t *testing.T) {
	// จำลอง API ที่สำเร็จทุก step
	client := &mockClient{
		stations: &solarman.StationListResponse{
			Success: true,
			StationList: []solarman.Station{
				{ID: 60650830, Name: "KK-Home", GenerationPower: 1500, BatterySoc: 85},
			},
		},
		devices: &solarman.DeviceListResponse{
			Success: true,
			DeviceListItems: []solarman.Device{
				{DeviceSn: "ABC123", Status: 1},
			},
		},
		realtime: &solarman.RealtimeDataResponse{
			Success:  true,
			DeviceSn: "ABC123",
		},
	}
	store := &mockStore{}
	logger := &mockLogger{}
	tmpDir := t.TempDir()

	ds := service.NewDataService(client, store, logger, tmpDir)
	err := ds.FetchAndStore()
	if err != nil {
		t.Fatalf("FetchAndStore error: %v", err)
	}

	// ตรวจ store ถูกเรียก
	if store.stationCalls != 1 {
		t.Errorf("WriteStation: got %d calls, want 1", store.stationCalls)
	}
	if store.deviceCalls != 1 {
		t.Errorf("WriteDeviceData: got %d calls, want 1", store.deviceCalls)
	}

	// ตรวจ logger บันทึก 3 API calls (station list, device list, realtime)
	if len(logger.logs) != 3 {
		t.Errorf("LogAPIRequest: got %d logs, want 3", len(logger.logs))
	}
	for _, l := range logger.logs {
		if l.Status != "success" {
			t.Errorf("expected status=success, got %s for %s", l.Status, l.Endpoint)
		}
	}
}

func TestFetchAndStore_StationListError(t *testing.T) {
	// จำลอง API station list fail
	client := &mockClient{
		stationsErr: fmt.Errorf("API timeout"),
	}
	logger := &mockLogger{}
	tmpDir := t.TempDir()

	ds := service.NewDataService(client, nil, logger, tmpDir)
	err := ds.FetchAndStore()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// ตรวจว่า logger บันทึก error
	if len(logger.logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logger.logs))
	}
	if logger.logs[0].Status != "error" {
		t.Errorf("expected status=error, got %s", logger.logs[0].Status)
	}
}

func TestFetchAndStore_DeviceListError(t *testing.T) {
	// station list สำเร็จ แต่ device list fail (เหมือนจริง — 2101009 locked)
	client := &mockClient{
		stations: &solarman.StationListResponse{
			Success: true,
			StationList: []solarman.Station{
				{ID: 60650830, Name: "KK-Home"},
			},
		},
		devicesErr: fmt.Errorf("code=2101009 appId or api is locked"),
	}
	store := &mockStore{}
	logger := &mockLogger{}
	tmpDir := t.TempDir()

	ds := service.NewDataService(client, store, logger, tmpDir)
	err := ds.FetchAndStore()

	// FetchAndStore ไม่ return error ถ้า device list fail (แค่ข้าม)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// station ถูกเขียน แต่ device ไม่ถูกเขียน
	if store.stationCalls != 1 {
		t.Errorf("WriteStation: got %d, want 1", store.stationCalls)
	}
	if store.deviceCalls != 0 {
		t.Errorf("WriteDeviceData: got %d, want 0", store.deviceCalls)
	}

	// ตรวจ logger มี error log จาก device list
	hasError := false
	for _, l := range logger.logs {
		if l.Status == "error" && l.Endpoint == "/station/v1.0/device/list" {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected error log for device list, but not found")
	}
}

func TestFetchAndStore_NilStoreAndLogger(t *testing.T) {
	// ทดสอบ graceful nil handling — ไม่ store ไม่ logger ก็รันได้
	client := &mockClient{
		stations: &solarman.StationListResponse{
			Success:     true,
			StationList: []solarman.Station{{ID: 1, Name: "Test"}},
		},
		devices: &solarman.DeviceListResponse{
			Success:         true,
			DeviceListItems: []solarman.Device{},
		},
	}
	tmpDir := t.TempDir()

	ds := service.NewDataService(client, nil, nil, tmpDir)
	err := ds.FetchAndStore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchAndStore_SavesJSONFile(t *testing.T) {
	client := &mockClient{
		stations: &solarman.StationListResponse{
			Success:     true,
			StationList: []solarman.Station{{ID: 1, Name: "Solar1"}},
		},
		devices: &solarman.DeviceListResponse{Success: true},
	}
	tmpDir := t.TempDir()

	ds := service.NewDataService(client, nil, nil, tmpDir)
	_ = ds.FetchAndStore()

	// ตรวจ output directory มีไฟล์ JSON
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Error("expected JSON file in output dir, got 0 files")
	}
}

func TestFetchAndStore_RealtimeError(t *testing.T) {
	// processDevice: GetDeviceRealtime คืน error → early return, ไม่ WriteDeviceData
	client := &mockClient{
		stations: &solarman.StationListResponse{
			Success: true,
			StationList: []solarman.Station{
				{ID: 1, Name: "Station1"},
			},
		},
		devices: &solarman.DeviceListResponse{
			Success: true,
			DeviceListItems: []solarman.Device{
				{DeviceSn: "SN-001"},
			},
		},
		realtimeErr: fmt.Errorf("api timeout"),
	}
	store := &mockStore{}
	logger := &mockLogger{}
	tmpDir := t.TempDir()

	ds := service.NewDataService(client, store, logger, tmpDir)
	err := ds.FetchAndStore()

	// FetchAndStore ไม่ propagate error จาก processDevice
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ต้องไม่มีการ WriteDeviceData (เพราะ realtime fail)
	if store.deviceCalls != 0 {
		t.Errorf("expected 0 device writes, got %d", store.deviceCalls)
	}
	// ต้องมี log error สำหรับ realtime endpoint
	found := false
	for _, l := range logger.logs {
		if l.Endpoint == "/device/v1.0/currentData" && l.Status == "error" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error log for /device/v1.0/currentData")
	}
}

func TestFetchAndStore_WriteDeviceDataError(t *testing.T) {
	// processDevice: WriteDeviceData คืน error → log warning, ไม่ panic, FetchAndStore คืน nil
	client := &mockClient{
		stations: &solarman.StationListResponse{
			Success: true,
			StationList: []solarman.Station{
				{ID: 1, Name: "Station1"},
			},
		},
		devices: &solarman.DeviceListResponse{
			Success: true,
			DeviceListItems: []solarman.Device{
				{DeviceSn: "SN-001"},
			},
		},
		realtime: &solarman.RealtimeDataResponse{Success: true},
	}
	store := &mockStore{writeDeviceErr: fmt.Errorf("db write failed")}
	tmpDir := t.TempDir()

	ds := service.NewDataService(client, store, nil, tmpDir)
	err := ds.FetchAndStore()

	// error ถูก log ภายใน ไม่ propagate ออกมา
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// WriteDeviceData ต้องถูกเรียก 1 ครั้ง
	if store.deviceCalls != 1 {
		t.Errorf("expected 1 device write call, got %d", store.deviceCalls)
	}
}

func TestFetchAndStore_OutputDirError(t *testing.T) {
	// FetchAndStore: MkdirAll ล้มเหลวเมื่อ outputDir เป็น regular file → คืน error
	tmpDir := t.TempDir()
	blockingFile := filepath.Join(tmpDir, "notadir")
	if err := os.WriteFile(blockingFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	client := &mockClient{
		stations: &solarman.StationListResponse{Success: true},
	}
	ds := service.NewDataService(client, nil, nil, blockingFile)
	err := ds.FetchAndStore()
	if err == nil {
		t.Error("expected error when outputDir is a regular file, got nil")
	}
}

func TestSetAPIOpts_SkipStationList(t *testing.T) {
	// FETCH_STATION_LIST=false → ไม่เรียก GetStationList
	// ใช้ stationID ที่กำหนดแทน, ไม่ crash
	client := &mockClient{
		// stationsErr ยืนยันว่าถ้าโค้ดเรียก /station/v1.0/list จะ error
		// SetAPIOpts(false,...) ต้องทำให้ผ่านโดยไม่เรียก
		stationsErr: fmt.Errorf("should not call GetStationList"),
		devices:     &solarman.DeviceListResponse{Success: true},
	}
	store := &mockStore{}
	tmpDir := t.TempDir()

	ds := service.NewDataService(client, store, nil, tmpDir).
		SetAPIOpts(false, true, 999)
	err := ds.FetchAndStore()

	// ไม่ควร error (ไม่ได้เรียก API ที่จะ fail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// WriteStation ถูกเรียก 1 ครั้งจาก synthetic station ID=999
	if store.stationCalls != 1 {
		t.Errorf("expected 1 WriteStation call, got %d", store.stationCalls)
	}
}

func TestSetAPIOpts_SkipDeviceList(t *testing.T) {
	// FETCH_DEVICE_LIST=false → ไม่เรียก GetDeviceList และ GetDeviceRealtime
	client := &mockClient{
		stations: &solarman.StationListResponse{
			Success: true,
			StationList: []solarman.Station{{ID: 1, Name: "Test"}},
		},
		// devicesErr จะทำให้ fail ถ้าเรียก — ยืนยันว่าไม่เรียก
		devicesErr: fmt.Errorf("should not be called"),
	}
	store := &mockStore{}
	tmpDir := t.TempDir()

	ds := service.NewDataService(client, store, nil, tmpDir).
		SetAPIOpts(true, false, 0)
	err := ds.FetchAndStore()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// station ถูกเขียน แต่ device ไม่ถูกเขียนเพราะ skip
	if store.stationCalls != 1 {
		t.Errorf("expected 1 station write, got %d", store.stationCalls)
	}
	if store.deviceCalls != 0 {
		t.Errorf("expected 0 device writes, got %d", store.deviceCalls)
	}
}

