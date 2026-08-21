# =========================
# BUILD STAGE
# =========================
FROM golang:alpine AS builder

ENV GOWORK=off
WORKDIR /app

COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY dashboard_api ./dashboard_api

WORKDIR /app/dashboard_api
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /app/dashboard-api ./cmd/server/main.go

# =========================
# MINIMAL PROD RUNTIME
# =========================
FROM alpine:3.20 AS prod

RUN apk add --no-cache tzdata ca-certificates netcat-openbsd && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder --chown=appuser:appgroup /app/dashboard-api /usr/local/bin/dashboard-api

USER appuser
EXPOSE 50005 8005
ENTRYPOINT ["dashboard-api"]
