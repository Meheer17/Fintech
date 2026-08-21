package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// AddressDoc represents an address in MongoDB.
type AddressDoc struct {
	Street    string  `bson:"street"`
	City      string  `bson:"city"`
	State     string  `bson:"state"`
	ZipCode   string  `bson:"zip_code"`
	Country   string  `bson:"country"`
	Latitude  float64 `bson:"latitude"`
	Longitude float64 `bson:"longitude"`
}

// PickupDetailsDoc represents pickup information.
type PickupDetailsDoc struct {
	SenderName   string     `bson:"sender_name"`
	ContactPhone string     `bson:"contact_phone"`
	Address      AddressDoc `bson:"address"`
	PickupTime   int64      `bson:"pickup_time"`
	Instructions string     `bson:"instructions"`
}

// OrderHistoryItemDoc represents an entry in order history.
type OrderHistoryItemDoc struct {
	Status    string `bson:"status"`
	Timestamp int64  `bson:"timestamp"`
	Note      string `bson:"note"`
}

// OrderDoc represents an order document stored in MongoDB.
type OrderDoc struct {
	ID              primitive.ObjectID    `bson:"_id,omitempty"`
	OrderID         string                `bson:"order_id"`
	CustomerID      string                `bson:"customer_id"`
	Items           []string              `bson:"items"`
	TotalAmount     float64               `bson:"total_amount"`
	DeliveryAddress AddressDoc            `bson:"delivery_address"`
	PickupDetails   PickupDetailsDoc      `bson:"pickup_details"`
	Status          string                `bson:"status"`
	AssignedDroneID string                `bson:"assigned_drone_id"`
	History         []OrderHistoryItemDoc `bson:"history"`
	CreatedAt       int64                 `bson:"created_at"`
	UpdatedAt       int64                 `bson:"updated_at"`
}
