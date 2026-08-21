# RevenueIQ Development Kit (revenueiq_dev_kit)

This is the centralized development kit for Go services in the RevenueIQ ecosystem. It contains shared database models, protobuf definitions, generated gRPC clients/servers, and data-mapping helpers.

By consolidating API contracts and schemas here, we prevent duplication, enforce strict type safety, and simplify service-to-service communication.

---

## Directory Structure

```text
revenueiq_dev_kit/
├── bin/
│   └── protoc                  # Local protobuf compiler binary
├── include/
│   └── google/protobuf/        # Standard protobuf include files (any.proto, timestamp.proto, etc.)
├── pkg/
│   ├── models/                 # Database entity models (e.g., MongoDB BSON structs)
│   └── proto_helpers/          # Converter functions between models and proto messages
├── proto/
│   ├── common/                 # Common shared messages (Product, Session, etc.)
│   ├── mongo_service/          # MongoDB user service contracts
│   └── redis_service/          # Redis generic cache service contracts
├── compile_protos.sh           # Local script to compile all protobuf definitions
├── go.mod                      # Module definition
└── README.md                   # This instruction guide
```

---

## Workspace Setup & Development Flow

To ensure local changes to `revenueiq_dev_kit` are instantly available across all your services without needing to commit and push to GitHub, we use a **Go Workspace** and a local **replace directive** in `go.mod`.

### 1. Go Workspace (Local Development)
Ensure a `go.work` file exists at the root folder of your project (containing all modules):
```go
go 1.26.4

use (
	./revenueiq_dev_kit
	./mongodb_service
	./redis_service
)
```

### 2. Local Replace Directives in Services
Inside any service's `go.mod` (e.g., `redis_service/go.mod`), add a `replace` directive pointing to the local directory:
```go
replace github.com/RevenueIQ/revenueiq_dev_kit => ../revenueiq_dev_kit
```
This tells Go to resolve all imports starting with `github.com/RevenueIQ/revenueiq_dev_kit` directly from the local disk.

### 3. Production/CI Deployments
Before pushing to production or staging:
1. Commit and push your `revenueiq_dev_kit` changes to GitHub.
2. Create/push a Git tag (e.g., `git tag v1.0.0` and `git push origin v1.0.0`).
3. In your service's `go.mod`, remove the local `replace` directive and upgrade the version:
   `go get github.com/RevenueIQ/revenueiq_dev_kit@v1.0.0`

---

## How to Add or Modify Protobuf Contracts

If you need to add a new endpoint or define a new service:

### Step 1: Define the Contract
Create or modify a `.proto` file under `proto/{your_service}/`. Ensure you set the `go_package` option correctly. For example:
```protobuf
syntax = "proto3";

package catalog_service;

option go_package = "github.com/RevenueIQ/revenueiq_dev_kit/proto/catalog;catalog_proto";

message GetProductRequest {
  string id = 1;
}
...
```

### Step 2: Update the Compilation Script
Open [compile_protos.sh](compile_protos.sh) and append your new compilation command. For example:
```bash
echo "Compiling Catalog Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/catalog/catalog.proto
```

### Step 3: Run the Compilation Script
Run the script from the root of `revenueiq_dev_kit` to generate the `.pb.go` and `_grpc.pb.go` stubs:
```bash
./compile_protos.sh
```

### Step 4: Add Helpers (Optional)
If your service maps internal structs (e.g., database BSON models) to these new proto messages, define them under `pkg/models/` and write converters under `pkg/proto_helpers/`.

---

## Cache Namespaces Guidelines

All Redis cache operations must specify a valid namespace. To prevent services from creating random, ad-hoc namespaces that clutter Redis:
1. Open [pkg/namespaces/namespaces.go](pkg/namespaces/namespaces.go).
2. Register your namespace as a constant under the `CacheNamespace` type.
3. Update the `IsValid()` switch statement to include your new namespace.

Services must import `github.com/RevenueIQ/revenueiq_dev_kit/pkg/namespaces` and use these typed constants for cache operations. The Redis Service validates all requests at the gRPC boundary and rejects unregistered namespaces.

---

## Coding Standards & Architectural Patterns

To maintain a clean and scalable codebase across all repos and services, follow these design principles:

### 1. Database Separation of Concerns
Never place all database queries into a single file. Under `internal/db/`, maintain separate files for each domain/model:
- `connection.go`: Handles DB connection pooling and initialization logic.
- `{domain_name}.go` (e.g. `user.go`, `product.go`): Contains only DB CRUD operations for that specific model/collection.

### 2. Clean Bootstrapping (`main.go`)
Keep `cmd/server/main.go` thin. Do not inline gRPC server setups or client connections:
- Place gRPC server builder/registry functions under `internal/grpc_servers/`.
- Place gRPC client connections under `internal/grpc_clients/`.
- `main.go` should only retrieve environment variables, initialize connections, and trigger the builders.
