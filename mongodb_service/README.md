# MongoDB Service (Golang gRPC + MongoDB + Redis Cache)

This service provides a gRPC interface to MongoDB and utilizes Redis Cache (via `redis_service`) for cache-aside queries and cache invalidations.

It is structured strictly to separate database and server operations, keeping bootstrap scripts minimal and domain modules decoupled.

---

## Project Structure

```text
mongodb_service/
├── cmd/
│   ├── client/
│   │   └── main.go             # gRPC integration test client (runs User and Cached Product flows)
│   └── server/
│       └── main.go             # Server entry point (handles bootstrap and configurations only)
├── internal/
│   ├── db/                     # MongoDB Database Layer (strictly decoupled by domain)
│   │   ├── connection.go       # MongoDB connection pool setup
│   │   ├── user.go             # BSON queries/CRUD for Users
│   │   └── product.go          # BSON queries/CRUD for Products
│   ├── grpc_clients/           # Outgoing gRPC Clients
│   │   └── redis_service.go    # Client connector for the Redis Cache Service
│   ├── grpc_servers/           # Incoming gRPC Server Configurations
│   │   └── mongo_service.go    # gRPC server builder and handler registration
│   └── server/
│       └── server.go           # Service API handlers (implements UserService & ProductService)
├── go.mod                      # Go module definition
└── README.md                   # This instruction guide
```

---

## Coding Standards & Developer Guidelines

When extending this service, adding new endpoints, or developing new microservices, you **MUST** follow these structural design patterns:

### 1. DECUPLED DATABASE LAYER
Do not mix database connections and operations, and do not place different entity operations in the same file.
- **Connection pool**: Keep in `internal/db/connection.go`.
- **Entities**: Keep each entity's operations in its own file (`internal/db/{entity}.go`). Files must only contain CRUD and queries relating to that entity.

### 2. CLEAN SERVER BOOTSTRAP (`main.go`)
The `cmd/server/main.go` file must remain thin (bootstrap only). 
- Do not initialize TCP listeners, start gRPC servers, or dial outgoing clients in `main.go`.
- **Outgoing connections**: Build them under `internal/grpc_clients/`.
- **Server initialization**: Set them up under `internal/grpc_servers/` and return the `*grpc.Server` and `net.Listener` to `main.go` to serve.

### 3. CACHE NAMESPACE RESTRICTION
Never use raw string literals for cache namespaces (e.g. do not use `"my_custom_ns"`). 
- Always import `github.com/RevenueIQ/revenueiq_dev_kit/pkg/namespaces`.
- Only use the strongly-typed constants defined in the dev kit (e.g. `namespaces.Products`). 
- If you need a new namespace, register it in the `revenueiq_dev_kit` first.

---

## How to Run the Service

### 1. Spin up Databases (MongoDB & Redis)
```bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
docker run -d -p 6379:6379 --name redis redis:alpine
```

### 2. Start the Services
Start the Redis cache service first:
```bash
cd redis_service
go run cmd/server/main.go
```

Start this service:
```bash
cd mongodb_service
go run cmd/server/main.go
```

### 3. Run the Test Client
To run the automated User CRUD and Product Cache-aside integration tests:
```bash
cd mongodb_service
go run cmd/client/main.go
```
