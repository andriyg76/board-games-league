package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type NotificationType string

const (
	NotificationLeagueJoin NotificationType = "league_join"
	NotificationLeagueBan  NotificationType = "league_ban"
)

type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Type      NotificationType   `bson:"type"`
	Title     string             `bson:"title"`
	Message   string             `bson:"message"`
	LeagueID  primitive.ObjectID `bson:"league_id,omitempty"`
	IsRead    bool               `bson:"is_read"`
	CreatedAt time.Time          `bson:"created_at"`
}
