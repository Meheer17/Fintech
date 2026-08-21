# =========================
# BUILD STAGE
# =========================
FROM golang:alpine AS deps

RUN apk add --no-cache git
ENV GOWORK=off

WORKDIR /app

COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY mongodb_service ./mongodb_service

WORKDIR /app/mongodb_service
RUN go mod tidy

# =========================
# PROD BUILD
# =========================
FROM deps AS prod-builder

WORKDIR /app/mongodb_service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /app/mongodb-service ./cmd/server/main.go

# =========================
# DEV MODE (Hot Reload)
# =========================
FROM deps AS dev

WORKDIR /app
RUN GOFLAGS="" go install github.com/cespare/reflex@latest

WORKDIR /app/mongodb_service
CMD ["reflex", "-r", "\\.go$", "-s", "--", "go", "run", "./cmd/server/main.go"]

# =========================
# PROD RUNTIME
# =========================
FROM alpine:3.20 AS prod

RUN apk add --no-cache tzdata ca-certificates netcat-openbsd && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=prod-builder --chown=appuser:appgroup /app/mongodb-service /usr/local/bin/mongodb-service

USER appuser
EXPOSE 50010
ENTRYPOINT ["mongodb-service"]
