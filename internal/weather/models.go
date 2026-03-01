package weather

// WMO Weather code ตาม WMO standard
var WMODesc = map[int]string{
	0:  "Clear sky",
	1:  "Mainly clear",
	2:  "Partly cloudy",
	3:  "Overcast",
	45: "Fog",
	48: "Icy fog",
	51: "Light drizzle",
	53: "Moderate drizzle",
	55: "Dense drizzle",
	61: "Slight rain",
	63: "Moderate rain",
	65: "Heavy rain",
	71: "Slight snow",
	73: "Moderate snow",
	75: "Heavy snow",
	80: "Slight showers",
	81: "Moderate showers",
	82: "Violent showers",
	95: "Thunderstorm",
	96: "Thunderstorm + hail",
	99: "Thunderstorm + heavy hail",
}

// OpenMeteoResponse — JSON response จาก Open-Meteo API (forecast + archive)
type OpenMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Hourly    struct {
		Time            []string  `json:"time"`             // "2023-05-23T08:00"
		ShortWaveRad    []float64 `json:"shortwave_radiation"`   // GHI W/m²
		DirectRad       []float64 `json:"direct_radiation"`
		DiffuseRad      []float64 `json:"diffuse_radiation"`
		CloudCover      []float64 `json:"cloud_cover"`       // %
		CloudCoverLow   []float64 `json:"cloud_cover_low"`
		CloudCoverMid   []float64 `json:"cloud_cover_mid"`
		CloudCoverHigh  []float64 `json:"cloud_cover_high"`
		Temperature2m   []float64 `json:"temperature_2m"`    // °C
		Precipitation   []float64 `json:"precipitation"`     // mm
		WeatherCode     []int     `json:"weather_code"`      // WMO code
	} `json:"hourly"`
}

// HourlyWeather — record เดียวที่จะเขียนลง DB
type HourlyWeather struct {
	StationID      int64
	ObservedAt     string  // "2023-05-23 08:00:00" Asia/Bangkok
	Latitude       float64
	Longitude      float64
	GHIWm2         float64
	DirectRadWm2   float64
	DiffuseRadWm2  float64
	CloudCoverPct  float64
	CloudLowPct    float64
	CloudMidPct    float64
	CloudHighPct   float64
	WeatherCode    int
	WeatherDesc    string
	TemperatureC   float64
	PrecipitationMm float64
}
