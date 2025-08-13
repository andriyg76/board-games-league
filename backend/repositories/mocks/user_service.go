package mocks

import (
	"context"
	"github.com/andriyg76/bgl/models"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) FindAll(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) FindByID(ctx context.Context, ID primitive.ObjectID) (*models.User, error) {
	args := m.Called(ctx, ID)
	if user := args.Get(0); user != nil {
		return user.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserService) FindByCode(ctx context.Context, code string) (*models.User, error) {
	args := m.Called(ctx, code)
	if user := args.Get(0); user != nil {
		return user.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
