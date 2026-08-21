package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// UserDoc represents the MongoDB document structure for a user with rich profile metadata.
type UserDoc struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	FirstName     string             `bson:"first_name"`
	LastName      string             `bson:"last_name"`
	Name          string             `bson:"name"`
	Email         string             `bson:"email"`
	Age           int32              `bson:"age"`
	Password      string             `bson:"password"`
	Role          string             `bson:"role"`
	PhoneNumber   string             `bson:"phone_number"`
	EmailVerified bool               `bson:"email_verified"`
	PhoneVerified bool               `bson:"phone_verified"`
	Active        bool               `bson:"active"`
	CreatedAt     int64              `bson:"created_at"`
	UpdatedAt     int64              `bson:"updated_at"`
	LastLoginAt   int64              `bson:"last_login_at"`
}
