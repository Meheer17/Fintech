package db

import (
	"context"
	"math"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	models "github.com/RevenueIQ/revenueiq_dev_kit/pkg/models"
)

// CreateUser inserts a new user document with initialized names, status, and timestamps.
func (m *MongoDatabase) CreateUser(ctx context.Context, doc *models.UserDoc) (*models.UserDoc, error) {
	doc.ID = primitive.NewObjectID()
	doc.Name = doc.FirstName + " " + doc.LastName
	now := time.Now().Unix()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	doc.Active = true // Newly created users are active by default
	
	_, err := m.Col.InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// GetUser retrieves a user document by its hex ID.
func (m *MongoDatabase) GetUser(ctx context.Context, id string) (*models.UserDoc, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var doc models.UserDoc
	err = m.Col.FindOne(ctx, bson.M{"_id": objID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetUserByEmail retrieves a user document by email.
func (m *MongoDatabase) GetUserByEmail(ctx context.Context, email string) (*models.UserDoc, error) {
	var doc models.UserDoc
	err := m.Col.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// UpdateUser updates an existing user document by merging only the provided fields.
func (m *MongoDatabase) UpdateUser(ctx context.Context, id string, updateFields map[string]interface{}) (*models.UserDoc, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Fetch existing user to calculate name if first_name or last_name changes
	existing, err := m.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	fnUpdated := false
	lnUpdated := false
	if val, ok := updateFields["first_name"]; ok {
		existing.FirstName = val.(string)
		fnUpdated = true
	}
	if val, ok := updateFields["last_name"]; ok {
		existing.LastName = val.(string)
		lnUpdated = true
	}
	if fnUpdated || lnUpdated {
		updateFields["name"] = existing.FirstName + " " + existing.LastName
	}

	updateFields["updated_at"] = time.Now().Unix()

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": updateFields,
	}

	_, err = m.Col.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return m.GetUser(ctx, id)
}

// DeleteUser deletes a user document by its hex ID.
func (m *MongoDatabase) DeleteUser(ctx context.Context, id string) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}

	res, err := m.Col.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return false, err
	}

	return res.DeletedCount > 0, nil
}

// QueryUsers queries users supporting pagination, text search on name/email, and active status filters.
// It also accepts dynamic key-value string filters for any other database document field.
func (m *MongoDatabase) QueryUsers(ctx context.Context, page, limit int32, search string, active bool, filters map[string]string) ([]*models.UserDoc, int64, int32, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	filter := bson.M{}
	
	// Apply active filter
	filter["active"] = active

	// Apply search filter (name or email regex match)
	if search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	// Apply dynamic key-value filters
	for k, v := range filters {
		// Skip reserved keys if they happen to be passed in the map
		if k == "page" || k == "limit" || k == "search" || k == "active" {
			continue
		}
		
		// Attempt dynamic type conversions
		if v == "true" {
			filter[k] = true
		} else if v == "false" {
			filter[k] = false
		} else if valInt, err := strconv.Atoi(v); err == nil {
			filter[k] = int32(valInt)
		} else {
			filter[k] = v
		}
	}

	// Count total documents matching the query
	totalCount, err := m.Col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, 0, err
	}

	// Pagination parameters
	findOpts := options.Find()
	findOpts.SetSkip(int64((page - 1) * limit))
	findOpts.SetLimit(int64(limit))
	findOpts.SetSort(bson.M{"created_at": -1}) // Sort newest first

	cursor, err := m.Col.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, 0, err
	}
	defer cursor.Close(ctx)

	var docs []*models.UserDoc
	for cursor.Next(ctx) {
		var doc models.UserDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, 0, err
		}
		docs = append(docs, &doc)
	}

	totalPages := int32(math.Ceil(float64(totalCount) / float64(limit)))
	if totalPages == 0 && totalCount > 0 {
		totalPages = 1
	}

	return docs, totalCount, totalPages, nil
}
