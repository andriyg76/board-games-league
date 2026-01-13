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

type LeagueMembership struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty"`
	Version      int64                  `bson:"version"`
	LeagueID     primitive.ObjectID     `bson:"league_id"`
	UserID       primitive.ObjectID     `bson:"user_id,omitempty"`
	InvitationID primitive.ObjectID     `bson:"invitation_id,omitempty"`
	Alias        string                 `bson:"alias,omitempty"`
	Status       LeagueMembershipStatus `bson:"status"`
	JoinedAt     time.Time              `bson:"joined_at"`
	CreatedAt    time.Time              `bson:"created_at"`
	UpdatedAt    time.Time              `bson:"updated_at"`
}
