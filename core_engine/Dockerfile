FROM golang:1.24-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /app

COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY core_engine ./core_engine

WORKDIR /app/core_engine
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/core-engine ./cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/core-engine .

EXPOSE 8005 50005 50010 50011 50007 50001
CMD ["./core-engine"]
