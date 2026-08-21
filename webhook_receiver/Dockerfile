# =========================
# BUILD STAGE
# =========================
FROM golang:alpine AS builder

ENV GOWORK=off
WORKDIR /app

COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY webhook_receiver ./webhook_receiver

WORKDIR /app/webhook_receiver
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /app/webhook-receiver ./cmd/server/main.go

# =========================
# MINIMAL PROD RUNTIME
# =========================
FROM alpine:3.20 AS prod

RUN apk add --no-cache tzdata ca-certificates netcat-openbsd && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder --chown=appuser:appgroup /app/webhook-receiver /usr/local/bin/webhook-receiver

USER appuser
EXPOSE 50001 8001
ENTRYPOINT ["webhook-receiver"]
