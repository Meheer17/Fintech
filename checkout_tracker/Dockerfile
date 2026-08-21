# =========================
# BUILD STAGE
# =========================
FROM golang:1.26-alpine AS builder

ENV GOWORK=off
WORKDIR /app

COPY revenueiq_dev_kit/go.mod revenueiq_dev_kit/go.sum* ./revenueiq_dev_kit/
COPY checkout_tracker/go.mod checkout_tracker/go.sum* ./checkout_tracker/

WORKDIR /app/checkout_tracker
RUN go mod download || true

WORKDIR /app
COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY checkout_tracker ./checkout_tracker

WORKDIR /app/checkout_tracker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /app/checkout-tracker ./cmd/server/main.go

# =========================
# MINIMAL PROD RUNTIME
# =========================
FROM alpine:3.20 AS prod

RUN apk add --no-cache tzdata ca-certificates netcat-openbsd && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder --chown=appuser:appgroup /app/checkout-tracker /usr/local/bin/checkout-tracker

USER appuser
EXPOSE 50008 8008
ENTRYPOINT ["checkout-tracker"]
