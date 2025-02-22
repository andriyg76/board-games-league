package models

import (
	"github.com/andriyg76/bgl/utils"
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
	Name  string `bson:"name" json:"name"`
	Color string `bson:"color" json:"color"`
	Icon  string `bson:"icon" json:"icon"`
}

type GameType struct {
	utils.IdCode
	Version    int64     `bson:"version" json:"version"`
	Name       string    `bson:"name" json:"name"`
	Icon       string    `bson:"icon" json:"icon"`
	Labels     []Label   `bson:"labels" json:"labels"`
	Teams      []Label   `bson:"teams" json:"teams"`
	MinPlayers int       `bson:"min_players" json:"min_players"`
	MaxPlayers int       `bson:"max_players" json:"max_players"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}
