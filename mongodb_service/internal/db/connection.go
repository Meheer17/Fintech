package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDatabase wraps the MongoDB client, database, and collections.
type MongoDatabase struct {
	Client      *mongo.Client
	Db          *mongo.Database
	Col         *mongo.Collection // Users collection
	OrdersCol   *mongo.Collection // Orders collection
	DronesCol   *mongo.Collection // Drones collection
	NoFlyCol    *mongo.Collection // No-Fly Zones collection
	StationsCol *mongo.Collection // Charging Stations collection
	MaintCol    *mongo.Collection // Maintenance logs collection
	DeliveryCol *mongo.Collection // Delivery logs collection
	TxCol       *mongo.Collection // Transaction ledger collection
	WalletCol             *mongo.Collection // User credit wallets collection
	BusinessesCol         *mongo.Collection // Businesses collection
	WarehousesCol         *mongo.Collection // Warehouses collection
	InventoryCol          *mongo.Collection // Inventory SKUs collection
	AnalyticsSnapshotsCol *mongo.Collection // Analytics snapshots collection
	FailureEventsCol      *mongo.Collection // Failure events collection
	WorkflowsCol          *mongo.Collection // Workflows collection
}

// ConnectMongo initializes a connection to MongoDB.
func ConnectMongo(ctx context.Context, uri, dbName, colName string) (*MongoDatabase, error) {
	// Set client options
	clientOpts := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	// Ping the primary deployment to verify connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = client.Ping(pingCtx, nil)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	col := db.Collection(colName)
	ordersCol := db.Collection("orders")
	dronesCol := db.Collection("drones")
	noFlyCol := db.Collection("nofly_zones")
	stationsCol := db.Collection("charging_stations")
	maintCol := db.Collection("maintenance_records")
	deliveryCol := db.Collection("delivery_records")
	txCol := db.Collection("transactions")
	walletCol := db.Collection("user_wallets")
	businessesCol := db.Collection("businesses")
	warehousesCol := db.Collection("warehouses")
	inventoryCol := db.Collection("inventory_skus")
	analyticsSnapshotsCol := db.Collection("analytics_snapshots")
	failureEventsCol := db.Collection("failure_events")
	workflowsCol := db.Collection("workflows")

	return &MongoDatabase{
		Client:                client,
		Db:                    db,
		Col:                   col,
		OrdersCol:             ordersCol,
		DronesCol:             dronesCol,
		NoFlyCol:              noFlyCol,
		StationsCol:           stationsCol,
		MaintCol:              maintCol,
		DeliveryCol:           deliveryCol,
		TxCol:                 txCol,
		WalletCol:             walletCol,
		BusinessesCol:         businessesCol,
		WarehousesCol:         warehousesCol,
		InventoryCol:          inventoryCol,
		AnalyticsSnapshotsCol: analyticsSnapshotsCol,
		FailureEventsCol:      failureEventsCol,
		WorkflowsCol:          workflowsCol,
	}, nil
}

