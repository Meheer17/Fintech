# Auth Service (Golang gRPC + User check + OTP)

This is a gRPC microservice that handles authentication flows (Login, SendOTP, and VerifyOTP) within the RevenueIQ ecosystem.

It communicates with:
- **`mongodb_service`** (via gRPC port `50051`) to look up and query users by email.
- **`redis_service`** (via gRPC port `50052`) to cache OTP codes securely under the `sessions_ns` namespace.

---

## Project Structure

```text
auth_service/
├── cmd/
│   └── server/
│       └── main.go             # Server entry point (binds ports, dials mongodb & redis)
├── internal/
│   └── server/
│       └── server.go           # gRPC handler logic (Login, SendOTP, VerifyOTP)
├── go.mod                      # Go module dependencies
└── README.md                   # This instruction guide
```

---

## Configuration

The service is configured using the following environment variables:
- `PORT`: gRPC listening port (default: `50053`).
- `MONGO_SERVICE_ADDR`: Target gRPC address for the User/MongoDB service (default: `localhost:50051`).
- `REDIS_SERVICE_ADDR`: Target gRPC address for the Redis cache service (default: `localhost:50052`).

---

## How to Run

1. Make sure `mongodb_service` and `redis_service` are running.
2. Run the Auth Service server:
   ```bash
   go run cmd/server/main.go
   ```
