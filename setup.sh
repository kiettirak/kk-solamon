#!/bin/bash
# ============================================================
# setup.sh — ติดตั้งระบบ Solar Monitor บน Ubuntu VM (Proxmox)
# วิธีใช้: ssh เข้า VM แล้วรัน
#   chmod +x setup.sh && sudo ./setup.sh
# ============================================================

set -e  # หยุดทำงานทันทีถ้ามี command ใดๆ error

echo "🔧 [1/6] อัพเดทระบบ..."
# อัพเดท package list + ติดตั้ง tools พื้นฐาน (git, curl, etc.)
apt-get update -y
apt-get install -y git curl ca-certificates gnupg

echo "🐳 [2/6] ติดตั้ง Docker Engine..."
# ใช้ Docker official script — ติดตั้ง Docker + Docker Compose plugin อัตโนมัติ
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com | sh
    # เพิ่ม user ปัจจุบันเข้ากลุ่ม docker (ไม่ต้อง sudo ทุกครั้ง)
    usermod -aG docker "$SUDO_USER"
    echo "   Docker ติดตั้งเรียบร้อย"
else
    echo "   Docker มีอยู่แล้ว: $(docker --version)"
fi

# ตั้ง timezone เป็น Bangkok (สำคัญ — ข้อมูลทุกอย่างใช้เวลาไทย)
echo "🕐 [3/6] ตั้ง timezone เป็น Asia/Bangkok..."
timedatectl set-timezone Asia/Bangkok

echo "📂 [4/6] Clone project..."
PROJECT_DIR="/opt/solar-monitor"

# ถ้ามีอยู่แล้วให้ pull ใหม่ ไม่ต้อง clone ซ้ำ
if [ -d "$PROJECT_DIR" ]; then
    echo "   Project มีอยู่แล้ว — ทำ git pull..."
    cd "$PROJECT_DIR" && git pull
else
    # ← เปลี่ยน URL เป็น repo จริงของคุณ
    # git clone https://github.com/YOUR_USER/solarman-go-client.git "$PROJECT_DIR"
    # ถ้ายังไม่ push ขึ้น git ให้ copy ด้วย scp แทน:
    echo "   ⚠ กรุณา copy project มาที่ $PROJECT_DIR ด้วย scp หรือ git clone"
    echo "   ตัวอย่าง: scp -r solarman-go-client/ user@<VM_IP>:$PROJECT_DIR"
    mkdir -p "$PROJECT_DIR"
fi

cd "$PROJECT_DIR"

echo "📝 [5/6] ตั้งค่า .env.prod..."
# ถ้ายังไม่มี .env.prod ให้ copy จาก template แล้วเตือนให้แก้
if [ ! -f .env.prod ]; then
    if [ -f .env.prod.example ]; then
        cp .env.prod.example .env.prod
    fi
    echo "   ⚠ กรุณาแก้ไข .env.prod ก่อนรัน:"
    echo "   nano $PROJECT_DIR/.env.prod"
    echo ""
    echo "   ค่าที่ต้องตั้ง:"
    echo "   - API_ID, API_SECRET  (จาก Solarman)"
    echo "   - EMAIL, PASSWORD     (account Solarman)"
    echo "   - INFLUX_TOKEN        (เปลี่ยนเป็น token ที่ปลอดภัย)"
    echo "   - MYSQL_PASSWORD      (เปลี่ยนรหัสผ่าน)"
    echo "   - GRAFANA_PASSWORD    (เปลี่ยนรหัสผ่าน)"
    echo "   - POLL_MINUTES        (ดึงข้อมูลทุกกี่นาที เช่น 5)"
fi

echo "🚀 [6/6] เริ่มทำงาน Docker Compose..."
# สร้าง Go binary + รัน containers ทั้งหมด (InfluxDB, MySQL, Grafana, Go app)
# --build = build Dockerfile ใหม่ทุกครั้ง (ข้ามถ้า cache เดิมยังใช้ได้)
# -d = detach mode (รัน background)
docker compose -f docker-compose.prod.yml --env-file .env.prod up --build -d

echo ""
echo "============================================"
echo "✅ ติดตั้งเรียบร้อย!"
echo ""
echo "📊 Grafana:    http://$(hostname -I | awk '{print $1}'):3000"
echo "📈 InfluxDB:   http://$(hostname -I | awk '{print $1}'):8086"
echo "🐬 MySQL:      internal only (port 3306 ไม่ expose)"
echo "⚡ Solar App:  ดึงข้อมูลอัตโนมัติทุก ${POLL_MINUTES:-5} นาที"
echo ""
echo "📋 ดู logs:     docker compose -f docker-compose.prod.yml logs -f solar_app"
echo "🔄 restart:    docker compose -f docker-compose.prod.yml restart solar_app"
echo "🛑 หยุด:       docker compose -f docker-compose.prod.yml down"
echo "============================================"
