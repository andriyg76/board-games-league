package mocks

import (
	"context"
	"github.com/andriyg76/bgl/models"
	mock2 "github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockLeagueRepository is a mock implementation of LeagueRepository
type MockLeagueRepository struct {
	mock2.Mock
}

func (m *MockLeagueRepository) Create(ctx context.Context, league *models.League) error {
	args := m.Called(ctx, league)
	return args.Error(0)
}

func (m *MockLeagueRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.League, error) {
	args := m.Called(ctx, id)
	league := args.Get(0)
	if league == nil {
		return nil, args.Error(1)
	}
	return league.(*models.League), args.Error(1)
}

func (m *MockLeagueRepository) FindAll(ctx context.Context) ([]*models.League, error) {
	args := m.Called(ctx)
	leagues := args.Get(0)
	if leagues == nil {
		return nil, args.Error(1)
	}
	return leagues.([]*models.League), args.Error(1)
}

func (m *MockLeagueRepository) FindByStatus(ctx context.Context, status models.LeagueStatus) ([]*models.League, error) {
	args := m.Called(ctx, status)
	leagues := args.Get(0)
	if leagues == nil {
		return nil, args.Error(1)
	}
	return leagues.([]*models.League), args.Error(1)
}

func (m *MockLeagueRepository) Update(ctx context.Context, league *models.League) error {
	args := m.Called(ctx, league)
	return args.Error(0)
}


