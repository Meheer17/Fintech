# =========================
# BUILD STAGE
# =========================
FROM golang:1.26-alpine AS builder

ENV GOWORK=off
WORKDIR /app

COPY revenueiq_dev_kit/go.mod revenueiq_dev_kit/go.sum* ./revenueiq_dev_kit/
COPY voice_recovery_worker/go.mod voice_recovery_worker/go.sum* ./voice_recovery_worker/

WORKDIR /app/voice_recovery_worker
RUN go mod download || true

WORKDIR /app
COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY voice_recovery_worker ./voice_recovery_worker

WORKDIR /app/voice_recovery_worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /app/voice-recovery-worker ./cmd/server/main.go

# =========================
# MINIMAL PROD RUNTIME
# =========================
FROM alpine:3.20 AS prod

RUN apk add --no-cache tzdata ca-certificates netcat-openbsd && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder --chown=appuser:appgroup /app/voice-recovery-worker /usr/local/bin/voice-recovery-worker

USER appuser
EXPOSE 8009
ENTRYPOINT ["voice-recovery-worker"]
