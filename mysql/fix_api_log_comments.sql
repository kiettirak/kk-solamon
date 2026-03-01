SET NAMES utf8mb4;

ALTER TABLE api_request_log
  COMMENT = 'บันทึกจำนวนและเวลาที่ request Solarman API',
  MODIFY COLUMN no           BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'ลำดับ (auto increment)',
  MODIFY COLUMN requested_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'เวลาที่ส่ง request',
  MODIFY COLUMN endpoint     VARCHAR(100) NOT NULL COMMENT 'endpoint ที่เรียก เช่น /station/v1.0/history',
  MODIFY COLUMN station_id   BIGINT NULL           COMMENT 'station ID ถ้ามี',
  MODIFY COLUMN date_param   DATE NULL             COMMENT 'วันที่ที่ขอข้อมูล (สำหรับ history endpoint)',
  MODIFY COLUMN status       VARCHAR(20) DEFAULT 'success' COMMENT 'ผลลัพธ์: success หรือ error',
  MODIFY COLUMN note         VARCHAR(255) NULL     COMMENT 'หมายเหตุ เช่น records=274 หรือ error message';
