package server

import (
	"context"

	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/models"
	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateMongoOrderRequest) (*pb.MongoOrderResponse, error) {
	if req.GetOrder() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "order is required")
	}

	o := req.GetOrder()
	orderDoc := convertOrderProtoToDoc(o)

	res, err := s.db.CreateOrder(ctx, orderDoc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create order in db: %v", err)
	}

	return &pb.MongoOrderResponse{
		Success: true,
		Message: "Order created in MongoDB",
		Order:   convertOrderDocToProto(res),
	}, nil
}

func (s *Server) GetOrder(ctx context.Context, req *pb.GetMongoOrderRequest) (*pb.MongoOrderResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "order_id is required")
	}

	doc, err := s.db.GetOrderByID(ctx, req.GetOrderId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get order: %v", err)
	}
	if doc == nil {
		return nil, status.Errorf(codes.NotFound, "order not found: %s", req.GetOrderId())
	}

	return &pb.MongoOrderResponse{
		Success: true,
		Message: "Order retrieved from MongoDB",
		Order:   convertOrderDocToProto(doc),
	}, nil
}

func (s *Server) UpdateOrderStatus(ctx context.Context, req *pb.UpdateMongoOrderStatusRequest) (*pb.MongoOrderResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "order_id is required")
	}

	doc, err := s.db.UpdateOrderStatus(ctx, req.GetOrderId(), req.GetStatus(), req.GetAssignedDroneId(), req.GetNote())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update order status in DB: %v", err)
	}
	if doc == nil {
		return nil, status.Errorf(codes.NotFound, "order not found: %s", req.GetOrderId())
	}

	return &pb.MongoOrderResponse{
		Success: true,
		Message: "Order status updated in MongoDB",
		Order:   convertOrderDocToProto(doc),
	}, nil
}

func (s *Server) ListOrders(ctx context.Context, req *pb.ListMongoOrdersRequest) (*pb.ListMongoOrdersResponse, error) {
	docs, err := s.db.ListOrders(ctx, req.GetCustomerId(), req.GetStatus(), req.GetLimit())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list orders from DB: %v", err)
	}

	var protoOrders []*pb.OrderData
	for _, doc := range docs {
		protoOrders = append(protoOrders, convertOrderDocToProto(&doc))
	}

	return &pb.ListMongoOrdersResponse{
		Orders:     protoOrders,
		TotalCount: int32(len(protoOrders)),
	}, nil
}

func convertOrderProtoToDoc(p *pb.OrderData) *models.OrderDoc {
	if p == nil {
		return &models.OrderDoc{}
	}
	doc := &models.OrderDoc{
		OrderID:         p.GetOrderId(),
		CustomerID:      p.GetCustomerId(),
		Items:           p.GetItems(),
		TotalAmount:     p.GetTotalAmount(),
		Status:          p.GetStatus(),
		AssignedDroneID: p.GetAssignedDroneId(),
		CreatedAt:       p.GetCreatedAt(),
		UpdatedAt:       p.GetUpdatedAt(),
	}
	if p.GetDeliveryAddress() != nil {
		doc.DeliveryAddress = models.AddressDoc{
			Street:    p.GetDeliveryAddress().GetStreet(),
			City:      p.GetDeliveryAddress().GetCity(),
			State:     p.GetDeliveryAddress().GetState(),
			ZipCode:   p.GetDeliveryAddress().GetZipCode(),
			Country:   p.GetDeliveryAddress().GetCountry(),
			Latitude:  p.GetDeliveryAddress().GetLatitude(),
			Longitude: p.GetDeliveryAddress().GetLongitude(),
		}
	}
	if p.GetPickupDetails() != nil {
		pd := p.GetPickupDetails()
		doc.PickupDetails = models.PickupDetailsDoc{
			SenderName:   pd.GetSenderName(),
			ContactPhone: pd.GetContactPhone(),
			PickupTime:   pd.GetPickupTime(),
			Instructions: pd.GetInstructions(),
		}
		if pd.GetAddress() != nil {
			doc.PickupDetails.Address = models.AddressDoc{
				Street:    pd.GetAddress().GetStreet(),
				City:      pd.GetAddress().GetCity(),
				State:     pd.GetAddress().GetState(),
				ZipCode:   pd.GetAddress().GetZipCode(),
				Country:   pd.GetAddress().GetCountry(),
				Latitude:  pd.GetAddress().GetLatitude(),
				Longitude: pd.GetAddress().GetLongitude(),
			}
		}
	}
	for _, h := range p.GetHistory() {
		doc.History = append(doc.History, models.OrderHistoryItemDoc{
			Status:    h.GetStatus(),
			Timestamp: h.GetTimestamp(),
			Note:      h.GetNote(),
		})
	}
	return doc
}

func convertOrderDocToProto(doc *models.OrderDoc) *pb.OrderData {
	if doc == nil {
		return nil
	}
	var history []*pb.OrderHistoryData
	for _, h := range doc.History {
		history = append(history, &pb.OrderHistoryData{
			Status:    h.Status,
			Timestamp: h.Timestamp,
			Note:      h.Note,
		})
	}
	return &pb.OrderData{
		OrderId:     doc.OrderID,
		CustomerId:  doc.CustomerID,
		Items:       doc.Items,
		TotalAmount: doc.TotalAmount,
		DeliveryAddress: &pb.AddressData{
			Street:    doc.DeliveryAddress.Street,
			City:      doc.DeliveryAddress.City,
			State:     doc.DeliveryAddress.State,
			ZipCode:   doc.DeliveryAddress.ZipCode,
			Country:   doc.DeliveryAddress.Country,
			Latitude:  doc.DeliveryAddress.Latitude,
			Longitude: doc.DeliveryAddress.Longitude,
		},
		PickupDetails: &pb.PickupDetailsData{
			SenderName:   doc.PickupDetails.SenderName,
			ContactPhone: doc.PickupDetails.ContactPhone,
			Address: &pb.AddressData{
				Street:    doc.PickupDetails.Address.Street,
				City:      doc.PickupDetails.Address.City,
				State:     doc.PickupDetails.Address.State,
				ZipCode:   doc.PickupDetails.Address.ZipCode,
				Country:   doc.PickupDetails.Address.Country,
				Latitude:  doc.PickupDetails.Address.Latitude,
				Longitude: doc.PickupDetails.Address.Longitude,
			},
			PickupTime:   doc.PickupDetails.PickupTime,
			Instructions: doc.PickupDetails.Instructions,
		},
		Status:          doc.Status,
		AssignedDroneId: doc.AssignedDroneID,
		History:         history,
		CreatedAt:       doc.CreatedAt,
		UpdatedAt:       doc.UpdatedAt,
	}
}
