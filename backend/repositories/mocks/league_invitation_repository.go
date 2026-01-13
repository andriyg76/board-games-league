package mocks

import (
	"context"
	"time"

	"github.com/andriyg76/bgl/models"
	mock2 "github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockLeagueInvitationRepository is a mock implementation of LeagueInvitationRepository
type MockLeagueInvitationRepository struct {
	mock2.Mock
}

func (m *MockLeagueInvitationRepository) Create(ctx context.Context, invitation *models.LeagueInvitation) error {
	args := m.Called(ctx, invitation)
	return args.Error(0)
}

func (m *MockLeagueInvitationRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.LeagueInvitation, error) {
	args := m.Called(ctx, id)
	inv := args.Get(0)
	if inv == nil {
		return nil, args.Error(1)
	}
	return inv.(*models.LeagueInvitation), args.Error(1)
}

func (m *MockLeagueInvitationRepository) FindByToken(ctx context.Context, token string) (*models.LeagueInvitation, error) {
	args := m.Called(ctx, token)
	inv := args.Get(0)
	if inv == nil {
		return nil, args.Error(1)
	}
	return inv.(*models.LeagueInvitation), args.Error(1)
}

func (m *MockLeagueInvitationRepository) FindActiveByCreator(ctx context.Context, leagueID, createdBy primitive.ObjectID) ([]*models.LeagueInvitation, error) {
	args := m.Called(ctx, leagueID, createdBy)
	inv := args.Get(0)
	if inv == nil {
		return nil, args.Error(1)
	}
	return inv.([]*models.LeagueInvitation), args.Error(1)
}

func (m *MockLeagueInvitationRepository) FindExpiredByCreator(ctx context.Context, leagueID, createdBy primitive.ObjectID) ([]*models.LeagueInvitation, error) {
	args := m.Called(ctx, leagueID, createdBy)
	inv := args.Get(0)
	if inv == nil {
		return nil, args.Error(1)
	}
	return inv.([]*models.LeagueInvitation), args.Error(1)
}

func (m *MockLeagueInvitationRepository) MarkAsUsed(ctx context.Context, id primitive.ObjectID, usedBy primitive.ObjectID) error {
	args := m.Called(ctx, id, usedBy)
	return args.Error(0)
}

func (m *MockLeagueInvitationRepository) Cancel(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLeagueInvitationRepository) Extend(ctx context.Context, id primitive.ObjectID, duration time.Duration) error {
	args := m.Called(ctx, id, duration)
	return args.Error(0)
}


