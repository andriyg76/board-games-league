package mocks

import (
	"context"
	"github.com/andriyg76/bgl/models"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockGameRoundRepository struct {
	mock.Mock
}

func (m *MockGameRoundRepository) Create(ctx context.Context, round *models.GameRound) error {
	args := m.Called(ctx, round)
	return args.Error(0)
}

func (m *MockGameRoundRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.GameRound, error) {
	args := m.Called(ctx, id)
	if round := args.Get(0); round != nil {
		return round.(*models.GameRound), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGameRoundRepository) FindAll(ctx context.Context) ([]*models.GameRound, error) {
	args := m.Called(ctx)
	if rounds := args.Get(0); rounds != nil {
		return rounds.([]*models.GameRound), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGameRoundRepository) Update(ctx context.Context, round *models.GameRound) error {
	args := m.Called(ctx, round)
	return args.Error(0)
}

func (m *MockGameRoundRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGameRoundRepository) FindByLeague(ctx context.Context, leagueID primitive.ObjectID) ([]*models.GameRound, error) {
	args := m.Called(ctx, leagueID)
	if rounds := args.Get(0); rounds != nil {
		return rounds.([]*models.GameRound), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGameRoundRepository) HasGamesForMembership(ctx context.Context, membershipID primitive.ObjectID) (bool, error) {
	args := m.Called(ctx, membershipID)
	return args.Bool(0), args.Error(1)
}

func (m *MockGameRoundRepository) FindByLeagueAndStatus(ctx context.Context, leagueID primitive.ObjectID, statuses []models.GameRoundStatus) ([]*models.GameRound, error) {
	args := m.Called(ctx, leagueID, statuses)
	if rounds := args.Get(0); rounds != nil {
		return rounds.([]*models.GameRound), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGameRoundRepository) FindActiveByLeague(ctx context.Context, leagueID primitive.ObjectID) ([]*models.GameRound, error) {
	args := m.Called(ctx, leagueID)
	if rounds := args.Get(0); rounds != nil {
		return rounds.([]*models.GameRound), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGameRoundRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.GameRoundStatus, version int64) error {
	args := m.Called(ctx, id, status, version)
	return args.Error(0)
}
