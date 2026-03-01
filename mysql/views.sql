USE solardata;

-- ================================================================
-- solar_daily: สรุปรายวัน
-- คำนวณ est_kwh จาก SUM(power * 5นาที/60) / 1000
-- เพราะ API ไม่ส่ง generation_value (kWh) มาใน station history
-- ================================================================
CREATE OR REPLACE VIEW solar_daily AS
SELECT
    station_id,
    station_name,
    DATE(recorded_at)                                           AS date,

    -- พลังงานประมาณ (kWh) คำนวณจาก power × เวลา
    -- 5 นาที = 5/60 ชั่วโมง → หาร 1000 เป็น kWh
    ROUND(SUM(generation_power)    * 5.0 / 60.0 / 1000.0, 2) AS est_solar_kwh,
    ROUND(SUM(use_power)           * 5.0 / 60.0 / 1000.0, 2) AS est_use_kwh,
    ROUND(SUM(purchase_power)      * 5.0 / 60.0 / 1000.0, 2) AS est_buy_kwh,
    ROUND(SUM(grid_power)          * 5.0 / 60.0 / 1000.0, 2) AS est_export_kwh,
    ROUND(SUM(GREATEST(battery_power, 0)) * 5.0 / 60.0 / 1000.0, 2) AS est_discharge_kwh,
    ROUND(SUM(ABS(LEAST(battery_power, 0))) * 5.0 / 60.0 / 1000.0, 2) AS est_charge_kwh,

    -- Peak power
    MAX(generation_power)                                       AS peak_solar_w,
    MAX(wire_power)                                             AS peak_load_w,
    MAX(charge_power)                                           AS peak_charge_w,

    -- ชั่วโมงที่โซลาร์ผลิต (นับ 5 นาที × records ที่ > 10W)
    ROUND(SUM(generation_power > 10) * 5.0 / 60.0, 1)         AS solar_hours,

    -- สัดส่วน self-consumption: โซลาร์ที่ใช้เองใน % (ประมาณ)
    ROUND(
        CASE WHEN SUM(generation_power) > 0
        THEN (SUM(generation_power) - SUM(grid_power)) / SUM(generation_power) * 100
        ELSE 0 END
    , 1)                                                        AS self_consume_pct,

    -- Capacity factor: เปรียบกำลังเฉลี่ยกับ installed (5750W)
    ROUND(AVG(generation_power) / 5750.0 * 100.0, 1)          AS capacity_factor_pct,

    -- ค่าไฟประหยัดได้ (ประมาณ) อัตรา 4.72 บาท/หน่วย (TOD peak TOU rate)
    ROUND(SUM(use_power) * 5.0 / 60.0 / 1000.0 * 4.72, 2)    AS est_saving_thb,

    COUNT(*)                                                    AS records
FROM solar_history
GROUP BY station_id, station_name, DATE(recorded_at);


-- ================================================================
-- solar_hourly: Power Curve เฉลี่ยรายชั่วโมง (ทุกวันรวมกัน)
-- ใช้ดู pattern การผลิตว่า peak ช่วงกี่โมง
-- ================================================================
CREATE OR REPLACE VIEW solar_hourly AS
SELECT
    station_id,
    HOUR(recorded_at)               AS hour,
    ROUND(AVG(generation_power))    AS avg_solar_w,
    ROUND(AVG(use_power))           AS avg_use_w,
    ROUND(AVG(purchase_power))      AS avg_buy_w,
    ROUND(AVG(grid_power))          AS avg_export_w,
    ROUND(AVG(wire_power))          AS avg_load_w,
    ROUND(MAX(generation_power))    AS max_solar_w,
    COUNT(*)                        AS records
FROM solar_history
GROUP BY station_id, HOUR(recorded_at)
ORDER BY hour;


-- ================================================================
-- solar_monthly: สรุปรายเดือน
-- เหมาะทำ bar chart ผลิตกี่ kWh แต่ละเดือน
-- ================================================================
CREATE OR REPLACE VIEW solar_monthly AS
SELECT
    station_id,
    station_name,
    DATE_FORMAT(recorded_at, '%Y-%m')                          AS month,
    COUNT(DISTINCT DATE(recorded_at))                          AS days,

    ROUND(SUM(generation_power)    * 5.0 / 60.0 / 1000.0, 1) AS est_solar_kwh,
    ROUND(SUM(use_power)           * 5.0 / 60.0 / 1000.0, 1) AS est_use_kwh,
    ROUND(SUM(purchase_power)      * 5.0 / 60.0 / 1000.0, 1) AS est_buy_kwh,
    ROUND(SUM(grid_power)          * 5.0 / 60.0 / 1000.0, 1) AS est_export_kwh,

    ROUND(MAX(generation_power))                               AS peak_solar_w,

    -- ค่าไฟประหยัดต่อเดือน (ประมาณ)
    ROUND(SUM(use_power) * 5.0 / 60.0 / 1000.0 * 4.72, 0)    AS est_saving_thb,

    -- ค่าไฟที่ขายกริดได้ (TOD rate 2.20 บาท/หน่วย)
    ROUND(SUM(grid_power) * 5.0 / 60.0 / 1000.0 * 2.20, 0)   AS est_export_income_thb
FROM solar_history
GROUP BY station_id, station_name, DATE_FORMAT(recorded_at, '%Y-%m')
ORDER BY month;


-- ================================================================
-- solar_latest: ข้อมูล Real-time ล่าสุด
-- ใช้ทำ Stat panel (single number) ใน Grafana/Power BI
-- ================================================================
CREATE OR REPLACE VIEW solar_latest AS
SELECT
    h.station_id,
    h.station_name,
    h.recorded_at,
    h.generation_power      AS solar_w,
    h.use_power             AS consumption_w,
    h.purchase_power        AS grid_import_w,
    h.grid_power            AS grid_export_w,
    h.wire_power            AS home_load_w,
    h.battery_power         AS battery_w,        -- ลบ=ชาร์จ, บวก=คาย
    h.charge_power          AS charge_w,
    h.generation_ratio      AS capacity_pct,

    -- derived
    ROUND(h.generation_power / 5750.0 * 100.0, 1) AS capacity_factor_pct,
    CASE
        WHEN h.battery_power < -10 THEN 'charging'
        WHEN h.battery_power > 10  THEN 'discharging'
        ELSE 'idle'
    END                     AS battery_status
FROM solar_history h
INNER JOIN (
    -- ดึงเฉพาะ record ล่าสุดต่อ station
    SELECT station_id, MAX(recorded_at) AS latest
    FROM solar_history
    GROUP BY station_id
) latest ON h.station_id = latest.station_id AND h.recorded_at = latest.latest;


-- ================================================================
-- solar_day_compare: เปรียบเทียบวันนี้ vs เมื่อวาน vs เฉลี่ย 30 วัน
-- ใช้แสดงใน dashboard สรุปหน้าหลัก
-- ================================================================
CREATE OR REPLACE VIEW solar_day_compare AS
WITH daily AS (
    SELECT
        DATE(recorded_at) AS date,
        ROUND(SUM(generation_power) * 5.0/60.0/1000.0, 2) AS kwh,
        MAX(generation_power) AS peak_w
    FROM solar_history
    WHERE recorded_at >= DATE_SUB(CURDATE(), INTERVAL 32 DAY)
    GROUP BY DATE(recorded_at)
)
SELECT
    (SELECT kwh   FROM daily WHERE date = CURDATE())                               AS today_kwh,
    (SELECT peak_w FROM daily WHERE date = CURDATE())                              AS today_peak_w,
    (SELECT kwh   FROM daily WHERE date = DATE_SUB(CURDATE(), INTERVAL 1 DAY))    AS yesterday_kwh,
    (SELECT peak_w FROM daily WHERE date = DATE_SUB(CURDATE(), INTERVAL 1 DAY))   AS yesterday_peak_w,
    ROUND((SELECT AVG(kwh) FROM daily WHERE date < CURDATE()), 2)                 AS avg30d_kwh;
