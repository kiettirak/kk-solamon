package weather

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	forecastBaseURL    = "https://api.open-meteo.com/v1/forecast"
	archiveBaseURL     = "https://archive-api.open-meteo.com/v1/archive"
	airQualityBaseURL  = "https://air-quality-api.open-meteo.com/v1/air-quality"

	// ตัวแปรที่ดึง — solar radiation + cloud + ambient
	hourlyVars    = "shortwave_radiation,direct_radiation,diffuse_radiation," +
		"cloud_cover,cloud_cover_low,cloud_cover_mid,cloud_cover_high," +
		"temperature_2m,precipitation,weather_code"
	airQualityVars = "pm10,pm2_5"
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
// Source จะถูก tag เป็น "forecast" (ECMWF IFS model)
func (c *Client) FetchForecast(pastDays int) ([]HourlyWeather, error) {
	params := url.Values{
		"latitude":  {fmt.Sprintf("%.4f", c.Latitude)},
		"longitude": {fmt.Sprintf("%.4f", c.Longitude)},
		"hourly":    {hourlyVars},
		"timezone":  {"Asia/Bangkok"},
		"past_days": {fmt.Sprintf("%d", pastDays)},
	}
	records, err := c.fetch(forecastBaseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}

	// ดึง PM2.5 / PM10 จาก Air Quality API แล้ว merge
	pmParams := url.Values{
		"latitude":  {fmt.Sprintf("%.4f", c.Latitude)},
		"longitude": {fmt.Sprintf("%.4f", c.Longitude)},
		"timezone":  {"Asia/Bangkok"},
		"past_days": {fmt.Sprintf("%d", pastDays)},
	}
	pmMap, pmErr := c.fetchPM(pmParams)
	if pmErr != nil {
		log.Printf("[weather] AQ forecast warning (PM data skipped): %v", pmErr)
	}

	for i := range records {
		records[i].Source = "forecast"
		if pm, ok := pmMap[records[i].ObservedAt]; ok {
			records[i].PM25ugm3 = pm[0]
			records[i].PM10ugm3 = pm[1]
		}
	}
	return records, nil
}

// FetchHistorical ดึงข้อมูลย้อนหลัง (ERA5-Land, hourly, ตั้งแต่ 1950)
// ใช้สำหรับ backfill — Source จะถูก tag เป็น "era5" (reanalysis, แม่นยำสูง)
func (c *Client) FetchHistorical(startDate, endDate string) ([]HourlyWeather, error) {
	params := url.Values{
		"latitude":   {fmt.Sprintf("%.4f", c.Latitude)},
		"longitude":  {fmt.Sprintf("%.4f", c.Longitude)},
		"hourly":     {hourlyVars},
		"timezone":   {"Asia/Bangkok"},
		"start_date": {startDate},
		"end_date":   {endDate},
	}
	records, err := c.fetch(archiveBaseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}

	// ดึง PM2.5 / PM10 จาก Air Quality API แล้ว merge
	// หมายเหตุ: CAMS reanalysis ย้อนหลังได้ตั้งแต่ 2022-07-29 เป็นต้นไป
	pmParams := url.Values{
		"latitude":   {fmt.Sprintf("%.4f", c.Latitude)},
		"longitude":  {fmt.Sprintf("%.4f", c.Longitude)},
		"timezone":   {"Asia/Bangkok"},
		"start_date": {startDate},
		"end_date":   {endDate},
	}
	pmMap, pmErr := c.fetchPM(pmParams)
	if pmErr != nil {
		log.Printf("[weather] AQ historical warning (PM data skipped): %v", pmErr)
	}

	for i := range records {
		records[i].Source = "era5"
		if pm, ok := pmMap[records[i].ObservedAt]; ok {
			records[i].PM25ugm3 = pm[0]
			records[i].PM10ugm3 = pm[1]
		}
	}
	return records, nil
}

// fetchPM ดึง PM2.5/PM10 จาก Open-Meteo Air Quality API
// คืน map[observedAt] → [pm25, pm10]
func (c *Client) fetchPM(params url.Values) (map[string][2]float64, error) {
	params.Set("hourly", airQualityVars)
	resp, err := c.httpClient.Get(airQualityBaseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("air-quality request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("air-quality HTTP %d", resp.StatusCode)
	}

	var aqr AirQualityResponse
	if err := json.NewDecoder(resp.Body).Decode(&aqr); err != nil {
		return nil, fmt.Errorf("air-quality decode: %w", err)
	}

	// build map: "2023-05-23 08:00:00" → [pm25, pm10]
	pmMap := make(map[string][2]float64, len(aqr.Hourly.Time))
	for i, ts := range aqr.Hourly.Time {
		if len(ts) < 16 {
			continue
		}
		key := ts[:10] + " " + ts[11:16] + ":00"
		pmMap[key] = [2]float64{safeF(aqr.Hourly.PM25, i), safeF(aqr.Hourly.PM10, i)}
	}
	return pmMap, nil
}

// FetchPMRange ดึง PM2.5/PM10 จาก Air Quality API สำหรับช่วงวันที่กำหนด
// คืน map[observedAt] → [pm25, pm10]  (key format: "2023-05-23 08:00:00")
func (c *Client) FetchPMRange(startDate, endDate string) (map[string][2]float64, error) {
	params := url.Values{
		"latitude":   {fmt.Sprintf("%.4f", c.Latitude)},
		"longitude":  {fmt.Sprintf("%.4f", c.Longitude)},
		"timezone":   {"Asia/Bangkok"},
		"start_date": {startDate},
		"end_date":   {endDate},
	}
	return c.fetchPM(params)
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
