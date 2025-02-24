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

var ScoringTypes = []ScoringType{ScoringTypeClassic, ScoringTypeCooperative, ScoringTypeCustom, ScoringTypeMafia,
	ScoringTypeCoopWithModerator, ScoringTypeTeamVsTeam}

type Label struct {
	Name  string `bson:"name"`
	Color string `bson:"color"`
	Icon  string `bson:"icon"`
}

type GameType struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Version     int64              `bson:"version"`
	ScoringType string             `bson:"scoring_type"`
	Name        string             `bson:"name"`
	Icon        string             `bson:"icon"`
	Labels      []Label            `bson:"labels"`
	Teams       []Label            `bson:"teams"`
	MinPlayers  int                `bson:"min_players"`
	MaxPlayers  int                `bson:"max_players"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}
