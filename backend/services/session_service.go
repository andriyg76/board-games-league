package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/andriyg76/glog"
	"github.com/andriyg76/hexerr"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SessionService interface {
	CreateSession(ctx context.Context, userID primitive.ObjectID, userCode string, externalIDs []string, name, avatar string, ipAddress, userAgent string) (rotateToken, actionToken string, err error)
	RefreshActionToken(ctx context.Context, rotateToken, ipAddress, userAgent string) (newRotateToken, actionToken string, err error)
	InvalidateSession(ctx context.Context, rotateToken string) error
	CleanupExpiredSessions(ctx context.Context) error
}

type sessionService struct {
	sessionRepository repositories.SessionRepository
	userRepository    repositories.UserRepository
}

func NewSessionService(sessionRepository repositories.SessionRepository, userRepository repositories.UserRepository) SessionService {
	return &sessionService{
		sessionRepository: sessionRepository,
		userRepository:    userRepository,
	}
}

func generateRotateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}

func (s *sessionService) shouldRotate(session *models.Session) bool {
	// Rotate if 12 hours have passed since last rotation
	return time.Since(session.LastRotationAt) >= 12*time.Hour
}

func (s *sessionService) CreateSession(ctx context.Context, userID primitive.ObjectID, userCode string, externalIDs []string, name, avatar string, ipAddress, userAgent string) (rotateToken, actionToken string, err error) {
	// Generate rotate token
	rotateToken, err = generateRotateToken()
	if err != nil {
		return "", "", hexerr.Wrapf(err, "failed to generate rotate token")
	}

	// Create action token (1 hour expiry)
	actionToken, err = user_profile.CreateAuthTokenWithExpiry(externalIDs, userCode, name, avatar, 1*time.Hour)
	if err != nil {
		return "", "", hexerr.Wrapf(err, "failed to create action token")
	}

	now := time.Now()
	session := &models.Session{
		RotateToken:    rotateToken,
		UserID:         userID,
		CreatedAt:      now,
		UpdatedAt:      now,
		LastRotationAt: now,
		ExpiresAt:      now.Add(30 * 24 * time.Hour), // 30 days
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		Version:        1,
	}

	if err := s.sessionRepository.Create(ctx, session); err != nil {
		return "", "", hexerr.Wrapf(err, "failed to create session")
	}

	// Update user's last activity
	if user, err := s.userRepository.FindByID(ctx, userID); err == nil && user != nil {
		user.LastActivity = now
		if updateErr := s.userRepository.Update(ctx, user); updateErr != nil {
			glog.Warn("Failed to update user last activity: %v", updateErr)
		}
	}

	return rotateToken, actionToken, nil
}

func (s *sessionService) RefreshActionToken(ctx context.Context, rotateToken, ipAddress, userAgent string) (newRotateToken, actionToken string, err error) {
	// Find session by rotate token
	session, err := s.sessionRepository.FindByRotateToken(ctx, rotateToken)
	if err != nil {
		return "", "", hexerr.Wrapf(err, "failed to find session")
	}
	if session == nil {
		return "", "", hexerr.New("session not found")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		_ = s.sessionRepository.Delete(ctx, rotateToken)
		return "", "", hexerr.New("session expired")
	}

	// Get user for creating action token
	user, err := s.userRepository.FindByID(ctx, session.UserID)
	if err != nil {
		return "", "", hexerr.Wrapf(err, "failed to find user")
	}
	if user == nil {
		return "", "", hexerr.New("user not found")
	}

	// Update session tracking
	session.IPAddress = ipAddress
	session.UserAgent = userAgent
	session.UpdatedAt = time.Now()

	// Check if we need to rotate the rotate token
	shouldRotate := s.shouldRotate(session)
	if shouldRotate {
		// Generate new rotate token
		newRotateToken, err = generateRotateToken()
		if err != nil {
			return "", "", hexerr.Wrapf(err, "failed to generate new rotate token")
		}

		// Delete old session
		if err := s.sessionRepository.Delete(ctx, rotateToken); err != nil {
			return "", "", hexerr.Wrapf(err, "failed to delete old session")
		}

		// Create new session with new rotate token
		session.RotateToken = newRotateToken
		session.LastRotationAt = time.Now()
		session.CreatedAt = time.Now()
		session.Version = 1

		if err := s.sessionRepository.Create(ctx, session); err != nil {
			return "", "", hexerr.Wrapf(err, "failed to create new session")
		}
	} else {
		// Update existing session
		if err := s.sessionRepository.Update(ctx, session); err != nil {
			// Handle optimistic locking failure - retry once
			if err.Error() == "concurrent modification detected" {
				// Retry: re-read session and update
				session, retryErr := s.sessionRepository.FindByRotateToken(ctx, rotateToken)
				if retryErr != nil || session == nil {
					return "", "", hexerr.Wrapf(retryErr, "failed to retry update")
				}
				session.IPAddress = ipAddress
				session.UserAgent = userAgent
				session.UpdatedAt = time.Now()
				if retryErr := s.sessionRepository.Update(ctx, session); retryErr != nil {
					return "", "", hexerr.Wrapf(retryErr, "failed to update session after retry")
				}
			} else {
				return "", "", hexerr.Wrapf(err, "failed to update session")
			}
		}
	}

	// Create new action token (1 hour expiry)
	userCode := utils.IdToCode(user.ID)
	actionToken, err = user_profile.CreateAuthTokenWithExpiry(user.ExternalIDs, userCode, user.Name, user.Avatar, 1*time.Hour)
	if err != nil {
		return "", "", hexerr.Wrapf(err, "failed to create action token")
	}

	// Update user's last activity
	user.LastActivity = time.Now()
	if updateErr := s.userRepository.Update(ctx, user); updateErr != nil {
		glog.Warn("Failed to update user last activity: %v", updateErr)
	}

	return newRotateToken, actionToken, nil
}

func (s *sessionService) InvalidateSession(ctx context.Context, rotateToken string) error {
	return s.sessionRepository.Delete(ctx, rotateToken)
}

func (s *sessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepository.DeleteExpired(ctx)
}
