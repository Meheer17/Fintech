#!/bin/bash
set -e

# Change directory to the script's directory (revenueiq_dev_kit root)
cd "$(dirname "$0")"

# Ensure local go bin path is included for protoc-gen-go and protoc-gen-go-grpc
export PATH="$PATH:$HOME/go/bin"

echo "Compiling MongoDB Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/mongo_service/mongodb_service.proto

echo "Compiling Redis Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/redis_service/redis_service.proto

echo "Compiling Auth Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/auth_service/auth_service.proto

echo "Compiling Queue Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/queue_service/queue_service.proto

echo "Compiling Notification Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/notification_service/notification_service.proto

echo "Compiling S3 Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/s3_service/s3_service.proto

echo "Compiling Monitoring Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/monitoring_service/monitoring_service.proto

echo "Compiling Order Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/order_service/order_service.proto

echo "Compiling Fleet Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/fleet_service/fleet_service.proto

echo "Compiling Flight Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/flight_service/flight_service.proto

echo "Compiling Telemetry Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/telemetry_service/telemetry_service.proto

echo "Compiling Operations Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/operations_service/operations_service.proto

echo "Compiling Delivery Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/delivery_service/delivery_service.proto

echo "Compiling Payment Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/payment_service/payment_service.proto

echo "Compiling Warehouse Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/warehouse_service/warehouse_service.proto

echo "Compiling Analytics Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/analytics_service/analytics_service.proto

echo "Compiling Common protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             proto/common/common.proto

echo "Compiling Webhook Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/webhook/webhook.proto

echo "Compiling Failure Detector Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/failure/failure.proto

echo "Compiling Recovery Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/recovery/recovery.proto

echo "Compiling Reconciliation Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/reconciliation/reconciliation.proto

echo "Compiling Audit Service protos..."
./bin/protoc -I. -Iinclude --go_out=. --go_opt=paths=source_relative \
             --go-grpc_out=. --go-grpc_opt=paths=source_relative \
             proto/audit/audit.proto

echo "All protobuf contracts compiled successfully!"





