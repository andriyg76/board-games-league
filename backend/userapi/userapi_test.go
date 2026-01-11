package userapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories/mocks"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateUser(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	handler := NewHandler(mockRepo)

	t.Run("Claims are null or bad", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/update-user", nil)
		rr := httptest.NewRecorder()

		handler.UpdateUser(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("User profile not found", func(t *testing.T) {
		claims := &user_profile.UserProfile{ExternalIDs: []string{"test@example.com"}, Code: "00"}
		ctx := context.WithValue(context.Background(), "user", claims)
		req := httptest.NewRequest("PUT", "/update-user", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		mockRepo.On("FindByExternalId", mock.Anything, []string{"test@example.com"}).Return(nil, nil)

		handler.UpdateUser(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Update user successfully", func(t *testing.T) {
		claims := &user_profile.UserProfile{ExternalIDs: []string{"test2@example.com"}, Code: "000000000000000000000000"}
		ctx := context.WithValue(context.Background(), "user", claims)
		user := &models.User{ExternalIDs: []string{"test2@example.com"}}
		reqBody, _ := json.Marshal(user)
		req := httptest.NewRequest("PUT", "/update-user", bytes.NewBuffer(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		mockRepo.On("FindByExternalId", mock.Anything, []string{"test2@example.com"}).Return(user, nil)
		mockRepo.On("Update", mock.Anything, user).Return(nil)

		handler.UpdateUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
