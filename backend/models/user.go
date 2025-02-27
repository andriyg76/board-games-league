package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Version     int64              `bson:"version"`
	ExternalIDs []string           `bson:"external_ids"`
	Name        string             `bson:"name"`
	Avatar      string             `bson:"picture,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	Alias       string             `bson:"alias"`
}
