package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type GameRoundPlayer struct {
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Order       int                `bson:"order" json:"order"`
	IsModerator bool               `bson:"is_moderator" json:"is_moderator"`
	TeamName    string             `bson:"team_name,omitempty" json:"team_name,omitempty"`
	LabelName   string             `bson:"label_name,omitempty" json:"label_name,omitempty"`
	Score       int64              `bson:"cooperative_score,omitempty" json:"cooperative_score,omitempty"`
	Position    int                `bson:"position,omitempty" json:"position,omitempty"`
}

type TeamScore struct {
	Name     string `bson:"name" json:"name"`
	Score    int64  `bson:"cooperative_score,omitempty" json:"cooperative_score,omitempty"`
	Position int    `bson:"position" json:"position"`
}

type GameRound struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	GameType         primitive.ObjectID `bson:"game_type_id" json:"game_type_id"`
	StartTime        time.Time          `bson:"start_time" json:"start_time"`
	EndTime          time.Time          `bson:"end_time" json:"end_time"`
	Players          []GameRoundPlayer  `bson:"players" json:"players"`
	TeamScores       []TeamScore        `bson:"team_scores,omitempty" json:"team_scores,omitempty"`
	CooperativeScore int64              `bson:"cooperative_score,omitempty" json:"cooperative_score,omitempty"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}
