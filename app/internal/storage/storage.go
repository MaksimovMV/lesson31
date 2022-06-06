package storage

import (
	"app/internal/model"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Storage struct {
	collection *mongo.Collection
}

func NewStorage(database *mongo.Database, collection string) *Storage {
	return &Storage{
		collection: database.Collection(collection),
	}
}

func (s *Storage) GetUser(targetID string) (u *model.User, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert hex to objectID, error: %w", err)
	}
	filter := bson.M{"_id": objectID}

	result := s.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to find user by id, error: %w", result.Err())
	}
	if err = result.Decode(&u); err != nil {
		return nil, fmt.Errorf("failed to decode document. error: %w", err)
	}

	return u, nil
}

func (s *Storage) PutUser(u model.User) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.collection.InsertOne(ctx, u)
	if err != nil {
		return "", fmt.Errorf("failed to create user, error: %w", err)
	}
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		return oid.Hex(), nil
	}
	return "", fmt.Errorf("failed to convert objective to hex")
}

func (s *Storage) MakeFriends(sourceID string, targetID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sOID, err := primitive.ObjectIDFromHex(sourceID)
	if err != nil {
		return fmt.Errorf("failed to convert hex to objectID, error: %w", err)
	}

	tOID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return fmt.Errorf("failed to convert hex to objectID, error %w", err)
	}

	if sOID == tOID {
		return fmt.Errorf("ID пользователей совпадают")
	}

	if _, err := s.collection.Find(ctx, bson.M{"_id": sOID}); err != nil {
		return fmt.Errorf("failed to find user id: %v, err: %v", sourceID, err)
	}
	if _, err := s.collection.Find(ctx, bson.M{"_id": tOID}); err != nil {
		return fmt.Errorf("failed to find user id: %v, err: %v", sourceID, err)
	}

	filter := bson.M{"_id": sOID}
	update := bson.M{
		"$addToSet": bson.M{"friends": targetID},
	}
	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to execute query, err: %v", err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("failed add user to friends list, id: %v", sourceID)
	}

	filter = bson.M{"_id": tOID}
	update = bson.M{
		"$addToSet": bson.M{"friends": sourceID},
	}
	result, err = s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to execute query, err: %v", err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("failed add user to friends list, id: %v", sourceID)
	}

	return nil
}

func (s *Storage) DeleteUser(targetID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return fmt.Errorf("failed to convert hex to objectID, error: %w", err)
	}

	filter := bson.M{"_id": objectID}

	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete user, error: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found, id: %v", targetID)
	}

	update := bson.M{
		"$pull": bson.M{"friends": targetID},
	}

	if _, err := s.collection.UpdateMany(ctx, bson.M{}, update); err != nil {
		return fmt.Errorf("failed to delete user from friends lists, error: %w", err)
	}

	return nil
}

func (s *Storage) GetFriends(targetID string) ([]string, error) {
	u, err := s.GetUser(targetID)
	if err != nil {
		return nil, err
	}
	return u.Friends, nil
}

func (s *Storage) Update(targetID string, u model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return fmt.Errorf("failed to convert hex to objectID, error: %w", err)
	}

	filter := bson.M{"_id": objectID}

	userByte, err := bson.Marshal(u)
	if err != nil {
		return fmt.Errorf("failed to marshal document, error: %w", err)
	}

	var updateObj bson.M
	err = bson.Unmarshal(userByte, &updateObj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal document, error: %w", err)
	}

	delete(updateObj, "_id")

	update := bson.M{
		"$set": updateObj,
	}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user, error: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found, id: %v", targetID)
	}

	return nil
}

func (s *Storage) DeleteFriend(sourceID string, targetID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sOID, err := primitive.ObjectIDFromHex(sourceID)
	if err != nil {
		return fmt.Errorf("failed to convert hex to objectID, error: %w", err)
	}

	filter := bson.M{"_id": sOID}

	update := bson.M{
		"$pull": bson.M{"friends": targetID},
	}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete user from friends lists, error: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found, id: %v", sourceID)
	}

	tOID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return fmt.Errorf("failed to convert hex to objectID, error: %w", err)
	}

	filter = bson.M{"_id": tOID}

	update = bson.M{
		"$pull": bson.M{"friends": sourceID},
	}

	result, err = s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete user from friends lists, error: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found, id: %v", targetID)
	}

	return nil
}
