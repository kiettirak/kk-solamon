# -*- coding: utf-8 -*-
"""
สร้าง Data Dictionary Excel สำหรับ Solar Monitoring System
"""
import openpyxl
from openpyxl.styles import (
    Font, PatternFill, Alignment, Border, Side, GradientFill
)
from openpyxl.utils import get_column_letter

wb = openpyxl.Workbook()
wb.remove(wb.active)  # ลบ sheet เปล่า

# ─── Color Palette ───────────────────────────────────────────────────────────
CLR_TABLE_HDR   = "1F4E79"   # navy — header tables
CLR_VIEW_HDR    = "375623"   # dark green — header views
CLR_COL_HDR     = "BDD7EE"   # light blue — column header row
CLR_COL_HDR_V   = "E2EFDA"   # light green — column header row (view)
CLR_ROW_ODD     = "F7FBFF"
CLR_ROW_ODD_V   = "F6FFED"
CLR_ROW_EVEN    = "FFFFFF"
CLR_PK          = "FFE699"   # yellow — primary key
CLR_FK          = "FCE4D6"   # orange — foreign key
CLR_SECTION     = "D9E2F3"   # section separator

FONT_TITLE  = Font(name="Calibri", bold=True, size=14, color="FFFFFF")
FONT_HEADER = Font(name="Calibri", bold=True, size=10, color="1F4E79")
FONT_NORMAL = Font(name="Calibri", size=10)
FONT_MONO   = Font(name="Consolas", size=9, color="1F3864")
FONT_NOTE   = Font(name="Calibri", size=9, italic=True, color="595959")

thin  = Side(style="thin",   color="BFBFBF")
thick = Side(style="medium", color="808080")
BORDER_THIN   = Border(left=thin,  right=thin,  top=thin,  bottom=thin)
BORDER_HEADER = Border(left=thick, right=thick, top=thick, bottom=thick)

CENTER  = Alignment(horizontal="center", vertical="center", wrap_text=True)
LEFT    = Alignment(horizontal="left",   vertical="center", wrap_text=True)
LEFT_NW = Alignment(horizontal="left",   vertical="top",    wrap_text=True)


def fill(hex_color):
    return PatternFill("solid", fgColor=hex_color)


def write_header_row(ws, row, cols, is_view=False):
    bg  = CLR_COL_HDR_V if is_view else CLR_COL_HDR
    for col_idx, text in enumerate(cols, 1):
        c = ws.cell(row=row, column=col_idx, value=text)
        c.font      = Font(name="Calibri", bold=True, size=10, color="1F4E79" if not is_view else "375623")
        c.fill      = fill(bg)
        c.alignment = CENTER
        c.border    = BORDER_THIN


def write_data_row(ws, row, values, is_pk=False, is_view=False, odd=True):
    bg = (CLR_PK if is_pk
          else (CLR_ROW_ODD_V if (is_view and odd) else CLR_ROW_ODD) if odd
          else CLR_ROW_EVEN)
    for col_idx, val in enumerate(values, 1):
        c = ws.cell(row=row, column=col_idx, value=val)
        c.font      = FONT_MONO if col_idx in (2, 3) else FONT_NORMAL
        c.fill      = fill(bg)
        c.alignment = CENTER if col_idx in (3, 4, 5) else LEFT
        c.border    = BORDER_THIN


def write_title_block(ws, title, subtitle, hex_bg):
    ws.merge_cells("A1:I1")
    c = ws["A1"]
    c.value     = title
    c.font      = FONT_TITLE
    c.fill      = fill(hex_bg)
    c.alignment = CENTER

    ws.merge_cells("A2:I2")
    c = ws["A2"]
    c.value     = subtitle
    c.font      = Font(name="Calibri", italic=True, size=10, color="FFFFFF")
    c.fill      = fill(hex_bg)
    c.alignment = CENTER
    ws.row_dimensions[1].height = 28
    ws.row_dimensions[2].height = 18


def set_col_widths(ws, widths):
    for i, w in enumerate(widths, 1):
        ws.column_dimensions[get_column_letter(i)].width = w


# ═══════════════════════════════════════════════════════════════════════════════
# Sheet 0: Overview (TOC)
# ═══════════════════════════════════════════════════════════════════════════════
ws0 = wb.create_sheet("Overview")
ws0.sheet_view.showGridLines = False
ws0.column_dimensions["A"].width = 4
ws0.column_dimensions["B"].width = 28
ws0.column_dimensions["C"].width = 18
ws0.column_dimensions["D"].width = 16
ws0.column_dimensions["E"].width = 50

ws0.merge_cells("B1:E1")
ws0["B1"].value     = "Solar Monitoring System — Data Dictionary"
ws0["B1"].font      = Font(name="Calibri", bold=True, size=16, color="1F4E79")
ws0["B1"].alignment = LEFT

ws0.merge_cells("B2:E2")
ws0["B2"].value     = "สร้างสำหรับระบบเก็บข้อมูลโซลาร์ จาก Solarman API → MySQL 8.0"
ws0["B2"].font      = FONT_NOTE
ws0["B2"].alignment = LEFT
ws0.row_dimensions[1].height = 32
ws0.row_dimensions[2].height = 18

toc_headers = ["Sheet", "ชื่อ Table / View", "ประเภท", "จำนวน Columns", "คำอธิบาย"]
for i, h in enumerate(toc_headers, 2):
    c = ws0.cell(row=4, column=i)
    c.value = h
    c.font  = Font(name="Calibri", bold=True, size=10, color="1F4E79")
    c.fill  = fill(CLR_COL_HDR)
    c.alignment = CENTER
    c.border = BORDER_THIN

toc_data = [
    ("Overview",         "— (แผ่นนี้)",               "—",     "—",  "สารบัญและคำอธิบายระบบ"),
    ("solar_history",    "solar_history",              "TABLE", 28,   "เก็บข้อมูล power/energy ทุก 5 นาที จาก Solarman API"),
    ("api_request_log",  "api_request_log",            "TABLE", 6,    "บันทึก log การเรียก API ทุกครั้ง"),
    ("v_solar_daily",    "solar_daily",                "VIEW",  14,   "สรุปรายวัน: kWh, peak, self-consume%, savings"),
    ("v_solar_hourly",   "solar_hourly",               "VIEW",  9,    "Power curve รายชั่วโมง (เฉลี่ยทุกวัน)"),
    ("v_solar_monthly",  "solar_monthly",              "VIEW",  9,    "สรุปรายเดือน: kWh, savings, export income"),
    ("v_solar_latest",   "solar_latest",               "VIEW",  12,   "ข้อมูล real-time ล่าสุด 1 record"),
    ("v_day_compare",    "solar_day_compare",          "VIEW",  5,    "เปรียบวันนี้ vs เมื่อวาน vs เฉลี่ย 30 วัน"),
]

for r_idx, row in enumerate(toc_data):
    odd = r_idx % 2 == 0
    bg  = CLR_ROW_ODD if odd else CLR_ROW_EVEN
    for c_idx, val in enumerate(row, 2):
        c = ws0.cell(row=5 + r_idx, column=c_idx)
        c.value     = val
        c.font      = FONT_MONO if c_idx == 3 else FONT_NORMAL
        c.fill      = fill(bg)
        c.alignment = CENTER if c_idx in (4, 5) else LEFT
        c.border    = BORDER_THIN

# Notes section
notes_row = 15
ws0.merge_cells(f"B{notes_row}:E{notes_row}")
ws0[f"B{notes_row}"].value = "หมายเหตุระบบ"
ws0[f"B{notes_row}"].font  = Font(name="Calibri", bold=True, size=11, color="1F4E79")

notes = [
    ("Database",     "MySQL 8.0, charset=utf8mb4_unicode_ci, timezone=Asia/Bangkok"),
    ("API Source",   "Solarman Global API — https://globalapi.solarmanpv.com"),
    ("Station",      "KK-Home, Station ID: 60650830, Capacity: 5,750 W"),
    ("Poll Interval","ทุก POLL_MINUTES นาที (default: 10 นาที สำหรับ real-time)"),
    ("Backfill",     "ย้อนหลังตั้งแต่ 2023-05-23, interval 5 นาที/record"),
    ("Null Fields",  "API ส่ง null สำหรับ _value (kWh) fields → Go แปลงเป็น 0.0 อัตโนมัติ"),
    ("kWh Estimate", "Views ใช้ SUM(power_w × 5/60/1000) แทน เพราะ _value fields = 0 ทั้งหมด"),
    ("Rates",        "ค่าไฟ: 4.72 บาท/kWh (TOD self-use), 2.20 บาท/kWh (export FiT)"),
]
for i, (k, v) in enumerate(notes):
    r = notes_row + 1 + i
    ws0.cell(row=r, column=2).value = k
    ws0.cell(row=r, column=2).font  = Font(name="Calibri", bold=True, size=10)
    ws0.cell(row=r, column=2).alignment = LEFT
    ws0.merge_cells(f"C{r}:E{r}")
    ws0.cell(row=r, column=3).value = v
    ws0.cell(row=r, column=3).font  = FONT_NOTE
    ws0.cell(row=r, column=3).alignment = LEFT_NW
    ws0.row_dimensions[r].height = 16

ws0.row_dimensions[notes_row].height = 22

# ═══════════════════════════════════════════════════════════════════════════════
# Sheet 1: solar_history
# ═══════════════════════════════════════════════════════════════════════════════
TABLE_COLS = ["#", "Column Name", "Data Type", "Nullable", "Default", "Key", "Description (TH)", "Description (EN)", "Notes / ตัวอย่างค่า"]

solar_history_rows = [
    (1,  "id",                "BIGINT",       "NO",  "AUTO_INCREMENT", "PK",  "ลำดับ (auto)", "Row ID",                     "1, 2, 3 ..."),
    (2,  "station_id",        "BIGINT",       "NO",  "",               "IDX", "รหัสสถานี",    "Station ID from Solarman",   "60650830"),
    (3,  "station_name",      "VARCHAR(100)", "NO",  "",               "",    "ชื่อสถานี",    "Station name",               "'KK-Home'"),
    (4,  "recorded_at",       "DATETIME",     "NO",  "",               "UNI", "เวลาของ record (Asia/Bangkok)", "Data point timestamp", "2025-10-01 08:00:00"),
    (5,  "generation_power",  "DOUBLE",       "YES", "0",              "",    "กำลังผลิตโซลาร์ (W)", "Solar generation power (W)", "0 – 5,750"),
    (6,  "battery_power",     "DOUBLE",       "YES", "0",              "",    "กำลังแบตเตอรี่ (W) ลบ=ชาร์จ บวก=คาย", "Battery power (W), neg=charging, pos=discharging", "-3000 – 3000"),
    (7,  "battery_soc",       "DOUBLE",       "YES", "0",              "",    "ระดับแบตเตอรี่ (%)", "Battery State of Charge (%)", "⚠ API ส่ง null → เก็บ 0"),
    (8,  "charge_power",      "DOUBLE",       "YES", "0",              "",    "กำลังชาร์จแบต (W)", "Battery charge power (W)",   "0 – 3000"),
    (9,  "discharge_power",   "DOUBLE",       "YES", "0",              "",    "กำลังคายแบต (W)", "Battery discharge power (W)", "⚠ API ส่ง null → 0"),
    (10, "grid_power",        "DOUBLE",       "YES", "0",              "",    "กำลัง export ไปกริด (W)", "Export to grid (W)",      "0 – 6000"),
    (11, "wire_power",        "DOUBLE",       "YES", "0",              "",    "โหลดบ้านทั้งหมด (W)", "Total home load (W)",      "0 – 8000"),
    (12, "use_power",         "DOUBLE",       "YES", "0",              "",    "พลังงานที่บ้านใช้จากโซลาร์ (W)", "Power consumed from solar", "0 – 5750"),
    (13, "purchase_power",    "DOUBLE",       "YES", "0",              "",    "ซื้อจากการไฟฟ้า (W)", "Grid import power (W)",     "0 – 8000"),
    (14, "generation_value",  "DOUBLE",       "YES", "0",              "",    "ผลิตสะสมวันนั้น (kWh)", "Daily cumulative generation (kWh)", "⚠ API ส่ง null → 0 ทุก record"),
    (15, "buy_value",         "DOUBLE",       "YES", "0",              "",    "ซื้อกริดสะสม (kWh)", "Daily cumulative grid import (kWh)", "⚠ null → 0"),
    (16, "use_value",         "DOUBLE",       "YES", "0",              "",    "ใช้สะสม (kWh)", "Daily cumulative consumption (kWh)", "⚠ null → 0"),
    (17, "charge_value",      "DOUBLE",       "YES", "0",              "",    "ชาร์จแบตสะสม (kWh)", "Daily cumulative charge (kWh)", "⚠ null → 0"),
    (18, "discharge_value",   "DOUBLE",       "YES", "0",              "",    "คายแบตสะสม (kWh)", "Daily cumulative discharge (kWh)", "⚠ null → 0"),
    (19, "grid_value",        "DOUBLE",       "YES", "0",              "",    "export กริดสะสม (kWh)", "Daily cumulative export (kWh)", "⚠ null → 0"),
    (20, "generation_ratio",  "DOUBLE",       "YES", "0",              "",    "อัตราผลิตเทียบ capacity (%)", "Generation ratio vs capacity (%)", "0 – 100"),
    (21, "pr",                "DOUBLE",       "YES", "0",              "",    "Performance Ratio (%)", "Performance Ratio (%)", "⚠ null → 0"),
    (22, "cpr",               "DOUBLE",       "YES", "0",              "",    "Capacity Performance Ratio", "Capacity Performance Ratio", ""),
    (23, "full_power_hours",  "DOUBLE",       "YES", "0",              "",    "ชั่วโมงผลิตเต็มกำลัง (h)", "Full power equivalent hours (h)", "⚠ null → 0"),
    (24, "theoretical_gen",   "DOUBLE",       "YES", "0",              "",    "ผลิตตามทฤษฎี (kWh)", "Theoretical generation (kWh)", ""),
    (25, "irradiate",         "DOUBLE",       "YES", "0",              "",    "รังสีสะสม (kWh/m²)", "Cumulative irradiation (kWh/m²)", ""),
    (26, "irradiate_intensity","DOUBLE",      "YES", "0",              "",    "ความเข้มแสง (W/m²)", "Solar irradiance intensity (W/m²)", ""),
]

ws1 = wb.create_sheet("solar_history")
ws1.sheet_view.showGridLines = False
write_title_block(ws1, "TABLE: solar_history", "เก็บข้อมูล power/energy ทุก 5 นาที จาก Solarman API  |  UNIQUE KEY: (station_id, recorded_at)", CLR_TABLE_HDR)
write_header_row(ws1, 3, TABLE_COLS, is_view=False)
for i, row in enumerate(solar_history_rows):
    write_data_row(ws1, 4 + i, list(row), is_pk=(row[5] == "PK"), is_view=False, odd=(i % 2 == 0))

# Index info
idx_row = 4 + len(solar_history_rows) + 1
ws1.merge_cells(f"A{idx_row}:I{idx_row}")
ws1[f"A{idx_row}"].value     = "Indexes:  PK id (auto_increment)  |  IDX idx_station_time (station_id, recorded_at)  |  UNI uq_station_time (station_id, recorded_at)"
ws1[f"A{idx_row}"].font      = FONT_NOTE
ws1[f"A{idx_row}"].fill      = fill(CLR_SECTION)
ws1[f"A{idx_row}"].alignment = LEFT
ws1[f"A{idx_row}"].border    = BORDER_THIN

set_col_widths(ws1, [4, 22, 16, 10, 14, 6, 34, 34, 36])


# ═══════════════════════════════════════════════════════════════════════════════
# Sheet 2: api_request_log
# ═══════════════════════════════════════════════════════════════════════════════
api_log_rows = [
    (1, "no",           "BIGINT",       "NO",  "AUTO_INCREMENT", "PK", "ลำดับ (auto)", "Row ID",                      ""),
    (2, "requested_at", "DATETIME",     "NO",  "CURRENT_TIMESTAMP", "", "เวลาที่ส่ง request", "Request timestamp",    "2026-03-01 08:00:05"),
    (3, "endpoint",     "VARCHAR(100)", "NO",  "",               "",   "endpoint ที่เรียก", "API endpoint path",       "'/station/v1.0/history'"),
    (4, "station_id",   "BIGINT",       "YES", "NULL",           "",   "Station ID ถ้ามี", "Station ID if applicable", "60650830 / NULL"),
    (5, "date_param",   "DATE",         "YES", "NULL",           "",   "วันที่ที่ขอข้อมูล", "Date parameter for history requests", "2025-10-01"),
    (6, "status",       "VARCHAR(20)",  "YES", "'success'",      "",   "สถานะ", "Request status",                     "'success' / 'error'"),
    (7, "note",         "VARCHAR(255)", "YES", "NULL",           "",   "หมายเหตุ", "Additional info or error message",  "'records=274' / 'code 2101009'"),
]

ws2 = wb.create_sheet("api_request_log")
ws2.sheet_view.showGridLines = False
write_title_block(ws2, "TABLE: api_request_log", "บันทึก log การเรียก Solarman API ทุกครั้ง ทั้ง success และ error", CLR_TABLE_HDR)
write_header_row(ws2, 3, TABLE_COLS, is_view=False)
for i, row in enumerate(api_log_rows):
    write_data_row(ws2, 4 + i, list(row), is_pk=(row[5] == "PK"), is_view=False, odd=(i % 2 == 0))
set_col_widths(ws2, [4, 20, 16, 10, 20, 6, 30, 34, 36])


# ═══════════════════════════════════════════════════════════════════════════════
# Views — shared column template
# ═══════════════════════════════════════════════════════════════════════════════
VIEW_COLS = ["#", "Column Name", "Data Type", "Source", "Formula / Logic", "Description (TH)", "Description (EN)", "Unit", "Example"]

def make_view_sheet(name, title, subtitle, rows):
    ws = wb.create_sheet(name)
    ws.sheet_view.showGridLines = False
    write_title_block(ws, f"VIEW: {title}", subtitle, CLR_VIEW_HDR)
    write_header_row(ws, 3, VIEW_COLS, is_view=True)
    for i, row in enumerate(rows):
        for col_idx, val in enumerate(list(row), 1):
            c = ws.cell(row=4 + i, column=col_idx, value=val)
            bg = CLR_ROW_ODD_V if i % 2 == 0 else CLR_ROW_EVEN
            c.font      = FONT_MONO if col_idx in (2, 5) else FONT_NORMAL
            c.fill      = fill(bg)
            c.alignment = CENTER if col_idx in (3, 8) else LEFT_NW
            c.border    = BORDER_THIN
        ws.row_dimensions[4 + i].height = 30 if len(str(row[4])) > 40 else 18
    set_col_widths(ws, [4, 22, 14, 16, 44, 32, 32, 10, 24])
    return ws


# solar_daily
daily_rows = [
    (1,  "station_id",          "BIGINT",   "solar_history", "GROUP BY",                              "รหัสสถานี",              "Station ID",                     "",     "60650830"),
    (2,  "station_name",        "VARCHAR",  "solar_history", "GROUP BY",                              "ชื่อสถานี",              "Station name",                   "",     "KK-Home"),
    (3,  "date",                "DATE",     "solar_history", "DATE(recorded_at)",                     "วันที่",                 "Date",                           "",     "2025-10-01"),
    (4,  "est_solar_kwh",       "DECIMAL",  "solar_history", "SUM(generation_power)*5/60/1000",       "พลังงานโซลาร์ประมาณ",   "Estimated solar energy (kWh)",   "kWh",  "16.5"),
    (5,  "est_use_kwh",         "DECIMAL",  "solar_history", "SUM(use_power)*5/60/1000",              "พลังงานที่ใช้ประมาณ",   "Estimated consumption (kWh)",    "kWh",  "12.3"),
    (6,  "est_buy_kwh",         "DECIMAL",  "solar_history", "SUM(purchase_power)*5/60/1000",         "ซื้อจากกริดประมาณ",     "Estimated grid import (kWh)",    "kWh",  "3.2"),
    (7,  "est_export_kwh",      "DECIMAL",  "solar_history", "SUM(grid_power)*5/60/1000",             "ขายไปกริดประมาณ",       "Estimated grid export (kWh)",    "kWh",  "4.8"),
    (8,  "est_discharge_kwh",   "DECIMAL",  "solar_history", "SUM(GREATEST(battery_power,0))*5/60/1000","คายแบตประมาณ",       "Estimated battery discharge (kWh)","kWh","1.2"),
    (9,  "est_charge_kwh",      "DECIMAL",  "solar_history", "SUM(ABS(LEAST(battery_power,0)))*5/60/1000","ชาร์จแบตประมาณ",  "Estimated battery charge (kWh)", "kWh",  "2.1"),
    (10, "peak_solar_w",        "DOUBLE",   "solar_history", "MAX(generation_power)",                 "peak กำลังผลิตวันนั้น", "Daily peak solar power",         "W",    "4,592"),
    (11, "peak_load_w",         "DOUBLE",   "solar_history", "MAX(wire_power)",                       "peak โหลดบ้านวันนั้น",  "Daily peak home load",           "W",    "3,100"),
    (12, "peak_charge_w",       "DOUBLE",   "solar_history", "MAX(charge_power)",                     "peak การชาร์จแบต",      "Daily peak charge power",        "W",    "2,500"),
    (13, "solar_hours",         "DECIMAL",  "solar_history", "SUM(generation_power>10)*5/60",         "ชั่วโมงที่ผลิตได้",     "Solar production hours (>10W)",  "h",    "10.7"),
    (14, "self_consume_pct",    "DECIMAL",  "solar_history", "(SUM(gen)-SUM(grid))/SUM(gen)*100",     "% ใช้เองจากโซลาร์",    "Self-consumption ratio (%)",     "%",    "72.5"),
    (15, "capacity_factor_pct", "DECIMAL",  "solar_history", "AVG(generation_power)/5750*100",        "Capacity Factor",        "Capacity factor vs 5750W rated", "%",    "22.3"),
    (16, "est_saving_thb",      "DECIMAL",  "solar_history", "SUM(use_power)*5/60/1000*4.72",         "ค่าไฟประหยัด (บาท)",   "Estimated savings (THB @ 4.72)", "฿",   "75.61"),
    (17, "records",             "BIGINT",   "solar_history", "COUNT(*)",                              "จำนวน records วันนั้น", "Number of 5-min records",        "",     "288"),
]
make_view_sheet("v_solar_daily", "solar_daily", "สรุปรายวัน — kWh ประมาณจาก power × 5min | อัตราค่าไฟ 4.72 ฿/kWh | installed capacity 5,750 W", daily_rows)

# solar_hourly
hourly_rows = [
    (1, "station_id",    "BIGINT",  "solar_history", "GROUP BY",                   "รหัสสถานี",              "Station ID",               "",  "60650830"),
    (2, "hour",          "INT",     "solar_history", "HOUR(recorded_at)",           "ชั่วโมง (0-23)",         "Hour of day (0–23)",       "",  "12"),
    (3, "avg_solar_w",   "DECIMAL", "solar_history", "ROUND(AVG(generation_power))","กำลังผลิตเฉลี่ย",       "Avg solar power (W)",      "W", "3,202"),
    (4, "avg_use_w",     "DECIMAL", "solar_history", "ROUND(AVG(use_power))",       "กำลังใช้เฉลี่ย",        "Avg consumption (W)",      "W", "1,959"),
    (5, "avg_buy_w",     "DECIMAL", "solar_history", "ROUND(AVG(purchase_power))",  "ซื้อกริดเฉลี่ย",        "Avg grid import (W)",      "W", "0"),
    (6, "avg_export_w",  "DECIMAL", "solar_history", "ROUND(AVG(grid_power))",      "ขายกริดเฉลี่ย",         "Avg grid export (W)",      "W", "1,243"),
    (7, "avg_load_w",    "DECIMAL", "solar_history", "ROUND(AVG(wire_power))",      "โหลดบ้านเฉลี่ย",        "Avg home load (W)",        "W", "1,959"),
    (8, "max_solar_w",   "DECIMAL", "solar_history", "ROUND(MAX(generation_power))","peak กำลังผลิตชั่วโมงนั้น","Max solar (W) in this hour","W","5,275"),
    (9, "records",       "BIGINT",  "solar_history", "COUNT(*)",                    "จำนวน records",          "Record count",             "",  "1,240"),
]
make_view_sheet("v_solar_hourly", "solar_hourly", "Power Curve เฉลี่ยรายชั่วโมง (รวมทุกวัน) — ใช้ดูพฤติกรรมการผลิตโซลาร์ตามช่วงเวลา", hourly_rows)

# solar_monthly
monthly_rows = [
    (1, "station_id",            "BIGINT",  "solar_history", "GROUP BY",                            "รหัสสถานี",                "Station ID",                    "",  "60650830"),
    (2, "station_name",          "VARCHAR", "solar_history", "GROUP BY",                            "ชื่อสถานี",                "Station name",                  "",  "KK-Home"),
    (3, "month",                 "VARCHAR", "solar_history", "DATE_FORMAT(recorded_at,'%Y-%m')",    "เดือน (YYYY-MM)",          "Month (YYYY-MM format)",        "",  "2025-10"),
    (4, "days",                  "BIGINT",  "solar_history", "COUNT(DISTINCT DATE(recorded_at))",   "จำนวนวันที่มีข้อมูล",     "Days with data",                "",  "31"),
    (5, "est_solar_kwh",         "DECIMAL", "solar_history", "SUM(generation_power)*5/60/1000",     "พลังงานโซลาร์ประมาณ",     "Estimated solar energy (kWh)", "kWh","712.5"),
    (6, "est_use_kwh",           "DECIMAL", "solar_history", "SUM(use_power)*5/60/1000",            "พลังงานที่ใช้ประมาณ",     "Estimated consumption (kWh)",  "kWh","508.2"),
    (7, "est_buy_kwh",           "DECIMAL", "solar_history", "SUM(purchase_power)*5/60/1000",       "ซื้อกริดประมาณ",          "Estimated grid import (kWh)",  "kWh","86.3"),
    (8, "est_export_kwh",        "DECIMAL", "solar_history", "SUM(grid_power)*5/60/1000",           "ขายกริดประมาณ",           "Estimated grid export (kWh)",  "kWh","290.8"),
    (9, "peak_solar_w",          "DOUBLE",  "solar_history", "MAX(generation_power)",               "peak ทั้งเดือน",          "Monthly peak solar power",     "W", "5,208"),
    (10,"est_saving_thb",        "DECIMAL", "solar_history", "SUM(use_power)*5/60/1000*4.72",       "ค่าไฟประหยัด (บาท)",     "Estimated savings (THB)",      "฿", "1,925"),
    (11,"est_export_income_thb", "DECIMAL", "solar_history", "SUM(grid_power)*5/60/1000*2.20",      "รายได้ขายไฟ (บาท)",      "Estimated export income (THB)","฿", "754"),
]
make_view_sheet("v_solar_monthly", "solar_monthly", "สรุปรายเดือน — อัตราค่าไฟ self-use 4.72 ฿/kWh | FiT export rate 2.20 ฿/kWh", monthly_rows)

# solar_latest
latest_rows = [
    (1,  "station_id",         "BIGINT",  "solar_history", "MAX(recorded_at) subquery",             "รหัสสถานี",              "Station ID",                    "",  "60650830"),
    (2,  "station_name",       "VARCHAR", "solar_history", "",                                       "ชื่อสถานี",              "Station name",                  "",  "KK-Home"),
    (3,  "recorded_at",        "DATETIME","solar_history", "",                                       "เวลา record ล่าสุด",    "Latest record timestamp",       "",  "2026-03-01 09:30:00"),
    (4,  "solar_w",            "DOUBLE",  "solar_history", "generation_power",                       "กำลังผลิตโซลาร์",       "Current solar power (W)",       "W", "3,450"),
    (5,  "consumption_w",      "DOUBLE",  "solar_history", "use_power",                              "กำลังที่บ้านใช้",        "Current consumption (W)",       "W", "1,800"),
    (6,  "grid_import_w",      "DOUBLE",  "solar_history", "purchase_power",                         "ซื้อจากกริด",           "Current grid import (W)",       "W", "0"),
    (7,  "grid_export_w",      "DOUBLE",  "solar_history", "grid_power",                             "ขายไปกริด",             "Current grid export (W)",       "W", "1,200"),
    (8,  "home_load_w",        "DOUBLE",  "solar_history", "wire_power",                             "โหลดบ้านทั้งหมด",       "Total home load (W)",           "W", "1,800"),
    (9,  "battery_w",          "DOUBLE",  "solar_history", "battery_power",                          "กำลังแบต (ลบ=ชาร์จ)",  "Battery power (neg=charging)",  "W", "-500"),
    (10, "charge_w",           "DOUBLE",  "solar_history", "charge_power",                           "กำลังชาร์จแบต",         "Charge power (W)",              "W", "500"),
    (11, "capacity_pct",       "DOUBLE",  "solar_history", "generation_ratio",                       "อัตราผลิตเทียบ capacity","Generation ratio (%)",         "%", "60"),
    (12, "capacity_factor_pct","DECIMAL", "Calculated",    "generation_power/5750*100",              "Capacity factor",        "Capacity factor (%)",           "%", "59.1"),
    (13, "battery_status",     "VARCHAR", "Calculated",    "CASE WHEN battery_power<-10 THEN 'charging' WHEN >10 THEN 'discharging' ELSE 'idle'", "สถานะแบตเตอรี่", "Battery status", "", "'charging'"),
]
make_view_sheet("v_solar_latest", "solar_latest", "ข้อมูล Real-time ล่าสุด 1 record — ใช้สำหรับ Stat panels / gauge ใน dashboard", latest_rows)

# solar_day_compare
compare_rows = [
    (1, "today_kwh",       "DECIMAL", "solar_daily (sub)", "est_solar_kwh WHERE date=CURDATE()",               "พลังงานวันนี้ (kWh)",         "Today's estimated kWh",          "kWh", "14.2"),
    (2, "today_peak_w",    "DOUBLE",  "solar_daily (sub)", "peak_solar_w WHERE date=CURDATE()",                "peak วันนี้ (W)",             "Today's peak solar power (W)",   "W",   "4,200"),
    (3, "yesterday_kwh",   "DECIMAL", "solar_daily (sub)", "est_solar_kwh WHERE date=CURDATE()-1",             "พลังงานเมื่อวาน (kWh)",       "Yesterday's estimated kWh",      "kWh", "15.68"),
    (4, "yesterday_peak_w","DOUBLE",  "solar_daily (sub)", "peak_solar_w WHERE date=CURDATE()-1",              "peak เมื่อวาน (W)",           "Yesterday's peak solar power (W)","W",  "4,007"),
    (5, "avg30d_kwh",      "DECIMAL", "solar_daily (sub)", "AVG(est_solar_kwh) last 30 days excl today",       "เฉลี่ย 30 วัน (kWh)",        "30-day avg estimated kWh",       "kWh", "18.86"),
]
make_view_sheet("v_day_compare", "solar_day_compare", "เปรียบเทียบวันนี้ vs เมื่อวาน vs เฉลี่ย 30 วัน — ใช้ detect anomaly เบื้องต้น", compare_rows)


# ─── Save ────────────────────────────────────────────────────────────────────
out = "/app/output/solar_data_dictionary.xlsx"
wb.save(out)
print(f"Saved: {out}")
