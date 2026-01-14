package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type LeagueMembershipStatus string

const (
	MembershipActive  LeagueMembershipStatus = "active"
	MembershipBanned  LeagueMembershipStatus = "banned"
	MembershipPending LeagueMembershipStatus = "pending"
	MembershipVirtual LeagueMembershipStatus = "virtual"
)

// RecentCoPlayer represents a player who recently played with this member
type RecentCoPlayer struct {
	MembershipID primitive.ObjectID `bson:"membership_id"`
	LastPlayedAt time.Time          `bson:"last_played_at"`
}

// MaxRecentCoPlayers is the maximum number of recent co-players to cache
const MaxRecentCoPlayers = 10

type LeagueMembership struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty"`
	Version         int64                  `bson:"version"`
	LeagueID        primitive.ObjectID     `bson:"league_id"`
	UserID          primitive.ObjectID     `bson:"user_id,omitempty"`
	InvitationID    primitive.ObjectID     `bson:"invitation_id,omitempty"`
	Alias           string                 `bson:"alias,omitempty"`
	Status          LeagueMembershipStatus `bson:"status"`
	JoinedAt        time.Time              `bson:"joined_at"`
	CreatedAt       time.Time              `bson:"created_at"`
	UpdatedAt       time.Time              `bson:"updated_at"`
	LastActivityAt  time.Time              `bson:"last_activity_at,omitempty"`  // Last game or invitation activity
	RecentCoPlayers []RecentCoPlayer       `bson:"recent_co_players,omitempty"` // Max 10 recent co-players
}
