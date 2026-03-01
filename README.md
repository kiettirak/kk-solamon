# ☀️ Solarman Go Client

ระบบเก็บข้อมูลโซลาร์เซลล์อัตโนมัติจาก **Solarman API** พร้อมระบบตรวจจับความผิดปกติของแผงโดยเปรียบเทียบกับข้อมูลสภาพอากาศจาก **Open-Meteo**

> **สถานี:** KK-Home | **ที่ตั้ง:** เสกา, บึงกาฬ (lat 17.9305°N, lon 103.9486°E) | **กำลังผลิต:** 5,750 W

---

## สารบัญ

1. [ภาพรวมระบบ](#1-ภาพรวมระบบ)
2. [โครงสร้างไฟล์](#2-โครงสร้างไฟล์)
3. [ข้อมูลสภาพอากาศ — ERA5 vs ECMWF คืออะไร](#3-ข้อมูลสภาพอากาศ--era5-vs-ecmwf-คืออะไร)
4. [อธิบายสูตรทั้งหมด](#4-อธิบายสูตรทั้งหมด)
5. [การตั้งค่า Quick Start](#5-การตั้งค่า-quick-start)
6. [การใช้งาน Services](#6-การใช้งาน-services)
7. [การทดสอบ](#7-การทดสอบ)
8. [MySQL Tables and Views](#8-mysql-tables-and-views)
9. [Deploy บน Proxmox](#9-deploy-บน-proxmox)
10. [ตัวแปร Environment ทั้งหมด](#10-ตัวแปร-environment-ทั้งหมด)
11. [FAQ](#11-faq)

---

## 1. ภาพรวมระบบ

```
┌─────────────────────────────────────────────────────────────┐
│                       Services (Go)                         │
│                                                             │
│  solarman-client  ──► Solarman API ──► InfluxDB + MySQL     │
│  (ทุก POLL_MINUTES)                    solar_history        │
│                                                             │
│  weather-collector ─► Open-Meteo  ──► MySQL                 │
│  (ทุก POLL_MINUTES)                    weather_history       │
│                                                             │
│  backfill (one-shot) ► Solarman API ──► MySQL (ย้อนหลัง)   │
└─────────────────────────────────────────────────────────────┘
                │                          │
                ▼                          ▼
          InfluxDB 2.7               MySQL 8.0
                │                          │
                └──────────────────────────┘
                               │
                               ▼
                         Grafana :3000

MySQL Views (วิเคราะห์อัตโนมัติ):
  solar_daily              สรุปรายวัน
  solar_hourly             power curve รายชั่วโมง
  solar_monthly            สรุปรายเดือน + รายได้
  solar_latest             real-time ล่าสุด
  solar_day_compare        วันนี้ vs เมื่อวาน vs 30d avg
  solar_performance_hourly actual vs expected kWh (Performance Ratio)
  solar_anomaly_daily      ตรวจ anomaly: แผงสกปรก/เงาบัง/inverter fault
```

---

## 2. โครงสร้างไฟล์

```
pvmon/
├── cmd/
│   ├── solarman-client/main.go     # service: poll Solarman ทุก N นาที
│   ├── weather-collector/main.go   # service: ดึงสภาพอากาศ + backfill
│   └── backfill/main.go            # one-shot: ดึงข้อมูลโซลาร์ย้อนหลัง
│
├── internal/
│   ├── config/config.go            # อ่าน env vars
│   ├── port/interfaces.go          # interfaces สำหรับ testability
│   ├── service/
│   │   ├── data_service.go         # business logic หลัก
│   │   └── data_service_test.go    # unit tests (5/5 PASS)
│   ├── solarman/
│   │   ├── auth.go                 # OAuth token management
│   │   ├── client.go               # Solarman HTTP client
│   │   ├── endpoints.go            # API endpoint constants
│   │   └── models.go               # request/response structs
│   ├── store/
│   │   ├── influxdb.go             # InfluxDB writer
│   │   ├── mysql.go                # MySQL UPSERT + API request log
│   │   └── weather_mysql.go        # weather UPSERT + batch write
│   └── weather/
│       ├── client.go               # Open-Meteo HTTP client
│       └── models.go               # weather structs + WMO code map
│
├── mysql/
│   ├── init.sql                    # tables: solar_history, api_request_log
│   ├── views.sql                   # views: solar_daily/hourly/monthly/latest/compare
│   └── weather.sql                 # weather_history + anomaly views
│
├── grafana/provisioning/
│   └── datasources/
│       ├── influxdb.yml            # auto-provision Grafana datasources
│       └── mysql.yml
├── scripts/
│   └── gen_data_dict.py            # สร้าง Data Dictionary Excel
│
├── Dockerfile                      # multi-stage build (builder + alpine runtime)
├── docker-compose.yml              # dev: influxdb + mysql + grafana
├── docker-compose.prod.yml         # prod: + solar_app + weather_collector
├── setup.sh                        # deploy script สำหรับ Ubuntu VM
└── .env                            # configuration (ดู .env.example)
```

---

## 3. ข้อมูลสภาพอากาศ — ERA5 vs ECMWF คืออะไร

### ERA5 คืออะไร?

**ERA5** ย่อมาจาก *ECMWF Reanalysis version 5* คือชุดข้อมูลสภาพอากาศย้อนหลังที่สร้างโดยการ **"วิเคราะห์ซ้ำ" (Reanalysis)** — ไม่ใช่การพยากรณ์ล่วงหน้า แต่คือการประมวลผลข้อมูลจริงย้อนหลัง จากแหล่งต่างๆ ได้แก่:

- สถานีตรวจอากาศภาคพื้นดินทั่วโลก
- ดาวเทียมสังเกตการณ์ภูมิอากาศ
- เรดาร์อากาศ
- บอลลูนตรวจอากาศ (radiosonde)
- เรือและทุ่นลอย

ข้อมูลทั้งหมดนำมาประมวลผลด้วยโมเดลคณิตศาสตร์เพื่อให้ได้ข้อมูลสภาพอากาศ **สม่ำเสมอ, ต่อเนื่อง, และครบพื้นที่ทั่วโลก**

| | ERA5 | ERA5-Land |
|---|---|---|
| ผู้พัฒนา | ECMWF (สหภาพยุโรป) | ECMWF (สหภาพยุโรป) |
| Resolution | 0.25° (~25 km) | **0.1° (~11 km)** |
| ย้อนหลังถึง | 1940 | 1950 |
| อัปเดต | ทุกวัน (delay ~5 วัน) | ทุกวัน (delay ~5 วัน) |
| จุดเด่น | ครบทุก atmospheric variable | เน้นผิวดิน เหมาะโซลาร์/เกษตร |

### ECMWF คืออะไร?

**ECMWF** (*European Centre for Medium-Range Weather Forecasts*) คือ **สถาบันพยากรณ์อากาศระดับโลกของยุโรป** ก่อตั้งปี 1975 ตั้งอยู่ที่อังกฤษ เป็นองค์กรที่**พัฒนา ERA5** และ **IFS** (Integrated Forecasting System — โมเดลพยากรณ์ล่วงหน้า)

```
ความสัมพันธ์:
ECMWF (องค์กร)
  ├── พัฒนา ERA5       ← ข้อมูลย้อนหลัง   (ระบบนี้ใช้สำหรับ backfill)
  └── พัฒนา IFS/AIFS  ← พยากรณ์ล่วงหน้า  (ระบบนี้ใช้สำหรับ real-time poll)
```

### ระบบนี้ใช้อะไร?

| สถานการณ์ | โมเดลที่ใช้ | API Endpoint |
|---|---|---|
| **Backfill ย้อนหลัง** | ERA5-Land 0.1° hourly (ตั้งแต่ 1950) | `archive-api.open-meteo.com/v1/archive` |
| **Poll ปัจจุบัน** | ECMWF IFS 9 km + ERA5 blend | `api.open-meteo.com/v1/forecast` |

ทั้งหมดผ่าน **Open-Meteo** ซึ่งเป็น wrapper ฟรี ไม่ต้อง API key

### ความแม่นยำของพิกัด (เสกา, บึงกาฬ)

| | พิกัดที่ขอ | Grid Cell ที่ Open-Meteo ใช้จริง | ห่าง |
|---|---|---|---|
| Latitude | 17.93046° N | 17.891° N | ~4.4 km |
| Longitude | 103.94865° E | 103.981° E | ~3.5 km |
| **ระยะทางรวม** | | | **~5.5 km** |

> **ทำไมถึงไม่ตรงพิกัด?** ERA5-Land มี grid cell ขนาด ~11 km (0.1 degree) ข้อมูลจะ snap ไปยัง center ของ grid cell ที่ใกล้ที่สุด ซึ่งปกติห่างไม่เกิน ~5-6 km ข้อมูลยังคง represent สภาพอากาศของพื้นที่เดียวกันได้ดี

---

## 4. อธิบายสูตรทั้งหมด

### 4.1 การประมาณพลังงาน kWh จาก Power (W)

**ปัญหา:** Solarman API ส่ง `generation_value` (kWh สะสม) เป็น `null` ทุก record → Go แปลงเป็น `0.0` อัตโนมัติ จึงต้องประมาณเองจาก Power × เวลา

**หลักการ:** พลังงาน (kWh) = กำลัง (kW) × เวลา (ชั่วโมง)

```
E (kWh) = Σ P(W) ÷ 1000 × Δt(ชั่วโมง)
        = Σ P(W) × ( 5 นาที ÷ 60 นาที/ชั่วโมง ) ÷ 1000
        = Σ P(W) × 5/60/1000
```

**ตัวอย่าง:** แผงผลิต 2,000 W เฉลี่ย 11 ชั่วโมง (132 records × 5 นาที)
```
est_solar_kwh = 2000 × 132 × 5/60/1000 = 22 kWh
```

```sql
-- ใน MySQL VIEW:
ROUND(SUM(generation_power) * 5.0 / 60.0 / 1000.0, 2) AS est_solar_kwh
```

---

### 4.2 ชั่วโมงที่โซลาร์ผลิต (Solar Hours)

```
solar_hours = (จำนวน records ที่ generation_power > 10 W) × 5 นาที ÷ 60
```

> เกณฑ์ **10 W** ใช้กรองสัญญาณรบกวน (noise) ตอนเช้ามืด/พลบค่ำ ซึ่งแผงอาจอ่านค่าเล็กน้อยโดยไม่ได้ผลิตจริง

```sql
ROUND(SUM(generation_power > 10) * 5.0 / 60.0, 1) AS solar_hours
```

---

### 4.3 Self-Consumption Rate — อัตราการใช้ไฟโซลาร์เอง

```
self_consume_pct (%) = (โซลาร์ทั้งหมด − ขายออกกริด) ÷ โซลาร์ทั้งหมด × 100
                     = (Σ generation_power − Σ grid_power) ÷ Σ generation_power × 100
```

**ตีความ:**
- `100%` = ใช้หมดเอง ไม่ขายออก (เหมาะมาก)
- `72%` = ทุก 100 หน่วยที่ผลิต ใช้เอง/ชาร์จแบต 72 หน่วย ขายออก 28 หน่วย
- `0%` = ขายออกหมด (บ้านไม่ใช้ไฟเลย)

```sql
ROUND(
    CASE WHEN SUM(generation_power) > 0
    THEN (SUM(generation_power) - SUM(grid_power)) / SUM(generation_power) * 100
    ELSE 0 END
, 1) AS self_consume_pct
```

---

### 4.4 Capacity Factor — อัตราการใช้กำลังผลิตเทียบกับ Rated

```
capacity_factor (%) = กำลังผลิตเฉลี่ยตลอด 24 ชั่วโมง ÷ กำลังติดตั้ง × 100
                    = AVG(generation_power) ÷ 5750 × 100
```

**ตีความ:** โซลาร์ในไทย ตามทฤษฎีผลิตได้ ~12 ชั่วโมง/วัน (50% ของ 24 ชั่วโมง) และมีแสงดีช่วงกลางวัน ดังนั้น capacity factor ปกติอยู่ที่ **15–22%**

```sql
ROUND(AVG(generation_power) / 5750.0 * 100.0, 1) AS capacity_factor_pct
```

---

### 4.5 ค่าไฟประหยัดโดยประมาณ (Estimated Savings)

```
est_saving_thb = พลังงานที่ใช้จากโซลาร์ (kWh) × อัตราค่าไฟ (฿/kWh)
               = SUM(use_power) × 5/60/1000 × 4.72
```

> **4.72 บาท/kWh** = อัตราค่าไฟ TOD (Time-of-Day) ช่วง peak ของ PEA (การไฟฟ้าส่วนภูมิภาค)

> **หมายเหตุ:** ค่าจริงขึ้นกับ TOU/TOD ที่ใช้ ปรับได้ใน SQL view

```sql
ROUND(SUM(use_power) * 5.0 / 60.0 / 1000.0 * 4.72, 2) AS est_saving_thb
```

---

### 4.6 รายได้จากการขายไฟ (Export Income)

```
est_export_income_thb = SUM(grid_power) × 5/60/1000 × 2.20
```

> **2.20 บาท/kWh** = อัตรา FiT (Feed-in Tariff) ของ VSPP — ผู้ผลิตไฟฟ้าจากโซลาร์ขนาดเล็กมาก ขายคืนการไฟฟ้า

```sql
ROUND(SUM(grid_power) * 5.0 / 60.0 / 1000.0 * 2.20, 0) AS est_export_income_thb
```

---

### 4.7 Expected kWh — พลังงานที่ควรผลิตได้ตามแสงอาทิตย์

**สูตรพื้นฐาน:**

```
expected_kWh = GHI (W/m²) × Prated (kWp) × η_system / 1000

โดยที่:
  GHI         = Global Horizontal Irradiance (W/m²) ← จาก Open-Meteo ERA5
  Prated      = 5.750 kWp (กำลังติดตั้ง)
  η_system    = 0.80 (ประสิทธิภาพระบบ 80%)
```

**ทำไมประสิทธิภาพระบบ = 80%?**

| ชนิด Loss | ค่าประมาณ |
|---|---|
| Inverter efficiency | 95–97% |
| Cable & connection loss | ~2% |
| Temperature derating (แผงร้อนกว่า STC 25°C อีก ~25°C → -0.45%/°C × 25) | ~89% |
| Soiling (ฝุ่นปกติ) | ~2% |
| **รวม ≈** | **~80%** |

```sql
-- ใน solar_performance_hourly VIEW:
ROUND(w.ghi_wm2 * 5.75 * 0.80 / 1000.0, 4) AS expected_kwh
```

---

### 4.8 Performance Ratio (PR) — ดัชนีวัดสุขภาพแผง

```
PR = actual_kWh ÷ expected_kWh

(คำนวณเฉพาะชั่วโมงที่ GHI > 20 W/m² เพื่อหลีกเลี่ยงชั่วโมงกลางคืน)
```

**ตีความ PR:**

| ช่วง PR | ความหมาย | การดำเนินการ |
|---|---|---|
| > 0.90 | ✅ ดีเยี่ยม | — |
| 0.75 – 0.90 | ✅ ปกติ | — |
| 0.65 – 0.75 | ⚠️ ต่ำกว่าปกติ | ติดตาม, อาจมีฝุ่น |
| < 0.65 | 🚨 ต่ำผิดปกติ | ตรวจสอบด่วน |

```sql
ROUND(
    CASE WHEN w.ghi_wm2 > 20
    THEN s.actual_kwh / NULLIF(w.ghi_wm2 * 5.75 * 0.80 / 1000.0, 0)
    ELSE NULL END
, 3) AS performance_ratio
```

---

### 4.9 PR Ratio — เปรียบ PR วันนี้กับ Baseline

```
pr_ratio = PR วันนั้น ÷ Baseline PR (เฉลี่ย 30 วันที่ผ่านมา ท้องฟ้าโปร่ง)

Baseline ใช้เฉพาะวันที่:
  - อยู่ใน 31 วันย้อนหลัง
  - avg_cloud_pct < 25% (ท้องฟ้าโปร่ง = ไม่มีตัวแปรรบกวน)
```

**ตรรกะการวินิจฉัย (Diagnosis Logic):**

```
ถ้า peak_ghi < 50               → night_or_no_data
ถ้า rain > 5mm AND cloud > 60%  → rainy_day ☁️🌧 (ไม่ใช่ปัญหาแผง)
ถ้า cloud > 60%                 → cloudy ☁️ (ไม่ใช่ปัญหาแผง)
ถ้า cloud 30-60%                → partly_cloudy ⛅
ถ้า pr_ratio < 0.50 + cloud < 25% → 🚨 anomaly: panel_issue?
ถ้า pr_ratio < 0.65 + cloud < 40% → ⚠ low_performance
ไม่เข้าเงื่อนไขใด              → normal ✅
```

> **หัวใจของระบบ:** ระบบแยกแยะ "ผลิตน้อยเพราะเมฆ" ออกจาก "ผลิตน้อยเพราะแผงมีปัญหา" โดยใช้ข้อมูลสภาพอากาศเป็น **control variable** — ถ้าท้องฟ้าโปร่ง แต่ PR ต่ำผิดปกติ แสดงว่ามีปัญหาจริง

---

### 4.10 อุณหภูมิแผงโดยประมาณ

```
est_panel_temp_c = อุณหภูมิอากาศ (°C) + 25°C
```

> กฎทั่วไป: แผงโซลาร์จะร้อนกว่าอากาศ **20–30°C** เมื่ออยู่กลางแดด (NOCT: Nominal Operating Cell Temperature ~45°C ที่ ambient 20°C, GHI 800 W/m²)

> ผลกระทบต่อประสิทธิภาพ: แผงซิลิคอนชนิด mono-PERC อุณหภูมิสูงขึ้น **1°C → ประสิทธิภาพลด ~0.35–0.45%**

---

## 5. การตั้งค่า Quick Start

### ข้อกำหนดเบื้องต้น

- Docker Desktop (Windows/Mac) หรือ Docker Engine (Linux)
- Go 1.21+ (สำหรับ development)

### ขั้นตอน

```bash
# 1. Clone
git clone https://github.com/kiettirak/pvmon.git
cd pvmon

# 2. ตั้งค่า
cp .env.example .env
# แก้ไข .env ตาม credentials จริง

# 3. เริ่ม services
docker compose up -d

# 4. ดู logs
docker logs solar_app -f
docker logs weather_collector -f
```

**URLs:**
- Grafana: http://localhost:3000 (admin / solar1234)
- InfluxDB: http://localhost:8086

---

## 6. การใช้งาน Services

### 6.1 solarman-client — Poll ข้อมูลโซลาร์ Real-time

```bash
# Development
go run ./cmd/solarman-client

# Build + run
go build -o bin/solarman-client ./cmd/solarman-client
./bin/solarman-client
```

**Flow การทำงาน:**
```
Loop ทุก POLL_MINUTES:
  1. Auth → ขอ token จาก Solarman (auto-refresh ถ้า expired)
  2. GetStationList → ดึงรายชื่อสถานีทั้งหมด
  3. สำหรับแต่ละสถานี:
     a. GetStationHistory → ดึงข้อมูล power 5 นาที (วันนี้)
     b. เขียนลง InfluxDB (time-series)
     c. เขียนลง MySQL solar_history (UPSERT)
  4. บันทึก api_request_log ทุกครั้ง (success/error)
  5. Sleep ตาม POLL_MINUTES
```

### 6.2 weather-collector — เก็บสภาพอากาศ

```bash
go run ./cmd/weather-collector

# ตั้ง env เฉพาะ:
WEATHER_LAT=17.93046324892076 WEATHER_LON=103.94864700242748 \
WEATHER_BACKFILL=true go run ./cmd/weather-collector
```

**Flow การทำงาน:**
```
Startup:
  1. ตรวจ weather_history: ล่าสุดถึงวันไหน?
  2. ถ้าขาด → Backfill ทีละ 30 วัน จาก ERA5-Land ย้อนหลัง
     (ดึงช้าๆ 200ms delay เพื่อไม่ spam Open-Meteo)
  3. บันทึก weather_history (batch UPSERT)

Loop ทุก POLL_MINUTES:
  4. FetchForecast(past_days=1) → ดึงข้อมูลวันนี้ + เมื่อวาน
  5. Filter เฉพาะชั่วโมงที่ไม่เกิน "ตอนนี้"
  6. WriteWeatherBatch → UPSERT ลง MySQL
```

### 6.3 backfill — ดึงโซลาร์ย้อนหลัง

```bash
# ย้อนหลัง 3 เดือน
go run ./cmd/backfill -start 2025-01-01 -end 2025-03-31

# ทดสอบโดยไม่เขียน DB
go run ./cmd/backfill -start 2025-06-01 -end 2025-06-30 -dry-run
```

| Flag | ความหมาย |
|---|---|
| `-start YYYY-MM-DD` | วันเริ่มต้น (required) |
| `-end YYYY-MM-DD` | วันสิ้นสุด (required) |
| `-dry-run` | แสดงผลโดยไม่เขียน DB |

---

## 7. การทดสอบ

### Unit Tests (ไม่ต้องการ DB หรือ API จริง)

```bash
# รันทั้งหมด
go test ./...

# verbose
go test -v ./internal/service/...

# coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

**Tests ที่มี (5/5 PASS):**

| Test | ทดสอบสถานการณ์ |
|---|---|
| `TestFetchAndStore_Success` | ดึงสถานีและ device สำเร็จ บันทึก JSON |
| `TestFetchAndStore_StationListError` | API คืน error → ไม่ panic, log และ return |
| `TestFetchAndStore_DeviceListError` | device API ล้มเหลว → สถานีอื่นยังทำงานต่อ |
| `TestFetchAndStore_NilStoreAndLogger` | store = nil → ระบบไม่ crash |
| `TestFetchAndStore_SavesJSONFile` | ตรวจว่าไฟล์ JSON ถูกสร้างและมีข้อมูล |

**Architecture สำหรับ Testability:**

ทุก dependency ผ่าน interface (`port/interfaces.go`) ทำให้ inject mock ได้:

```go
// interfaces.go
type SolarmanClient interface {
    GetStationList(page, size int) (*solarman.StationListResponse, error)
    GetStationHistory(...) (*solarman.StationHistoryResponse, error)
    // ...
}

type APILogger interface {
    LogAPIRequest(endpoint string, stationID int64, dateParam, status, note string)
}

type WeatherWriter interface {
    WriteWeatherBatch(stationID int64, records []weather.HourlyWeather) (int, error)
    LatestWeatherDate(stationID int64) (string, error)
}
```

### Integration Tests (ต้องมี Docker)

```bash
# 1. ตรวจ record จำนวน
docker exec solar_mysql mysql -usolar -psolar1234 solardata \
  -e "SELECT COUNT(*) as solar_records FROM solar_history;
      SELECT COUNT(*) as weather_records FROM weather_history;"

# 2. ตรวจ anomaly detection
docker exec solar_mysql mysql -usolar -psolar1234 solardata \
  -e "SELECT day, actual_kwh, avg_pr, pr_ratio, diagnosis, action
      FROM solar_anomaly_daily ORDER BY day DESC LIMIT 14;"

# 3. ตรวจ Performance Ratio รายชั่วโมง
docker exec solar_mysql mysql -usolar -psolar1234 solardata \
  -e "SELECT hour_ts, actual_kwh, expected_kwh, performance_ratio, ghi_wm2
      FROM solar_performance_hourly
      WHERE DATE(hour_ts) = CURDATE() - INTERVAL 1 DAY ORDER BY hour_ts;"

# 4. ตรวจ API request log
docker exec solar_mysql mysql -usolar -psolar1234 solardata \
  -e "SELECT requested_at, endpoint, status, note
      FROM api_request_log ORDER BY requested_at DESC LIMIT 10;"
```

### สร้าง Data Dictionary Excel

```bash
docker run --rm -v "$(pwd):/app" python:3.11-slim \
  bash -c "pip install openpyxl -q && python /app/scripts/gen_data_dict.py"
# ไฟล์: output/solar_data_dictionary.xlsx
```

---

## 8. MySQL Tables and Views

### Tables

| table | columns | คำอธิบาย |
|---|---|---|
| `solar_history` | 26 | ข้อมูล power/energy ทุก 5 นาที จาก Solarman |
| `api_request_log` | 7 | log การเรียก API ทุกครั้ง (success + error) |
| `weather_history` | 16 | สภาพอากาศรายชั่วโมง จาก Open-Meteo ERA5-Land |

### Views

| view | source tables | คำอธิบาย |
|---|---|---|
| `solar_daily` | solar_history | สรุปรายวัน: kWh, peak, solar_hours, savings, capacity factor |
| `solar_hourly` | solar_history | power curve เฉลี่ยรายชั่วโมงตลอดทั้งปี |
| `solar_monthly` | solar_history | สรุปรายเดือน: kWh, ค่าไฟ, รายได้ขายไฟ |
| `solar_latest` | solar_history | real-time ล่าสุด 1 record + battery_status |
| `solar_day_compare` | solar_daily | วันนี้ vs เมื่อวาน vs เฉลี่ย 30 วัน |
| `solar_performance_hourly` | solar_history + weather_history | PR รายชั่วโมง |
| `solar_anomaly_daily` | solar_performance_hourly | 🔍 วินิจฉัย anomaly รายวัน |

### ตรวจสอบ Views

```bash
docker exec solar_mysql mysql -usolar -psolar1234 solardata \
  -e "SHOW FULL TABLES WHERE Table_type='VIEW';"
```

---

## 9. Deploy บน Proxmox

### ข้อกำหนด
- Ubuntu 22.04+ VM (แนะนำ 2 vCPU, 4 GB RAM, 60 GB disk)
- Docker Engine + Docker Compose Plugin

### ขั้นตอน Deploy

```bash
# 1. ติดตั้ง Docker บน Ubuntu VM
curl -fsSL https://get.docker.com | sh
usermod -aG docker $USER

# 2. Clone code
git clone https://github.com/kiettirak/pvmon.git
cd pvmon

# 3. สร้าง .env.prod
cp .env.example .env.prod
nano .env.prod

# 4. Deploy
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build

# 5. ตรวจสอบ
docker compose -f docker-compose.prod.yml ps
docker logs solar_app --tail 20
docker logs weather_collector --tail 20
```

### Services ใน Production

| Container | Binary | ทำงาน |
|---|---|---|
| `solar_app` | `solarman-client` | poll Solarman API ทุก POLL_MINUTES |
| `weather_collector` | `weather-collector` | poll Open-Meteo + backfill อัตโนมัติ |
| `solar_influxdb` | InfluxDB 2.7 | time-series database |
| `solar_mysql` | MySQL 8.0 | relational database |
| `solar_grafana` | Grafana | dashboard :3000 |

### อัปเดต Code

```bash
git pull origin main
docker compose -f docker-compose.prod.yml up -d --build solar_app weather_collector
```

### Dockerfile (multi-stage)

```
Stage 1: Builder (golang:1.23-alpine)
  - Build ทุก binary พร้อมกัน: solarman-client, weather-collector, backfill

Stage 2: Runtime (alpine:3.20)
  - Copy เฉพาะ binary (ไม่มี Go toolchain)
  - Include tzdata + ca-certificates
  - Final image ขนาด ~20 MB
```

---

## 10. ตัวแปร Environment ทั้งหมด

| ตัวแปร | Default | คำอธิบาย |
|---|---|---|
| `API_ID` | — | Solarman App ID (จาก developer portal) |
| `API_SECRET` | — | Solarman App Secret |
| `BASE_URL` | `https://globalapi.solarmanpv.com` | Solarman API URL |
| `EMAIL` | — | email login Solarman |
| `PASSWORD` | — | password Solarman |
| `INFLUX_URL` | `http://localhost:8086` | InfluxDB URL |
| `INFLUX_TOKEN` | `solar-super-secret-token` | InfluxDB auth token |
| `INFLUX_ORG` | `solar-org` | InfluxDB organization |
| `INFLUX_BUCKET` | `solar-data` | InfluxDB bucket |
| `MYSQL_DSN` | — | MySQL connection string |
| `POLL_MINUTES` | `5` | interval ดึงข้อมูล Solarman (นาที) |
| `LOG_LEVEL` | `info` | ระดับ log |
| `OUTPUT_DIR` | `./output` | directory สำหรับ JSON output |
| `WEATHER_LAT` | `17.93046` | latitude สถานี (เสกา บึงกาฬ) |
| `WEATHER_LON` | `103.94865` | longitude สถานี |
| `WEATHER_START_DATE` | `2023-05-23` | วันเริ่ม backfill weather |
| `WEATHER_BACKFILL` | `true` | เปิด/ปิด backfill อัตโนมัติ |
| `STATION_ID` | `60650830` | Solarman Station ID |

---

## 11. FAQ

**Q: ทำไม generation_value, battery_soc ถึงเป็น 0 ทั้งหมด?**

A: Solarman API ส่ง `null` สำหรับ cumulative energy fields (`_value`) ใน `/station/v1.0/history` endpoint Go `float64` แปลง `null` → `0.0` โดยอัตโนมัติ ระบบจึงคำนวณ kWh ประมาณจาก power × เวลา แทน (ดูสูตร 4.1)

---

**Q: Open-Meteo ฟรีแบบไหน? มีข้อจำกัดอะไร?**

A: ฟรีสำหรับ non-commercial use, **ไม่ต้อง API key**, limit: <10,000 requests/วัน ระบบนี้ดึงวันละไม่กี่ครั้งจึงไม่มีปัญหา

---

**Q: ERA5-Land ช้าในการอัปเดตแค่ไหน?**

A: ERA5-Land อัปเดตทุกวัน แต่มี **delay ~5 วัน** สำหรับการ poll real-time ระบบใช้ Open-Meteo forecast (ECMWF IFS) แทน ซึ่งอัปเดตเกือบ real-time

---

**Q: ทำไม PR ถึงสูงกว่า 1.0 บางวัน?**

A: เป็นไปได้เพราะ: (1) ประสิทธิภาพระบบจริงสูงกว่า 80% ที่สูตรกำหนด (อากาศเย็น = แผงทำงานดี) (2) GHI จาก ERA5 อาจต่ำกว่าจริงเล็กน้อย เนื่องจาก resolution 11 km

---

**Q: แผงสกปรก vs inverter fault vs เงาบัง ดูต่างกันยังไง?**

| สถานการณ์ | Pattern |
|---|---|
| **ฝุ่น/สกปรก** | PR ค่อยๆ ลดทุกวัน → ฝนตกแล้ว PR กลับขึ้น = ยืนยันฝุ่น |
| **Inverter fault** | PR ตกทันทีวันเดียว → วันถัดไปก็ยังต่ำ |
| **เงาบัง** | PR ต่ำเฉพาะช่วงเวลาเดิมทุกวัน แต่ช่วงอื่น PR ปกติ |
| **เมฆ** | cloud_cover_pct สูง → diagnosis = cloudy/rainy (ไม่ใช่ปัญหาแผง) |

---

**Q: จะเพิ่มสถานีใหม่ได้ไหน?**

A: ได้ เพราะ `station_id` เป็น parameter ทุกที่ สถานีใหม่จะ auto-detect ผ่าน `GetStationList()` และบันทึกใน `solar_history` ด้วย `station_id` ของตัวเอง

---

## Authentication Notes

| ข้อมูล | วิธีส่ง |
|---|---|
| `appSecret` | plain text ใน JSON body |
| `password` | **SHA-256 hash** (lowercase hex) |
| Token | valid 60 วัน, auto-refresh เมื่อ expired |

---

## License

MIT License — ใช้งานได้อิสระ ทั้ง personal และ commercial
