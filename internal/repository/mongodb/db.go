package mongodb

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB represents a MongoDB database connection
type DB struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewDB creates a new MongoDB database connection
func NewDB(ctx context.Context, uri string, dbName string) (*DB, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to MongoDB")
	}

	// Ping the database to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "failed to ping MongoDB")
	}

	return &DB{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

// Close closes the database connection
func (db *DB) Close(ctx context.Context) error {
	return db.client.Disconnect(ctx)
}

// GetDatabase returns the MongoDB database instance
func (db *DB) GetDatabase() *mongo.Database {
	return db.db
}

// GetClient returns the MongoDB client instance
func (db *DB) GetClient() *mongo.Client {
	return db.client
}
