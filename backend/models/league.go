package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type LeagueStatus string

const (
	LeagueActive   LeagueStatus = "active"
	LeagueArchived LeagueStatus = "archived"
)

type League struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Version   int64              `bson:"version"`
	Name      string             `bson:"name"`
	Status    LeagueStatus       `bson:"status"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}
