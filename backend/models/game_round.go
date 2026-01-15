package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// GameRoundStatus визначає статус раунду гри
type GameRoundStatus string

const (
	// StatusPlayersSelected - гравці обрані, гра не почалась
	StatusPlayersSelected GameRoundStatus = "players_selected"
	// StatusInProgress - гра в процесі (ролі призначаються)
	StatusInProgress GameRoundStatus = "in_progress"
	// StatusScoring - введення очок
	StatusScoring GameRoundStatus = "scoring"
	// StatusCompleted - гра завершена
	StatusCompleted GameRoundStatus = "completed"
)

// ValidGameRoundStatuses - список валідних статусів
var ValidGameRoundStatuses = []GameRoundStatus{
	StatusPlayersSelected,
	StatusInProgress,
	StatusScoring,
	StatusCompleted,
}

// IsValidStatus перевіряє чи є статус валідним
func (s GameRoundStatus) IsValidStatus() bool {
	for _, valid := range ValidGameRoundStatuses {
		if s == valid {
			return true
		}
	}
	return false
}

type GameRoundPlayer struct {
	IsModerator   bool               `bson:"is_moderator" json:"is_moderator"`
	TeamName      string             `bson:"team_name,omitempty" json:"team_name,omitempty"`
	LabelName     string             `bson:"label_name,omitempty" json:"label_name,omitempty"`
	Score         int64              `bson:"cooperative_score,omitempty" json:"score,omitempty"`
	Position      int                `bson:"position,omitempty" json:"position,omitempty"`
	MembershipID primitive.ObjectID `bson:"membership_id,omitempty" json:"-"`
	MembershipCode string            `bson:"-" json:"membership_code,omitempty"` // Populated from MembershipID
	// Deprecated: use MembershipID instead. Kept for backward compatibility during migration.
	PlayerID primitive.ObjectID `bson:"player_id,omitempty" json:"-"`
}

type TeamScore struct {
	Name     string `bson:"name" json:"name"`
	Score    int64  `bson:"cooperative_score,omitempty" json:"score,omitempty"`
	Position int    `bson:"position" json:"position"`
}

type GameRound struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"-"` // Never expose ObjectID - use Code instead
	Code             string             `bson:"-" json:"code"`
	Version          int64              `bson:"version" json:"version,omitempty"`
	LeagueID         primitive.ObjectID `bson:"league_id,omitempty" json:"-"`
	Name             string             `bson:"name" json:"name"`
	GameTypeID       primitive.ObjectID `bson:"game_type_id,omitempty" json:"-"`
	GameType         string             `bson:"-" json:"game_type,omitempty"` // Game type key or code (populated from GameTypeID)
	Status           GameRoundStatus    `bson:"status" json:"status"`
	StartTime        time.Time          `bson:"start_time" json:"start_time"`
	EndTime          time.Time          `bson:"end_time" json:"end_time,omitempty"`
	Players          []GameRoundPlayer  `bson:"players" json:"players"`
	TeamScores       []TeamScore        `bson:"team_scores,omitempty" json:"team_scores,omitempty"`
	CooperativeScore int64              `bson:"cooperative_score,omitempty" json:"cooperative_score,omitempty"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}
