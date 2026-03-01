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

// ==================== Station History ====================

// StationDataPoint คือข้อมูล 1 จุดใน stationDataItems (ทุก ~5 นาที)
type StationDataPoint struct {
	DateTime               int64   `json:"dateTime"`               // unix timestamp (seconds)
	Year                   int     `json:"year"`
	Month                  int     `json:"month"`
	Day                    int     `json:"day"`
	// --- Power (W) ---
	GenerationPower        float64 `json:"generationPower"`        // กำลังผลิตจากโซลาร์ (W)
	BatteryPower           float64 `json:"batteryPower"`           // กำลังแบต (ลบ=ชาร์จ, บวก=คาย)
	BatterySoc             float64 `json:"batterySoc"`             // ระดับแบต (%)
	ChargePower            float64 `json:"chargePower"`            // กำลังชาร์จแบต (W)
	DischargePower         float64 `json:"dischargePower"`         // กำลังคายแบต (W)
	GridPower              float64 `json:"gridPower"`              // export ไปกริด (W)
	WirePower              float64 `json:"wirePower"`              // โหลดบ้านทั้งหมด (W)
	UsePower               float64 `json:"usePower"`               // พลังงานที่บ้านใช้ (W)
	PurchasePower          float64 `json:"purchasePower"`          // ซื้อจากการไฟฟ้า (W)
	// --- Energy cumulative (kWh) ---
	GenerationValue        float64 `json:"generationValue"`        // ผลิตสะสมวันนั้น (kWh)
	BuyValue               float64 `json:"buyValue"`               // ซื้อจากกริดสะสม (kWh)
	UseValue               float64 `json:"useValue"`               // ใช้สะสม (kWh)
	ChargeValue            float64 `json:"chargeValue"`            // ชาร์จแบตสะสม (kWh)
	DischargeValue         float64 `json:"dischargeValue"`         // คายแบตสะสม (kWh)
	GridValue              float64 `json:"gridValue"`              // export กริดสะสม (kWh)
	// --- Performance ---
	GenerationRatio        float64 `json:"generationRatio"`        // อัตราผลิตเทียบ capacity (%)
	PR                     float64 `json:"pr"`                     // Performance Ratio (%)
	CPR                    float64 `json:"cpr"`                    // Capacity Performance Ratio
	FullPowerHours         float64 `json:"fullPowerHours"`         // ชั่วโมงผลิตเต็มกำลัง (h)
	TheoreticalGeneration  float64 `json:"theoreticalGeneration"`  // ผลิตตามทฤษฎี (kWh)
	// --- Solar irradiation ---
	Irradiate              float64 `json:"irradiate"`              // รังสีสะสม (kWh/m²)
	IrradiateIntensity     float64 `json:"irradiateIntensity"`     // ความเข้มแสง (W/m²)
}

// StationHistoryResponse คือ response จาก /station/v1.0/history
type StationHistoryResponse struct {
	Code             interface{}        `json:"code"`
	Msg              interface{}        `json:"msg"`
	Success          bool               `json:"success"`
	Total            int                `json:"total"`
	StationDataItems []StationDataPoint `json:"stationDataItems"`
}