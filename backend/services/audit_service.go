package services

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditAction represents the type of action being audited
type AuditAction string

const (
	AuditActionInviteCreated   AuditAction = "invite_created"
	AuditActionInviteCancelled AuditAction = "invite_cancelled"
	AuditActionInviteAccepted  AuditAction = "invite_accepted"
	AuditActionUserBanned      AuditAction = "user_banned"
	AuditActionLeagueCreated   AuditAction = "league_created"
	AuditActionLeagueArchived  AuditAction = "league_archived"
	AuditActionLeagueUnarchived AuditAction = "league_unarchived"
	AuditActionGameCreated     AuditAction = "game_created"
	AuditActionGameFinalized   AuditAction = "game_finalized"
)

// AuditTargetType represents the type of object being acted upon
type AuditTargetType string

const (
	AuditTargetLeague     AuditTargetType = "league"
	AuditTargetInvitation AuditTargetType = "invitation"
	AuditTargetUser       AuditTargetType = "user"
	AuditTargetGame       AuditTargetType = "game"
)

// AuditDetails contains additional information about the audit event
type AuditDetails map[string]interface{}

// AuditService defines the interface for logging user actions
type AuditService interface {
	// LogAction logs a user action for audit purposes
	LogAction(ctx context.Context, userID primitive.ObjectID, action AuditAction, targetType AuditTargetType, targetID primitive.ObjectID, details AuditDetails) error
}

// NoopAuditService is a no-operation implementation of AuditService
// TODO: Replace with a real implementation that stores audit logs in MongoDB
type NoopAuditService struct{}

// NewNoopAuditService creates a new no-op audit service
func NewNoopAuditService() AuditService {
	return &NoopAuditService{}
}

// LogAction does nothing in the no-op implementation
func (s *NoopAuditService) LogAction(ctx context.Context, userID primitive.ObjectID, action AuditAction, targetType AuditTargetType, targetID primitive.ObjectID, details AuditDetails) error {
	// No-op: This is a placeholder implementation
	// In a real implementation, this would save the audit log to MongoDB
	return nil
}

