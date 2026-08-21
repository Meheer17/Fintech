# =========================
# BUILD STAGE
# =========================
FROM golang:1.26-alpine AS builder

ENV GOWORK=off
WORKDIR /app

COPY revenueiq_dev_kit/go.mod revenueiq_dev_kit/go.sum* ./revenueiq_dev_kit/
COPY audit_service/go.mod audit_service/go.sum* ./audit_service/

WORKDIR /app/audit_service
RUN go mod download || true

WORKDIR /app
COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY audit_service ./audit_service

WORKDIR /app/audit_service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /app/audit-service ./cmd/server/main.go

# =========================
# MINIMAL PROD RUNTIME
# =========================
FROM alpine:3.20 AS prod

RUN apk add --no-cache tzdata ca-certificates netcat-openbsd && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder --chown=appuser:appgroup /app/audit-service /usr/local/bin/audit-service

USER appuser
EXPOSE 50007
ENTRYPOINT ["audit-service"]
