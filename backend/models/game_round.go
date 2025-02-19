package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type GameRoundPlayer struct {
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name        string             `bson:"name" json:"name"`
	Avatar      string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	Order       int                `bson:"order" json:"order"`
	Color       *string            `bson:"color,omitempty" json:"color,omitempty"`
	IsModerator bool               `bson:"is_moderator" json:"is_moderator"`
	TeamName    *string            `bson:"team_name,omitempty" json:"team_name,omitempty"`
	TeamColor   *string            `bson:"team_color,omitempty" json:"team_color,omitempty"`
	Score       float64            `bson:"score,omitempty" json:"score,omitempty"`
	Position    int                `bson:"position,omitempty" json:"position,omitempty"`
}

type TeamScore struct {
	Name     string  `bson:"name" json:"name"`
	Color    string  `bson:"color" json:"color"`
	Score    float64 `bson:"score" json:"score"`
	Position int     `bson:"position" json:"position"`
}

type GameType string

const (
	GameTypeClassic           GameType = "classic"
	GameTypeMafia             GameType = "mafia"
	GameTypeCustom            GameType = "custom"
	GameTypeCooperative       GameType = "cooperative"
	GameTypeCoopWithModerator GameType = "cooperative_with_moderator"
	GameTypeTeamVsTeam        GameType = "team_vs_team"
)

type GameRound struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	Type       GameType           `bson:"type" json:"type"`
	StartTime  time.Time          `bson:"start_time" json:"start_time"`
	EndTime    time.Time          `bson:"end_time" json:"end_time"`
	Players    []GameRoundPlayer  `bson:"players" json:"players"`
	TeamScores []TeamScore        `bson:"team_scores,omitempty" json:"team_scores,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
