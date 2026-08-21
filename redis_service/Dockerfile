# =========================
# BUILD STAGE
# =========================
FROM golang:alpine AS deps

RUN apk add --no-cache git
ENV GOWORK=off

WORKDIR /app

COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY redis_service ./redis_service

WORKDIR /app/redis_service
RUN go mod tidy

# =========================
# PROD BUILD
# =========================
FROM deps AS prod-builder

WORKDIR /app/redis_service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /app/redis-service ./cmd/server/main.go

# =========================
# DEV MODE (Hot Reload)
# =========================
FROM deps AS dev

WORKDIR /app
RUN GOFLAGS="" go install github.com/cespare/reflex@latest

WORKDIR /app/redis_service
CMD ["reflex", "-r", "\\.go$", "-s", "--", "go", "run", "./cmd/server/main.go"]

# =========================
# PROD RUNTIME
# =========================
FROM alpine:3.20 AS prod

RUN apk add --no-cache tzdata ca-certificates netcat-openbsd && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=prod-builder --chown=appuser:appgroup /app/redis-service /usr/local/bin/redis-service

USER appuser
EXPOSE 50011
ENTRYPOINT ["redis-service"]
