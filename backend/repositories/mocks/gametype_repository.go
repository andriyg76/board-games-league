package mocks

import (
	"context"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	mock2 "github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockGameTypeRepository struct {
	repositories.GameTypeRepository
	mock2.Mock
}

func (m *MockGameTypeRepository) Create(ctx context.Context, gameType *models.GameType) error {
	args := m.Called(ctx, gameType)
	return args.Error(0)
}

func (m *MockGameTypeRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.GameType, error) {
	args := m.Called(ctx, id)
	if gameType := args.Get(0); gameType != nil {
		return gameType.(*models.GameType), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGameTypeRepository) FindByKey(ctx context.Context, key string) (*models.GameType, error) {
	args := m.Called(ctx, key)
	if gameType := args.Get(0); gameType != nil {
		return gameType.(*models.GameType), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGameTypeRepository) FindAll(ctx context.Context) ([]*models.GameType, error) {
	args := m.Called(ctx)
	if gameTypes := args.Get(0); gameTypes != nil {
		return gameTypes.([]*models.GameType), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGameTypeRepository) Update(ctx context.Context, gameType *models.GameType) error {
	args := m.Called(ctx, gameType)
	return args.Error(0)
}

func (m *MockGameTypeRepository) Upsert(ctx context.Context, gameType *models.GameType) error {
	args := m.Called(ctx, gameType)
	return args.Error(0)
}

func (m *MockGameTypeRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
