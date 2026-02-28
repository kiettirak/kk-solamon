package solarman

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Auth เก็บข้อมติเกี่ยวกับการยืนยันตัวตน
// ใน Go เราใช้ struct แทน class - struct เป็นแค่เก็บข้อมูล ไม่มี method ในตัว
type Auth struct {
	baseURL     string
	appID       string
	appSecret   string
	email       string
	password    string // plain text - จะ hash เป็น SHA256 ก่อนส่ง
	AccessToken string
	tokenExpiry time.Time
	httpClient  *http.Client
}

// NewAuth สร้าง Auth object ใหม่
// ใน Go เราใช้ constructor function แบบนี้ ไม่มี new keyword แบบ Java/C#
func NewAuth(baseURL, appID, appSecret, email, password string) *Auth {
	return &Auth{
		baseURL:   baseURL,
		appID:     appID,
		appSecret: appSecret,
		email:     email,
		password:  password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// hashSHA256 แปลง string เป็น SHA256 hash (Solarman ต้องการสำหรับ password)
func hashSHA256(text string) string {
	h := sha256.New()
	h.Write([]byte(text))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// FetchToken เรียก API เพื่อขอ access token
// ใน Go method คือ func ที่ผูกกับ struct ด้วย receiver (a *Auth)
func (a *Auth) FetchToken() error {
	// สร้าง request body ตามที่ Solarman API กำหนด
	// appSecret = plain text, password = SHA256 hash
	reqBody := TokenRequest{
		AppSecret: a.appSecret,
		Email:     a.email,
		Password:  hashSHA256(a.password),
	}

	// แปลง struct เป็น JSON bytes
	// json.Marshal เหมือน JSON.stringify() ใน JavaScript
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("แปลง request body เป็น JSON ไม่สำเร็จ: %w", err)
	}

	// สร้าง URL พร้อม query parameter appId
	url := fmt.Sprintf("%s%s?appId=%s&language=en", a.baseURL, EndpointToken, a.appID)
	log.Printf("[Auth] เรียก token จาก: %s", url)
	log.Printf("[Auth] Body: %s", string(bodyBytes))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("สร้าง request ไม่สำเร็จ: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// ส่ง request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ส่ง request ไม่สำเร็จ: %w", err)
	}
	// defer = ทำเมื่อออกจากฟังก์ชัน (เหมือน finally)
	defer resp.Body.Close()

	// แปลง JSON response เป็น struct
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("อ่าน response ไม่สำเร็จ: %w", err)
	}

	log.Printf("[Auth] Response: code=%s, success=%v, msg=%s", tokenResp.Code, tokenResp.Success, tokenResp.Msg)

	if !tokenResp.Success && tokenResp.AccessToken == "" {
		return fmt.Errorf("API error: code=%s, msg=%s", tokenResp.Code, tokenResp.Msg)
	}

	a.AccessToken = tokenResp.AccessToken
	expiry, _ := strconv.ParseInt(tokenResp.ExpiresIn, 10, 64)
	if expiry == 0 {
		expiry = 5183999 // 60 วัน (default)
	}
	a.tokenExpiry = time.Now().Add(time.Duration(expiry) * time.Second)
	log.Printf("[Auth] ได้รับ token สำเร็จ (หมดอายุ: %v)", a.tokenExpiry.Format(time.RFC3339))
	return nil
}

// IsTokenValid ตรวจสอบว่า token ยังใช้ได้อยู่ไหม
func (a *Auth) IsTokenValid() bool {
	return a.AccessToken != "" && time.Now().Before(a.tokenExpiry)
}

// EnsureToken ตรวจสอบ token และขอใหม่อัตโนมัติถ้าหมดอายุ
func (a *Auth) EnsureToken() error {
	if !a.IsTokenValid() {
		log.Println("[Auth] token หมดอายุหรือไม่มี - กำลังขอ token ใหม่")
		return a.FetchToken()
	}
	return nil
}