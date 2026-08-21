# Redis Service (Golang gRPC + Redis Cache)

This is a Golang-based service that acts as a gRPC interface to Redis cache. It is designed to support generic key-value storage of any compiled protobuf type from the `revenueiq_dev_kit` by leveraging `google.protobuf.Any`.

## Features
- **Generic Key-Value Caching**: Store, retrieve, and delete arbitrary data.
- **Dynamic Type Serialization**: Uses `google.protobuf.Any` to package and serialize any protobuf message dynamically, preserving type information across microservice boundaries.
- **Shared Contracts**: Uses central contracts and structures imported directly from `revenueiq_dev_kit`.

## Project Structure

```text
redis_service/
├── cmd/
│   ├── client/
│   │   └── main.go             # gRPC test client (demonstrates dynamic Set/Get/Delete)
│   └── server/
│       └── main.go             # gRPC server entry point
├── internal/
│   ├── db/
│   │   └── redis.go            # Redis client connection and database operations
│   └── server/
│       └── server.go           # gRPC Service server implementation
├── go.mod                      # Go module definition
└── README.md                   # This instruction guide
```

---

## How to Run the Service

### 1. Start Redis
You must have a Redis server running. If you have Docker, you can run:
```bash
docker run -d -p 6379:6379 --name redis redis:alpine
```

### 2. Configure and Run the Server
The server is configured via environment variables. You can run the server directly:
```bash
# Optional configuration variables:
# export PORT="50052"
# export REDIS_ADDR="localhost:6379"
# export REDIS_PASSWORD=""
# export REDIS_DB="0"

go run cmd/server/main.go
```

### 3. Run the Test Client
To test that gRPC calls are successfully storing, retrieving, and dynamically decoding any dev kit data type, run:
```bash
go run cmd/client/main.go
```
This test client:
1. Connects to the Redis cache gRPC service.
2. instantiates `User`, `Product`, and `Session` types from the `revenueiq_dev_kit`.
3. Wraps them using `anypb.New`.
4. Caches them via `Set` in Redis.
5. Retrieves them using `Get`.
6. Inspects the `google.protobuf.Any` type URL and unmarshals back to the correct underlying Go struct dynamically.
