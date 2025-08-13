package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type GameRoundPlayer struct {
	IsModerator bool               `bson:"is_moderator"`
	TeamName    string             `bson:"team_name,omitempty"`
	LabelName   string             `bson:"label_name,omitempty"`
	Score       int64              `bson:"cooperative_score,omitempty"`
	Position    int                `bson:"position,omitempty"`
	PlayerID    primitive.ObjectID `bson:"player_id,omitempty"`
}

type TeamScore struct {
	Name     string `bson:"name"`
	Score    int64  `bson:"cooperative_score,omitempty"`
	Position int    `bson:"position"`
}

type GameRound struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Code             string             `bson:"-"`
	Version          int64              `bson:"version"`
	Name             string             `bson:"name"`
	GameTypeID       primitive.ObjectID `bson:"game_type_id,omitempty"`
	StartTime        time.Time          `bson:"start_time"`
	EndTime          time.Time          `bson:"end_time"`
	Players          []GameRoundPlayer  `bson:"players"`
	TeamScores       []TeamScore        `bson:"team_scores,omitempty"`
	CooperativeScore int64              `bson:"cooperative_score,omitempty"`
	CreatedAt        time.Time          `bson:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at"`
}
