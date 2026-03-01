package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	forecastBaseURL = "https://api.open-meteo.com/v1/forecast"
	archiveBaseURL  = "https://archive-api.open-meteo.com/v1/archive"

	// ตัวแปรที่ดึง — solar radiation + cloud + ambient
	hourlyVars = "shortwave_radiation,direct_radiation,diffuse_radiation," +
		"cloud_cover,cloud_cover_low,cloud_cover_mid,cloud_cover_high," +
		"temperature_2m,precipitation,weather_code"
)

// Client ดึงข้อมูลสภาพอากาศจาก Open-Meteo API (ฟรี, ไม่ต้อง API key)
type Client struct {
	httpClient *http.Client
	Latitude   float64
	Longitude  float64
}

func NewClient(lat, lon float64) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		Latitude:   lat,
		Longitude:  lon,
	}
}

// FetchForecast ดึงข้อมูลพยากรณ์ + ย้อนหลัง pastDays วัน
// ใช้สำหรับ polling ปกติ (ดึงทุก POLL_MINUTES)
func (c *Client) FetchForecast(pastDays int) ([]HourlyWeather, error) {
	params := url.Values{
		"latitude":   {fmt.Sprintf("%.4f", c.Latitude)},
		"longitude":  {fmt.Sprintf("%.4f", c.Longitude)},
		"hourly":     {hourlyVars},
		"timezone":   {"Asia/Bangkok"},
		"past_days":  {fmt.Sprintf("%d", pastDays)},
	}
	return c.fetch(forecastBaseURL + "?" + params.Encode())
}

// FetchHistorical ดึงข้อมูลย้อนหลัง (ERA5-Land, hourly, ตั้งแต่ 1950)
// ใช้สำหรับ backfill
func (c *Client) FetchHistorical(startDate, endDate string) ([]HourlyWeather, error) {
	params := url.Values{
		"latitude":   {fmt.Sprintf("%.4f", c.Latitude)},
		"longitude":  {fmt.Sprintf("%.4f", c.Longitude)},
		"hourly":     {hourlyVars},
		"timezone":   {"Asia/Bangkok"},
		"start_date": {startDate},
		"end_date":   {endDate},
	}
	return c.fetch(archiveBaseURL + "?" + params.Encode())
}

func (c *Client) fetch(rawURL string) ([]HourlyWeather, error) {
	resp, err := c.httpClient.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("open-meteo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open-meteo HTTP %d", resp.StatusCode)
	}

	var omr OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&omr); err != nil {
		return nil, fmt.Errorf("open-meteo decode: %w", err)
	}

	return parseResponse(&omr, c.Latitude, c.Longitude)
}

// parseResponse แปลง OpenMeteoResponse → []HourlyWeather
func parseResponse(omr *OpenMeteoResponse, lat, lon float64) ([]HourlyWeather, error) {
	h := omr.Hourly
	n := len(h.Time)
	if n == 0 {
		return nil, fmt.Errorf("open-meteo: empty hourly data")
	}

	out := make([]HourlyWeather, 0, n)
	for i := 0; i < n; i++ {
		// "2023-05-23T08:00" → "2023-05-23 08:00:00"
		ts := h.Time[i]
		if len(ts) < 16 {
			continue
		}
		observed := ts[:10] + " " + ts[11:16] + ":00"

		code := safeInt(h.WeatherCode, i)
		desc, ok := WMODesc[code]
		if !ok {
			desc = fmt.Sprintf("code_%d", code)
		}

		out = append(out, HourlyWeather{
			ObservedAt:      observed,
			Latitude:        lat,
			Longitude:       lon,
			GHIWm2:          safeF(h.ShortWaveRad, i),
			DirectRadWm2:    safeF(h.DirectRad, i),
			DiffuseRadWm2:   safeF(h.DiffuseRad, i),
			CloudCoverPct:   safeF(h.CloudCover, i),
			CloudLowPct:     safeF(h.CloudCoverLow, i),
			CloudMidPct:     safeF(h.CloudCoverMid, i),
			CloudHighPct:    safeF(h.CloudCoverHigh, i),
			WeatherCode:     code,
			WeatherDesc:     desc,
			TemperatureC:    safeF(h.Temperature2m, i),
			PrecipitationMm: safeF(h.Precipitation, i),
		})
	}
	return out, nil
}

func safeF(s []float64, i int) float64 {
	if i < len(s) {
		return s[i]
	}
	return 0
}

func safeInt(s []int, i int) int {
	if i < len(s) {
		return s[i]
	}
	return 0
}
