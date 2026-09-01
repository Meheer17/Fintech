FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY core_engine ./core_engine

WORKDIR /app/core_engine
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/core-engine ./cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache tzdata ca-certificates netcat-openbsd
WORKDIR /app
COPY --from=builder /app/core-engine .

EXPOSE 8005 50005 50010 50011 50007 50001
CMD ["./core-engine"]
