package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"locations/internal/models"
)

type MongoDB struct {
	client *mongo.Client
}

func NewMongoDB(uri string) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB")

	return &MongoDB{client: client}, nil
}

func (c *MongoDB) InsertLocationUpdate(ctx context.Context, update models.LocationUpdate) error {
	collection := c.client.Database("database").Collection("locations")

	_, err := collection.InsertOne(ctx, update)
	return err
}

func (c *MongoDB) Close() error {
	return c.client.Disconnect(context.TODO())
}
