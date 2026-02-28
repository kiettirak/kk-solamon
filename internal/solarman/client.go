package solarman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Client ตัวหลักสำหรับเรียก Solarman API
type Client struct {
	baseURL    string
	auth       *Auth        // pointer หา Auth struct
	httpClient *http.Client
}

// NewClient สร้าง Client ใหม่
func NewClient(baseURL, appID, appSecret, email, password string) *Client {
	return &Client{
		baseURL: baseURL,
		auth:    NewAuth(baseURL, appID, appSecret, email, password),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// post ทำ POST request แล้ว decode response เป็น struct
// ใช้ any แทน interface{} - ใน Go 1.18+ สามารถใช้แทนกันได้
func (c *Client) post(endpoint string, body any, result any) error {
	// ตะหนากว่า token ยังใช้ได้อยู่
	if err := c.auth.EnsureToken(); err != nil {
		return fmt.Errorf("ไม่สามารถเชื่อมต่อได้: %w", err)
	}

	// แปลง body เป็น JSON
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("แปลง body เป็น JSON ไม่สำเร็จ: %w", err)
	}

	url := c.baseURL + endpoint + "?language=en"
	log.Printf("[Client] POST %s", url)

	// สร้าง HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("สร้าง request ไม่สำเร็จ: %w", err)
	}

	// ใส่ header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.auth.AccessToken)

	// ส่ง request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ส่ง request ไม่สำเร็จ: %w", err)
	}
	defer resp.Body.Close()

	// ตรวจสอบ HTTP status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: status code %d", resp.StatusCode)
	}

	// อ่าน raw body เพื่อ debug แล้วค่อย decode
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("อ่าน response body ไม่สำเร็จ: %w", err)
	}
	log.Printf("[Client] Response: %s", string(rawBody))

	if err := json.Unmarshal(rawBody, result); err != nil {
		return fmt.Errorf("แปลง response ไม่สำเร็จ: %w", err)
	}

	return nil
}

// GetStationList ดึงรายการโรงไฟฟ้าทั้งหมด
func (c *Client) GetStationList(page, size int) (*StationListResponse, error) {
	var result StationListResponse
	err := c.post(EndpointStationList, StationListRequest{Page: page, Size: size}, &result)
	if err != nil {
		return nil, err
	}
	if !result.Success && result.Code != "" && result.Code != "0" {
		return nil, fmt.Errorf("API error: code=%s msg=%s", result.Code, result.Msg)
	}
	return &result, nil
}

// GetDeviceList ดึงรายการอุปกรณ์ของโรงไฟฟ้า
func (c *Client) GetDeviceList(stationID int64) (*DeviceListResponse, error) {
	var result DeviceListResponse
	err := c.post(EndpointDeviceList, DeviceListRequest{StationID: stationID, Page: 1, Size: 50}, &result)
	if err != nil {
		return nil, err
	}
	if !result.Success && result.Code != "" && result.Code != "0" {
		return nil, fmt.Errorf("API error: code=%s msg=%s", result.Code, result.Msg)
	}
	return &result, nil
}

// GetDeviceRealtime ดึงข้อมูลสดของอุปกรณ์
func (c *Client) GetDeviceRealtime(deviceSn string) (*RealtimeDataResponse, error) {
	var result RealtimeDataResponse
	err := c.post(EndpointDeviceRealtime, RealtimeDataRequest{DeviceSn: deviceSn}, &result)
	if err != nil {
		return nil, err
	}
	if !result.Success && result.Code != "" && result.Code != "0" {
		return nil, fmt.Errorf("API error: code=%s msg=%s", result.Code, result.Msg)
	}
	return &result, nil
}