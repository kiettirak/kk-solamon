USE solardata;

-- ================================================================
-- weather_history: ข้อมูลสภาพอากาศรายชั่วโมง จาก Open-Meteo API
-- Source: https://archive-api.open-meteo.com / api.open-meteo.com
-- ไม่ต้อง API key (free, non-commercial, <10,000 req/day)
-- ================================================================
CREATE TABLE IF NOT EXISTS weather_history (
    id               BIGINT AUTO_INCREMENT PRIMARY KEY,
    station_id       BIGINT        NOT NULL                  COMMENT 'เชื่อมกับ solar_history.station_id',
    observed_at      DATETIME      NOT NULL                  COMMENT 'เวลาเริ่มต้นของชั่วโมง (Asia/Bangkok)',
    latitude         DOUBLE        NOT NULL                  COMMENT 'latitude ที่ใช้ query Open-Meteo',
    longitude        DOUBLE        NOT NULL                  COMMENT 'longitude ที่ใช้ query Open-Meteo',

    -- Solar Radiation (สำคัญมากสำหรับวิเคราะห์ anomaly)
    ghi_wm2          DOUBLE        DEFAULT 0                 COMMENT 'Global Horizontal Irradiance — แสงอาทิตย์รวม (W/m²)',
    direct_rad_wm2   DOUBLE        DEFAULT 0                 COMMENT 'Direct Normal Irradiance — แสงตรง (W/m²)',
    diffuse_rad_wm2  DOUBLE        DEFAULT 0                 COMMENT 'Diffuse Radiation — แสงกระจาย (W/m²)',

    -- Cloud & Weather
    cloud_cover_pct  DOUBLE        DEFAULT 0                 COMMENT 'เมฆปกคลุมทั้งหมด (%)',
    cloud_low_pct    DOUBLE        DEFAULT 0                 COMMENT 'เมฆต่ำ <3km (%)',
    cloud_mid_pct    DOUBLE        DEFAULT 0                 COMMENT 'เมฆกลาง 3-8km (%)',
    cloud_high_pct   DOUBLE        DEFAULT 0                 COMMENT 'เมฆสูง >8km (%)',
    weather_code     INT           DEFAULT 0                 COMMENT 'WMO weather code (0=clear, 1-3=partly, 45-99=bad)',
    weather_desc     VARCHAR(80)   DEFAULT ''                COMMENT 'คำอธิบาย WMO code ภาษาอังกฤษ',

    -- Ambient
    temperature_c    DOUBLE        DEFAULT 0                 COMMENT 'อุณหภูมิอากาศ 2m (°C) — ส่งผลต่อประสิทธิภาพแผง',
    precipitation_mm DOUBLE        DEFAULT 0                 COMMENT 'ปริมาณฝนชั่วโมงนั้น (mm)',

    source           VARCHAR(20)   DEFAULT 'open-meteo'      COMMENT 'แหล่งข้อมูล',
    fetched_at       DATETIME      DEFAULT CURRENT_TIMESTAMP COMMENT 'เวลาที่ดึงข้อมูล',

    INDEX  idx_station_hour (station_id, observed_at),
    UNIQUE KEY uq_station_hour (station_id, observed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='ข้อมูลสภาพอากาศรายชั่วโมง จาก Open-Meteo ERA5-Land';


-- ================================================================
-- solar_performance_hourly: เปรียบผลิตจริง vs คาดตามแสง
-- Logic: Expected_kWh = GHI(W/m²) × 5.75kWp × eff(80%) / 1000
--        Performance Ratio (PR) = actual / expected
-- ================================================================
CREATE OR REPLACE VIEW solar_performance_hourly AS
WITH hourly_solar AS (
    SELECT
        station_id,
        DATE_FORMAT(recorded_at, '%Y-%m-%d %H:00:00') AS hour_ts,
        -- 12 intervals per hour (5 min each) → หาร 12 แล้ว /1000 → kWh
        ROUND(SUM(generation_power) / 12.0 / 1000.0, 4) AS actual_kwh,
        MAX(generation_power)                            AS peak_w,
        AVG(generation_ratio)                            AS avg_gen_ratio,
        COUNT(*)                                         AS records
    FROM solar_history
    GROUP BY station_id, DATE_FORMAT(recorded_at, '%Y-%m-%d %H:00:00')
)
SELECT
    s.station_id,
    s.hour_ts,
    s.actual_kwh,
    s.peak_w,
    w.ghi_wm2,
    w.cloud_cover_pct,
    w.temperature_c,
    w.precipitation_mm,
    w.weather_code,
    w.weather_desc,

    -- Expected kWh = GHI × peak_kW × system_eff / 1000
    -- 5.75 kWp installed, system efficiency 80% (inverter+cable+temp losses)
    ROUND(w.ghi_wm2 * 5.75 * 0.80 / 1000.0, 4) AS expected_kwh,

    -- Performance Ratio (เฉพาะชั่วโมงที่มีแสงพอ GHI > 20 W/m²)
    ROUND(
        CASE WHEN w.ghi_wm2 > 20
        THEN s.actual_kwh / NULLIF(w.ghi_wm2 * 5.75 * 0.80 / 1000.0, 0)
        ELSE NULL END
    , 3) AS performance_ratio,

    -- อุณหภูมิแผงประมาณ (ปกติสูงกว่า ambient ~25°C)
    ROUND(w.temperature_c + 25.0, 1)              AS est_panel_temp_c

FROM hourly_solar s
JOIN weather_history w
  ON  s.station_id = w.station_id
  AND s.hour_ts    = DATE_FORMAT(w.observed_at, '%Y-%m-%d %H:00:00');


-- ================================================================
-- solar_anomaly_daily: ตรวจ anomaly รายวัน
-- เปรียบ PR จริง vs เฉลี่ย 30 วันล่าสุดที่ท้องฟ้าโปร่ง
-- ================================================================
CREATE OR REPLACE VIEW solar_anomaly_daily AS
WITH daily_perf AS (
    -- สรุปรายวาน: actual kWh + avg PR เฉพาะชั่วโมงที่มีแสง
    SELECT
        ph.station_id,
        DATE(ph.hour_ts)                                        AS day,
        ROUND(SUM(ph.actual_kwh), 2)                           AS actual_kwh,
        ROUND(SUM(ph.expected_kwh), 2)                         AS expected_kwh,
        ROUND(AVG(CASE WHEN ph.ghi_wm2 > 50 THEN ph.performance_ratio END), 3) AS avg_pr,
        ROUND(AVG(ph.cloud_cover_pct), 0)                      AS avg_cloud_pct,
        ROUND(MAX(ph.ghi_wm2), 0)                              AS peak_ghi_wm2,
        ROUND(SUM(ph.precipitation_mm), 1)                     AS rain_mm,
        MAX(ph.weather_code)                                   AS worst_weather_code,
        ROUND(AVG(ph.temperature_c), 1)                        AS avg_temp_c
    FROM solar_performance_hourly ph
    GROUP BY ph.station_id, DATE(ph.hour_ts)
),
baseline AS (
    -- PR เฉลี่ยจาก 30 วันที่ผ่านมาที่ท้องฟ้าโปร่ง (clear sky baseline)
    SELECT
        station_id,
        ROUND(AVG(avg_pr), 3)  AS baseline_pr_30d,
        ROUND(AVG(actual_kwh), 1) AS baseline_kwh_30d
    FROM daily_perf
    WHERE day    BETWEEN DATE_SUB(CURDATE(), INTERVAL 31 DAY)
                     AND DATE_SUB(CURDATE(), INTERVAL 1  DAY)
      AND avg_cloud_pct < 25
      AND avg_pr IS NOT NULL
    GROUP BY station_id
)
SELECT
    dp.station_id,
    dp.day,
    dp.actual_kwh,
    dp.expected_kwh,
    dp.avg_pr,
    b.baseline_pr_30d,
    dp.avg_cloud_pct,
    dp.peak_ghi_wm2,
    dp.rain_mm,
    dp.avg_temp_c,

    -- PR เทียบ baseline (1.0 = ปกติ, <0.7 = ต่ำผิดปกติ)
    ROUND(dp.avg_pr / NULLIF(b.baseline_pr_30d, 0), 2)  AS pr_ratio,

    -- วินิจฉัยสาเหตุ
    CASE
        WHEN dp.peak_ghi_wm2 < 50                              THEN 'night_or_no_data'
        WHEN dp.rain_mm > 5  AND dp.avg_cloud_pct > 60        THEN 'rainy_day ☁️🌧'
        WHEN dp.avg_cloud_pct > 60                             THEN 'cloudy ☁️'
        WHEN dp.avg_cloud_pct BETWEEN 30 AND 60               THEN 'partly_cloudy ⛅'
        WHEN dp.avg_pr IS NULL                                 THEN 'no_weather_match'
        WHEN dp.avg_pr / NULLIF(b.baseline_pr_30d, 0) < 0.50
         AND dp.avg_cloud_pct < 25                             THEN '🚨 anomaly: panel_issue?'
        WHEN dp.avg_pr / NULLIF(b.baseline_pr_30d, 0) < 0.65
         AND dp.avg_cloud_pct < 40                             THEN '⚠ low_performance'
        ELSE                                                        'normal ✅'
    END AS diagnosis,

    -- คำแนะนำ
    CASE
        WHEN dp.avg_pr / NULLIF(b.baseline_pr_30d, 0) < 0.50
         AND dp.avg_cloud_pct < 25
        THEN 'ตรวจ: สิ่งสกปรก/เงาบัง/inverter fault'
        WHEN dp.avg_pr / NULLIF(b.baseline_pr_30d, 0) < 0.65
         AND dp.avg_cloud_pct < 40
        THEN 'ติดตาม: อาจมีฝุ่นสะสม หรือ partial shading'
        ELSE ''
    END AS action

FROM daily_perf dp
LEFT JOIN baseline b ON dp.station_id = b.station_id
WHERE dp.peak_ghi_wm2 > 0
ORDER BY dp.day DESC;
