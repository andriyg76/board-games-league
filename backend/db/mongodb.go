// db/mongodb.go
package db

import (
	"context"
	"github.com/andriyg76/glog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
	"strings"
	"time"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoDB(uri string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		client.Disconnect(ctx)
		return nil, glog.Error("database is not specified in mongo url")
	}

	return &MongoDB{
		client:   client,
		database: client.Database(dbName),
	}, nil
}

func (db *MongoDB) Collection(name string) *mongo.Collection {
	return db.database.Collection(name)
}
