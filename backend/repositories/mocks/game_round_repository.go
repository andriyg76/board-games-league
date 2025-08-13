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

func (m *MockGameRoundRepository) Update(ctx context.Context, round *models.GameRound) error {
	args := m.Called(ctx, round)
	return args.Error(0)
}

func (m *MockGameRoundRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
