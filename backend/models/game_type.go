package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ScoringType string

const (
	ScoringTypeClassic           ScoringType = "classic"
	ScoringTypeMafia             ScoringType = "mafia"
	ScoringTypeCustom            ScoringType = "custom"
	ScoringTypeCooperative       ScoringType = "cooperative"
	ScoringTypeCoopWithModerator ScoringType = "cooperative_with_moderator"
	ScoringTypeTeamVsTeam        ScoringType = "team_vs_team"
)

type Label struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name  string             `bson:"name" json:"name"`
	Color string             `bson:"color" json:"color"`
	Icon  string             `bson:"icon" json:"icon"`
}

type GameType struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	Icon       string             `bson:"icon" json:"icon"`
	Labels     []Label            `bson:"labels" json:"labels"`
	Teams      []Label            `bson:"teams" json:"teams"`
	MinPlayers int                `bson:"min_players" json:"min_players"`
	MaxPlayers int                `bson:"max_players" json:"max_players"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
