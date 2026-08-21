# =========================
# BUILD STAGE
# =========================
FROM golang:alpine AS deps

RUN apk add --no-cache git
ENV GOWORK=off

WORKDIR /app

COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY auth_service ./auth_service

WORKDIR /app/auth_service
RUN go mod tidy

# =========================
# PROD BUILD
# =========================
FROM deps AS prod-builder

WORKDIR /app/auth_service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /app/auth-service ./cmd/server/main.go

# =========================
# DEV MODE (Hot Reload)
# =========================
FROM deps AS dev

WORKDIR /app
RUN GOFLAGS="" go install github.com/cespare/reflex@latest

WORKDIR /app/auth_service
CMD ["reflex", "-r", "\\.go$", "-s", "--", "go", "run", "./cmd/server/main.go"]

# =========================
# PROD RUNTIME
# =========================
FROM alpine:3.20 AS prod

RUN apk add --no-cache tzdata ca-certificates netcat-openbsd && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=prod-builder --chown=appuser:appgroup /app/auth-service /usr/local/bin/auth-service

USER appuser
EXPOSE 50009 8081
ENTRYPOINT ["auth-service"]
