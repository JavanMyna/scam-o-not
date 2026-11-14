# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /scam-checker-bot ./main.go

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /scam-checker-bot /scam-checker-bot

# Expose webhook port (ignored in polling mode)
EXPOSE 8080

ENTRYPOINT ["/scam-checker-bot"]
