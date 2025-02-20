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

func (m *MockUserRepository) FindByExternalId(ctx context.Context, externalIDs string) (*models.User, error) {
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

func TestCheckAliasUniquenessHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := CheckAliasUniquenessHandler(mockRepo)

	t.Run("Alias is required", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/check-alias", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "Alias is required\n", rr.Body.String())
	})

	t.Run("Alias is unique", func(t *testing.T) {
		mockRepo.On("AliasUnique", mock.Anything, "unique-alias").Return(true, nil)

		req := httptest.NewRequest("GET", "/check-alias?alias=unique-alias", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, `{"isUnique": true}`, rr.Body.String())
	})
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
		claims := &user_profile.UserProfile{Email: "test@example.com"}
		ctx := context.WithValue(context.Background(), "user", claims)
		req := httptest.NewRequest("PUT", "/update-user", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		mockRepo.On("FindByExternalId", mock.Anything, "test@example.com").Return(nil, nil)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Update user successfully", func(t *testing.T) {
		claims := &user_profile.UserProfile{Email: "test2@example.com"}
		ctx := context.WithValue(context.Background(), "user", claims)
		user := &models.User{Email: "test2@example.com"}
		reqBody, _ := json.Marshal(user)
		req := httptest.NewRequest("PUT", "/update-user", bytes.NewBuffer(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		mockRepo.On("FindByExternalId", mock.Anything, "test2@example.com").Return(user, nil)
		mockRepo.On("Update", mock.Anything, user).Return(nil)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestGetUserHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := GetUserHandler(mockRepo)

	t.Run("Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/get-user", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Get user successfully", func(t *testing.T) {
		claims := &user_profile.UserProfile{Email: "test@example.com"}
		ctx := context.WithValue(context.Background(), "user", claims)
		user := &models.User{Email: "test@example.com", Name: "Test User", Avatar: "http://example.com/avatar.jpg", Alias: "test-alias"}
		req := httptest.NewRequest("GET", "/get-user", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		mockRepo.On("FindByExternalId", mock.Anything, "test@example.com").Return(user, nil)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, `{"email":"test@example.com","name":"Test User","picture":"http://example.com/avatar.jpg","alias":"test-alias"}`, rr.Body.String())
	})
}

func TestAdminCreateUserHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := AdminCreateUserHandler(mockRepo)

	t.Run("Invalid request payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin-create-user", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Email is required", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{"email": ""})
		req := httptest.NewRequest("POST", "/admin-create-user", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Create user successfully", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{"email": "test@example.com"})
		req := httptest.NewRequest("POST", "/admin-create-user", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		mockRepo.On("FindByExternalId", mock.Anything, "test@example.com").Return(nil, nil)
		mockRepo.On("AliasUnique", mock.Anything, mock.Anything).Return(true, nil)
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Equal(t, "User created successfully", rr.Body.String())
	})
}
