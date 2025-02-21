package userapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
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

func TestUpdateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := UpdateUser(mockRepo)

	t.Run("Claims are null or bad", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/update-user", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("User profile not found", func(t *testing.T) {
		claims := &user_profile.UserProfile{IDs: []string{"test@example.com"}, ID: "00"}
		ctx := context.WithValue(context.Background(), "user", claims)
		req := httptest.NewRequest("PUT", "/update-user", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		mockRepo.On("FindByExternalId", mock.Anything, []string{"test@example.com"}).Return(nil, nil)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Update user successfully", func(t *testing.T) {
		claims := &user_profile.UserProfile{IDs: []string{"test2@example.com"}, ID: "000000000000000000000000"}
		ctx := context.WithValue(context.Background(), "user", claims)
		user := &models.User{ExternalIDs: []string{"test2@example.com"}}
		reqBody, _ := json.Marshal(user)
		req := httptest.NewRequest("PUT", "/update-user", bytes.NewBuffer(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		mockRepo.On("FindByExternalId", mock.Anything, []string{"test2@example.com"}).Return(user, nil)
		mockRepo.On("Update", mock.Anything, user).Return(nil)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
