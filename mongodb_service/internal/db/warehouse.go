package db

import (
	"context"
	"fmt"
	"time"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDatabase) SaveBusiness(ctx context.Context, biz *pb.BusinessData) (*pb.BusinessData, error) {
	now := time.Now().Unix()
	if biz.RegisteredAt == 0 {
		biz.RegisteredAt = now
	}
	if biz.BusinessId == "" {
		biz.BusinessId = fmt.Sprintf("biz-%d", now)
	}

	doc := bson.M{
		"_id":           biz.BusinessId,
		"name":          biz.Name,
		"owner_email":   biz.OwnerEmail,
		"category":      biz.Category,
		"registered_at": biz.RegisteredAt,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := m.BusinessesCol.ReplaceOne(ctx, bson.M{"_id": biz.BusinessId}, doc, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save business: %w", err)
	}

	return biz, nil
}

func (m *MongoDatabase) SaveWarehouse(ctx context.Context, wh *pb.WarehouseData) (*pb.WarehouseData, error) {
	now := time.Now().Unix()
	if wh.RegisteredAt == 0 {
		wh.RegisteredAt = now
	}
	if wh.WarehouseId == "" {
		wh.WarehouseId = fmt.Sprintf("wh-%d", now)
	}

	doc := bson.M{
		"_id":           wh.WarehouseId,
		"business_id":   wh.BusinessId,
		"name":          wh.Name,
		"address":       wh.Address,
		"latitude":      wh.Latitude,
		"longitude":     wh.Longitude,
		"registered_at": wh.RegisteredAt,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := m.WarehousesCol.ReplaceOne(ctx, bson.M{"_id": wh.WarehouseId}, doc, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save warehouse: %w", err)
	}

	return wh, nil
}

func (m *MongoDatabase) SaveInventory(ctx context.Context, item *pb.MongoInventoryData) (*pb.MongoInventoryData, error) {
	now := time.Now().Unix()
	if item.UpdatedAt == 0 {
		item.UpdatedAt = now
	}
	if item.Sku == "" {
		item.Sku = fmt.Sprintf("sku-%d", now)
	}

	key := fmt.Sprintf("%s_%s_%s", item.BusinessId, item.WarehouseId, item.Sku)
	doc := bson.M{
		"_id":          key,
		"sku":          item.Sku,
		"business_id":  item.BusinessId,
		"warehouse_id": item.WarehouseId,
		"name":         item.Name,
		"description":  item.Description,
		"quantity":     item.Quantity,
		"weight_kg":    item.WeightKg,
		"length_cm":    item.LengthCm,
		"width_cm":     item.WidthCm,
		"height_cm":    item.HeightCm,
		"image_url":    item.ImageUrl,
		"image_s3_key": item.ImageS3Key,
		"updated_at":   item.UpdatedAt,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := m.InventoryCol.ReplaceOne(ctx, bson.M{"_id": key}, doc, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save inventory: %w", err)
	}

	return item, nil
}

func (m *MongoDatabase) GetInventory(ctx context.Context, businessID, warehouseID, sku string) ([]*pb.MongoInventoryData, error) {
	filter := bson.M{}
	if businessID != "" {
		filter["business_id"] = businessID
	}
	if warehouseID != "" {
		filter["warehouse_id"] = warehouseID
	}
	if sku != "" {
		filter["sku"] = sku
	}

	cursor, err := m.InventoryCol.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventory: %w", err)
	}
	defer cursor.Close(ctx)

	var items []*pb.MongoInventoryData
	for cursor.Next(ctx) {
		var doc struct {
			SKU         string  `bson:"sku"`
			BusinessID  string  `bson:"business_id"`
			WarehouseID string  `bson:"warehouse_id"`
			Name        string  `bson:"name"`
			Description string  `bson:"description"`
			Quantity    int32   `bson:"quantity"`
			WeightKg    float64 `bson:"weight_kg"`
			LengthCm    float64 `bson:"length_cm"`
			WidthCm     float64 `bson:"width_cm"`
			HeightCm    float64 `bson:"height_cm"`
			ImageURL    string  `bson:"image_url"`
			ImageS3Key  string  `bson:"image_s3_key"`
			UpdatedAt   int64   `bson:"updated_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		items = append(items, &pb.MongoInventoryData{
			Sku:         doc.SKU,
			BusinessId:  doc.BusinessID,
			WarehouseId: doc.WarehouseID,
			Name:        doc.Name,
			Description: doc.Description,
			Quantity:    doc.Quantity,
			WeightKg:    doc.WeightKg,
			LengthCm:    doc.LengthCm,
			WidthCm:     doc.WidthCm,
			HeightCm:    doc.HeightCm,
			ImageUrl:    doc.ImageURL,
			ImageS3Key:  doc.ImageS3Key,
			UpdatedAt:   doc.UpdatedAt,
		})
	}
	return items, nil
}
