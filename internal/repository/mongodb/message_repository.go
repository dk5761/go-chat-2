package mongodb

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/chat-backend/internal/models"
)

type MessageRepository struct {
	collection *mongo.Collection
}

func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{
		collection: db.Collection("messages"),
	}
}

func (r *MessageRepository) Create(ctx context.Context, message *models.Message) error {
	_, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		return errors.Wrap(err, "failed to create message")
	}
	return nil
}

func (r *MessageRepository) GetByID(ctx context.Context, id string) (*models.Message, error) {
	var message models.Message
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&message)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get message")
	}
	return &message, nil
}

func (r *MessageRepository) Update(ctx context.Context, message *models.Message) error {
	result, err := r.collection.ReplaceOne(ctx, bson.M{"id": message.ID}, message)
	if err != nil {
		return errors.Wrap(err, "failed to update message")
	}
	if result.MatchedCount == 0 {
		return errors.New("message not found")
	}
	return nil
}

func (r *MessageRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return errors.Wrap(err, "failed to delete message")
	}
	if result.DeletedCount == 0 {
		return errors.New("message not found")
	}
	return nil
}

func (r *MessageRepository) GetMessagesBetween(ctx context.Context, userID1, userID2 string, limit int64, before time.Time) ([]*models.Message, error) {
	filter := bson.M{
		"$or": []bson.M{
			{
				"sender_id":    userID1,
				"recipient_id": userID2,
			},
			{
				"sender_id":    userID2,
				"recipient_id": userID1,
			},
		},
		"timestamp": bson.M{"$lt": before},
	}

	opts := options.Find().
		SetSort(bson.D{bson.E{Key: "timestamp", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get messages")
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, errors.Wrap(err, "failed to decode messages")
	}
	return messages, nil
}

func (r *MessageRepository) GetGroupMessages(ctx context.Context, groupID string, limit int, offset int) ([]models.Message, error) {
	filter := bson.M{
		"group_id": groupID,
	}

	opts := options.Find().
		SetSort(bson.D{bson.E{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get group messages")
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, errors.Wrap(err, "failed to decode messages")
	}
	return messages, nil
}

func (r *MessageRepository) MarkAsDelivered(ctx context.Context, messageID string, userID string) error {
	update := bson.M{
		"$addToSet": bson.M{
			"delivered_to": userID,
		},
	}
	result, err := r.collection.UpdateOne(ctx, bson.M{"id": messageID}, update)
	if err != nil {
		return errors.Wrap(err, "failed to mark message as delivered")
	}
	if result.MatchedCount == 0 {
		return errors.New("message not found")
	}
	return nil
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, messageID string, userID string) error {
	update := bson.M{
		"$addToSet": bson.M{
			"read_by": userID,
		},
	}
	result, err := r.collection.UpdateOne(ctx, bson.M{"id": messageID}, update)
	if err != nil {
		return errors.Wrap(err, "failed to mark message as read")
	}
	if result.MatchedCount == 0 {
		return errors.New("message not found")
	}
	return nil
}

func (r *MessageRepository) GetUserMessages(ctx context.Context, userID string, limit int, offset int) ([]models.Message, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"sender_id": userID},
			{"recipient_id": userID},
		},
	}

	opts := options.Find().
		SetSort(bson.D{bson.E{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user messages")
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, errors.Wrap(err, "failed to decode messages")
	}
	return messages, nil
}
