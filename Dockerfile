# ==========================================
# Stage 1: Build Binary Golang
# ==========================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Salin modul dan unduh dependensi
COPY go.mod ./
RUN go mod download

# Salin seluruh source code
COPY . .

# Compile binary static (CGO_ENABLED=0 agar portabel dan ringan)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/starship-dataplane ./cmd/server

# ==========================================
# Stage 2: Minimal Runtime Image (Ringan & Aman)
# ==========================================
FROM alpine:3.20

# Buat user non-root demi keamanan server
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Ambil binary hasil compile dari stage 1
COPY --from=builder /app/starship-dataplane .

# Ganti hak milik ke user non-root
USER appuser

# Port yang dibuka di dalam container
EXPOSE 8080

# Jalankan server
CMD ["./starship-dataplane"]
