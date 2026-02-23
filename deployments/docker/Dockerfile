# syntax=docker/dockerfile:1

# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Cache dependencies separately from application code.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags="-s -w -X main.Version=$(git describe --tags --always 2>/dev/null || echo dev)" \
      -o /out/riskengine \
      ./cmd/server

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM scratch

# Import CA certs and timezone data from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /out/riskengine /riskengine

# Default config directory (mount a volume or ConfigMap in production).
COPY configs/config.example.yaml /configs/config.yaml

EXPOSE 8080 9090

ENTRYPOINT ["/riskengine", "-config", "/configs/config.yaml"]
