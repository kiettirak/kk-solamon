CREATE DATABASE IF NOT EXISTS solardata CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE solardata;

CREATE TABLE IF NOT EXISTS solar_history (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    station_id      BIGINT          NOT NULL COMMENT 'Station ID จาก Solarman',
    station_name    VARCHAR(100)    NOT NULL COMMENT 'ชื่อสถานี',
    recorded_at     DATETIME        NOT NULL COMMENT 'เวลาของ data point (Asia/Bangkok)',
    -- Power (W) ณ ขณะนั้น
    generation_power    DOUBLE DEFAULT 0 COMMENT 'กำลังผลิตโซลาร์ (W)',
    battery_power       DOUBLE DEFAULT 0 COMMENT 'กำลังแบต (W, ลบ=ชาร์จ บวก=คาย)',
    battery_soc         DOUBLE DEFAULT 0 COMMENT 'ระดับแบต (%)',
    charge_power        DOUBLE DEFAULT 0 COMMENT 'กำลังชาร์จแบต (W)',
    discharge_power     DOUBLE DEFAULT 0 COMMENT 'กำลังคายแบต (W)',
    grid_power          DOUBLE DEFAULT 0 COMMENT 'export ไปกริด (W)',
    wire_power          DOUBLE DEFAULT 0 COMMENT 'โหลดบ้านทั้งหมด (W)',
    use_power           DOUBLE DEFAULT 0 COMMENT 'พลังงานที่บ้านใช้ (W)',
    purchase_power      DOUBLE DEFAULT 0 COMMENT 'ซื้อจากการไฟฟ้า (W)',
    -- Energy สะสมวันนั้น (kWh)
    generation_value    DOUBLE DEFAULT 0 COMMENT 'ผลิตสะสมวันนั้น (kWh)',
    buy_value           DOUBLE DEFAULT 0 COMMENT 'ซื้อกริดสะสม (kWh)',
    use_value           DOUBLE DEFAULT 0 COMMENT 'ใช้สะสม (kWh)',
    charge_value        DOUBLE DEFAULT 0 COMMENT 'ชาร์จแบตสะสม (kWh)',
    discharge_value     DOUBLE DEFAULT 0 COMMENT 'คายแบตสะสม (kWh)',
    grid_value          DOUBLE DEFAULT 0 COMMENT 'export กริดสะสม (kWh)',
    -- Performance
    generation_ratio    DOUBLE DEFAULT 0 COMMENT 'อัตราผลิตเทียบ capacity (%)',
    pr                  DOUBLE DEFAULT 0 COMMENT 'Performance Ratio (%)',
    cpr                 DOUBLE DEFAULT 0 COMMENT 'Capacity Performance Ratio',
    full_power_hours    DOUBLE DEFAULT 0 COMMENT 'ชั่วโมงผลิตเต็มกำลัง (h)',
    theoretical_gen     DOUBLE DEFAULT 0 COMMENT 'ผลิตตามทฤษฎี (kWh)',
    -- Irradiation
    irradiate           DOUBLE DEFAULT 0 COMMENT 'รังสีสะสม (kWh/m²)',
    irradiate_intensity DOUBLE DEFAULT 0 COMMENT 'ความเข้มแสง (W/m²)',

    INDEX idx_station_time (station_id, recorded_at),
    UNIQUE KEY uq_station_time (station_id, recorded_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='ข้อมูลประวัติโซลาร์ทุก 5 นาที จาก Solarman API';

-- ตารางบันทึกจำนวน API request
CREATE TABLE IF NOT EXISTS api_request_log (
    no              BIGINT AUTO_INCREMENT PRIMARY KEY  COMMENT 'ลำดับ (auto)',
    requested_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'เวลาที่ส่ง request',
    endpoint        VARCHAR(100) NOT NULL               COMMENT 'endpoint ที่เรียก เช่น /station/v1.0/history',
    station_id      BIGINT NULL                         COMMENT 'station ID ถ้ามี',
    date_param      DATE NULL                            COMMENT 'วันที่ที่ขอข้อมูล (สำหรับ history)',
    status          VARCHAR(20) DEFAULT 'success'       COMMENT 'success / error',
    note            VARCHAR(255) NULL                   COMMENT 'หมายเหตุ เช่น records=274 หรือ error message'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='บันทึกจำนวนและเวลาที่ request Solarman API';

-- Views (solar_daily, solar_hourly, solar_monthly, solar_latest, solar_day_compare)
-- ถูก create โดย views.sql ซึ่ง MySQL จะรันหลัง init.sql อัตโนมัติ (เรียงตาม alphabet)
