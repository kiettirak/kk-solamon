# ---- Stage 1: Build ----
# ใช้ Go image เพื่อ compile — ไม่ติดตาม runtime
FROM golang:1.23-alpine AS builder

WORKDIR /app

# copy dependency files ก่อน เพื่อ cache layer
COPY go.mod go.sum ./
RUN go mod download

# copy source code
COPY . .

# build ทุก binary ในครั้งเดียว
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/solarman-client ./cmd/solarman-client
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/weather-collector ./cmd/weather-collector
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/backfill ./cmd/backfill
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/pm-backfill ./cmd/pm-backfill

# ---- Stage 2: Runtime ----
# ใช้ alpine เล็กๆ — ไม่มี Go toolchain อีกต่อไป
FROM alpine:3.20

# timezone data สำหรับ Asia/Bangkok
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app

# copy ทุก binary จาก stage 1
COPY --from=builder /bin/solarman-client .
COPY --from=builder /bin/weather-collector .
COPY --from=builder /bin/backfill .
COPY --from=builder /bin/pm-backfill .

# output directory สำหรับ JSON files
RUN mkdir -p ./output

# default CMD คือ solar polling client
# เปลี่ยนได้ใน docker-compose ด้วย command: ["./weather-collector"]
CMD ["./solarman-client"]
