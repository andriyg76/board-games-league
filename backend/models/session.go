package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Session struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	RotateToken    string             `bson:"rotate_token"`
	UserID         primitive.ObjectID `bson:"user_id"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
	LastRotationAt time.Time          `bson:"last_rotation_at"`
	ExpiresAt      time.Time          `bson:"expires_at"`
	IPAddress      string             `bson:"ip_address,omitempty"`
	UserAgent      string             `bson:"user_agent,omitempty"`
	Version        int64              `bson:"version"`
}
