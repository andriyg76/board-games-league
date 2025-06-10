package mocks

import (
	"context"
	"github.com/andriyg76/bgl/models"
	mock2 "github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock2.Mock
}

func (m *MockUserRepository) ListAll(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) AliasUnique(ctx context.Context, alias string) (bool, error) {
	args := m.Called(ctx, alias)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) FindByExternalId(ctx context.Context, externalIDs []string) (*models.User, error) {
	args := m.Called(ctx, externalIDs)
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	} else {
		return user.(*models.User), args.Error(1)
	}
}

func (m *MockUserRepository) FindByID(ctx context.Context, ID primitive.ObjectID) (*models.User, error) {
	args := m.Called(ctx, ID)
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	} else {
		return user.(*models.User), args.Error(1)
	}
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
