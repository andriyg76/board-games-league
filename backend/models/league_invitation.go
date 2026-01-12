package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type LeagueInvitation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Version   int64              `bson:"version"`
	LeagueID  primitive.ObjectID `bson:"league_id"`
	CreatedBy primitive.ObjectID `bson:"created_by"`
	Token     string             `bson:"token"`
	IsUsed    bool               `bson:"is_used"`
	UsedBy    primitive.ObjectID `bson:"used_by,omitempty"`
	UsedAt    time.Time          `bson:"used_at,omitempty"`
	ExpiresAt time.Time          `bson:"expires_at"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}
