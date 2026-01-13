package mocks

import (
	"context"
	"github.com/andriyg76/bgl/models"
	mock2 "github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockLeagueMembershipRepository is a mock implementation of LeagueMembershipRepository
type MockLeagueMembershipRepository struct {
	mock2.Mock
}

func (m *MockLeagueMembershipRepository) Create(ctx context.Context, membership *models.LeagueMembership) error {
	args := m.Called(ctx, membership)
	return args.Error(0)
}

func (m *MockLeagueMembershipRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.LeagueMembership, error) {
	args := m.Called(ctx, id)
	membership := args.Get(0)
	if membership == nil {
		return nil, args.Error(1)
	}
	return membership.(*models.LeagueMembership), args.Error(1)
}

func (m *MockLeagueMembershipRepository) FindByLeagueAndUser(ctx context.Context, leagueID, userID primitive.ObjectID) (*models.LeagueMembership, error) {
	args := m.Called(ctx, leagueID, userID)
	membership := args.Get(0)
	if membership == nil {
		return nil, args.Error(1)
	}
	return membership.(*models.LeagueMembership), args.Error(1)
}

func (m *MockLeagueMembershipRepository) FindByLeague(ctx context.Context, leagueID primitive.ObjectID) ([]*models.LeagueMembership, error) {
	args := m.Called(ctx, leagueID)
	memberships := args.Get(0)
	if memberships == nil {
		return nil, args.Error(1)
	}
	return memberships.([]*models.LeagueMembership), args.Error(1)
}

func (m *MockLeagueMembershipRepository) FindByUser(ctx context.Context, userID primitive.ObjectID) ([]*models.LeagueMembership, error) {
	args := m.Called(ctx, userID)
	memberships := args.Get(0)
	if memberships == nil {
		return nil, args.Error(1)
	}
	return memberships.([]*models.LeagueMembership), args.Error(1)
}

func (m *MockLeagueMembershipRepository) Update(ctx context.Context, membership *models.LeagueMembership) error {
	args := m.Called(ctx, membership)
	return args.Error(0)
}

func (m *MockLeagueMembershipRepository) IsActiveMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error) {
	args := m.Called(ctx, leagueID, userID)
	return args.Bool(0), args.Error(1)
}

