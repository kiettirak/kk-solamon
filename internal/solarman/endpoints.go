package solarman

const (
	// Authentication
	EndpointToken = "/account/v1.0/token"

	// Station (โรงไฟฟ้า)
	EndpointStationList = "/station/v1.0/list"
	EndpointStationInfo = "/station/v1.0/detail"

	// Device (อุปกรณ์)
	EndpointDeviceList     = "/station/v1.0/device/list" // รายการอุปกรณ์ตามโรงไฟฟ้า
	EndpointDeviceRealtime = "/device/v1.0/currentData" // ข้อมูล real-time
	EndpointDeviceHistory  = "/device/v1.0/history"     // ข้อมูลย้อนหลัง
)