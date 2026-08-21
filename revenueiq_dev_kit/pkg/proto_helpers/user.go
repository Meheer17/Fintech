package protohelpers

import (
	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/models"
	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
)

// ToProtoUser maps a MongoDB UserDoc structure to the protobuf User message structure.
func ToProtoUser(doc *models.UserDoc) *pb.User {
	if doc == nil {
		return nil
	}
	return &pb.User{
		Id:            doc.ID.Hex(),
		FirstName:     doc.FirstName,
		LastName:      doc.LastName,
		Name:          doc.Name,
		Email:         doc.Email,
		Age:           doc.Age,
		Password:      doc.Password,
		Role:          doc.Role,
		PhoneNumber:   doc.PhoneNumber,
		EmailVerified: doc.EmailVerified,
		PhoneVerified: doc.PhoneVerified,
		Active:        doc.Active,
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
		LastLoginAt:   doc.LastLoginAt,
	}
}
