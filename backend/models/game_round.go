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
	IsModerator  bool               `bson:"is_moderator"`
	TeamName     string             `bson:"team_name,omitempty"`
	LabelName    string             `bson:"label_name,omitempty"`
	Score        int64              `bson:"cooperative_score,omitempty"`
	Position     int                `bson:"position,omitempty"`
	MembershipID primitive.ObjectID `bson:"membership_id,omitempty"`
	// Deprecated: use MembershipID instead. Kept for backward compatibility during migration.
	PlayerID primitive.ObjectID `bson:"player_id,omitempty"`
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
	LeagueID         primitive.ObjectID `bson:"league_id,omitempty"`
	Name             string             `bson:"name"`
	GameTypeID       primitive.ObjectID `bson:"game_type_id,omitempty"`
	Status           GameRoundStatus    `bson:"status" json:"status"`
	StartTime        time.Time          `bson:"start_time"`
	EndTime          time.Time          `bson:"end_time"`
	Players          []GameRoundPlayer  `bson:"players"`
	TeamScores       []TeamScore        `bson:"team_scores,omitempty"`
	CooperativeScore int64              `bson:"cooperative_score,omitempty"`
	CreatedAt        time.Time          `bson:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at"`
}
