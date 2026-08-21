package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// LocationDoc represents latitude and longitude coordinates.
type LocationDoc struct {
	Latitude  float64 `bson:"latitude"`
	Longitude float64 `bson:"longitude"`
}

// DroneDoc represents a drone record stored in MongoDB.
type DroneDoc struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	DroneID           string             `bson:"drone_id"`
	Model             string             `bson:"model"`
	Status            string             `bson:"status"`        // AVAILABLE, RESERVED, DELIVERING, RETURNING, CHARGING, MAINTENANCE, OFFLINE
	BatteryLevel      float64            `bson:"battery_level"` // 0.0 - 100.0
	PayloadCapacityKg float64            `bson:"payload_capacity_kg"`
	CurrentLocation   LocationDoc        `bson:"current_location"`
	CurrentOrderID    string             `bson:"current_order_id"`
	RegisteredAt      int64              `bson:"registered_at"`
	LastUpdatedAt     int64              `bson:"last_updated_at"`
}
