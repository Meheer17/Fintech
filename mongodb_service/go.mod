module github.com/RevenueIQ/mongo_service

go 1.26.4

require (
	github.com/RevenueIQ/revenueiq_dev_kit v0.0.0-20260711150257-dd13e3c65b41
	go.mongodb.org/mongo-driver v1.17.9
	google.golang.org/grpc v1.82.1
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/golang/snappy v1.0.0 // indirect
	github.com/klauspost/compress v1.19.0 // indirect
	github.com/montanaflynn/stats v0.9.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260706201446-f0a921348800 // indirect
)

replace github.com/RevenueIQ/revenueiq_dev_kit => ../revenueiq_dev_kit
