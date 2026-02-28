package solarman

// ==================== Authentication ====================

// TokenRequest คือ body ที่ส่งไปขอ token
// ต้องส่ง email + password (SHA256) + appSecret
type TokenRequest struct {
	AppSecret string `json:"appSecret"`       // App Secret (plain text)
	Email     string `json:"email,omitempty"` // Email ที่ใช้ login
	Password  string `json:"password"`        // SHA256 hash ของ password
}

// TokenResponse คือ response ที่ได้จาก API
type TokenResponse struct {
	Code        string `json:"code"`         // "0" = สำเร็จ
	Msg         string `json:"msg"`          // ข้อความ
	AccessToken string `json:"access_token"` // token สำหรับเรียก API อื่น
	TokenType   string `json:"token_type"`   // "bearer"
	ExpiresIn   string `json:"expires_in"`   // อายุ token (วินาที) - API ส่งมาเป็น string
	Success     bool   `json:"success"`
}

// ==================== Station (โรงไฟฟ้า) ====================

type StationListRequest struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

type Station struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	LocationLat      float64 `json:"locationLat"`
	LocationLng      float64 `json:"locationLng"`
	InstalledCapacity float64 `json:"installedCapacity"` // กำลังติดตั้ง (W)
	GenerationPower  float64 `json:"generationPower"`   // กำลังผลิตขณะนี้ (W)
	BatterySoc       float64 `json:"batterySoc"`        // แบตเตอรี่ (%)
	NetworkStatus    string  `json:"networkStatus"`     // NORMAL, ALARM
	CreatedDate      int64   `json:"createdDate"`
	LastUpdateTime   int64   `json:"lastUpdateTime"`
}

type StationListResponse struct {
	Code        string    `json:"code"`
	Msg         string    `json:"msg"`
	Success     bool      `json:"success"`
	Total       int       `json:"total"`
	StationList []Station `json:"stationList"`
}

// ==================== Device (อุปกรณ์) ====================

type DeviceListRequest struct {
	StationID int64 `json:"stationId"`
	Page      int   `json:"page"`
	Size      int   `json:"size"`
}

type Device struct {
	DeviceSn   string `json:"deviceSn"`
	DeviceType string `json:"deviceType"`
	Status     int    `json:"deviceState"` // 1=online, 0=offline
	StationID  int64  `json:"stationId"`
}

type DeviceListResponse struct {
	Code            string   `json:"code"`
	Msg             string   `json:"msg"`
	Success         bool     `json:"success"`
	DeviceListItems []Device `json:"deviceListItems"`
}

// ==================== Real-time Data ====================

type DataAttribute struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

type RealtimeDataRequest struct {
	DeviceSn string `json:"deviceSn"`
}

type RealtimeDataResponse struct {
	Code           string          `json:"code"`
	Msg            string          `json:"msg"`
	Success        bool            `json:"success"`
	DeviceSn       string          `json:"deviceSn"`
	DeviceState    int             `json:"deviceState"`
	CollectionTime int64           `json:"collectionTime"`
	DataList       []DataAttribute `json:"dataList"`
}