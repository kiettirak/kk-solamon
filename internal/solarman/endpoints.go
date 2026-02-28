package solarman

// Endpoints ของ Solarman Global API
// ใช้ const สำหรับค่าที่ไม่เปลี่ยนแปลง (เหมือน final/const ในภาษาอื่น)
const (
	// Authentication
	EndpointToken = "/account/v1.0/token" // ขอ access token

	// Station (โรงไฟฟ้า)
	EndpointStationList = "/station/v1.0/list" // รายการโรงไฟฟ้า
	EndpointStationInfo = "/station/v1.0/detail" // ข้อมูลโรงไฟฟ้า

	// Device (อุปกรณ์)
	EndpointDeviceList    = "/device/v1.0/listByPage"  // รายการอุปกรณ์
	EndpointDeviceRealtime = "/device/v1.0/currentData" // ข้อมูล real-time
	EndpointDeviceHistory = "/device/v1.0/history"      // ข้อมูลย้อนหลัง
)