# =========================
# DEPENDENCY LAYER (cached)
# =========================
FROM golang:1.26.4-alpine AS deps

RUN apk add --no-cache git

# Disable Go workspace mode so this Dockerfile runs independently of go.work
ENV GOWORK=off
ENV GOPROXY=https://goproxy.io,https://proxy.golang.org,direct

WORKDIR /app

# Copy ONLY the required module manifests for compiling notification_service
COPY revenueiq_dev_kit/go.mod revenueiq_dev_kit/go.sum* ./revenueiq_dev_kit/
COPY notification_service/go.mod notification_service/go.sum* ./notification_service/

# Download dependencies for the notification_service module
WORKDIR /app/notification_service
RUN go mod download

# =========================
# PROD BUILD
# =========================
FROM deps AS prod-builder

WORKDIR /app
# Copy source code for target service and its local dependency
COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY notification_service ./notification_service

# Download full transitive source dependencies and build static minimal binary
WORKDIR /app/notification_service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build     -ldflags="-s -w"     -trimpath     -o /app/notification-service ./cmd/server/main.go

# =========================
# DEV MODE (Hot Reload)
# =========================
FROM deps AS dev

WORKDIR /app

# Install reflex for hot reload
RUN GOFLAGS="" go install github.com/cespare/reflex@latest

COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY notification_service ./notification_service

WORKDIR /app/notification_service
CMD ["reflex", "-r", "\.go$", "-s", "--", "go", "run", "./cmd/server/main.go"]

# =========================
# PROD RUNTIME
# =========================
FROM alpine:3.20 AS prod

# Install security certificates & timezone data, create non-root user
RUN apk add --no-cache tzdata ca-certificates &&     addgroup -S appgroup &&     adduser -S appuser -G appgroup

WORKDIR /app

# Copy compiled binary from builder with non-root ownership
COPY --from=prod-builder --chown=appuser:appgroup /app/notification-service /usr/local/bin/notification-service

# Run as non-root user for security
USER appuser

ENTRYPOINT ["notification-service"]
