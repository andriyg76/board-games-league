package gameapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories/mocks"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testSuperAdminID = "test-superadmin-id"

func setupTestRouter(repo *mocks.MockGameTypeRepository) *chi.Mux {
	r := chi.NewRouter()
	handler := &Handler{
		gameTypeRepository: repo,
		leagueService:      nil, // Not needed for game type tests
	}
	handler.RegisterRoutes(r, nil)
	return r
}

// Helper to create request with superadmin context
func addSuperAdminContext(req *http.Request) *http.Request {
	profile := &user_profile.UserProfile{
		Code:        "test-code",
		ExternalIDs: []string{testSuperAdminID},
		Name:        "Test Admin",
	}
	ctx := context.WithValue(req.Context(), "user", profile)
	return req.WithContext(ctx)
}

func TestListGameTypes(t *testing.T) {
	mockRepo := new(mocks.MockGameTypeRepository)
	router := setupTestRouter(mockRepo)

	t.Run("Successfully list game types", func(t *testing.T) {
		gameTypes := []*models.GameType{
			{
				ID:          primitive.NewObjectID(),
				Key:         "test_game",
				Names:       map[string]string{"en": "Test Game Type"},
				ScoringType: models.ScoringTypeClassic,
			},
		}

		mockRepo.On("FindAll", mock.Anything).Return(gameTypes, nil).Once()

		req := httptest.NewRequest("GET", "/game_types", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response []gameTypeAPI
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, gameTypes[0].Key, response[0].Key)
	})
}

func TestCreateGameType(t *testing.T) {
	// Setup superadmin for tests
	restore := auth.SetSuperAdminsForTesting([]string{testSuperAdminID})
	defer restore()

	mockRepo := new(mocks.MockGameTypeRepository)
	router := setupTestRouter(mockRepo)

	t.Run("Successfully create game type", func(t *testing.T) {
		gt := gameTypeAPI{
			Key:         "new_game",
			Names:       map[string]string{"en": "New Game Type"},
			ScoringType: "classic",
			MinPlayers:  2,
			MaxPlayers:  10,
		}

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.GameType")).Return(nil).Once()

		reqBody, _ := json.Marshal(gt)
		req := httptest.NewRequest("POST", "/game_types", bytes.NewBuffer(reqBody))
		req = addSuperAdminContext(req)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Forbidden for non-superadmin", func(t *testing.T) {
		gt := gameTypeAPI{
			Key:   "new_game",
			Names: map[string]string{"en": "New Game Type"},
		}

		reqBody, _ := json.Marshal(gt)
		req := httptest.NewRequest("POST", "/game_types", bytes.NewBuffer(reqBody))
		// Add non-superadmin context
		profile := &user_profile.UserProfile{
			Code:        "test-code",
			ExternalIDs: []string{"non-admin-id"},
		}
		ctx := context.WithValue(req.Context(), "user", profile)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("Invalid request payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/game_types", bytes.NewBuffer([]byte("invalid json")))
		req = addSuperAdminContext(req)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestGetGameType(t *testing.T) {
	mockRepo := new(mocks.MockGameTypeRepository)
	router := setupTestRouter(mockRepo)

	t.Run("Successfully get game type", func(t *testing.T) {
		id := primitive.NewObjectID()
		gameType := &models.GameType{
			ID:          id,
			Key:         "test_game",
			Names:       map[string]string{"en": "Test Game Type"},
			ScoringType: models.ScoringTypeClassic,
		}

		mockRepo.On("FindByID", mock.Anything, id).Return(gameType, nil).Once()

		req := httptest.NewRequest("GET", "/game_types/"+utils.IdToCode(id), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response gameTypeAPI
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, gameType.Key, response.Key)
	})

	t.Run("Game type not found", func(t *testing.T) {
		id := primitive.NewObjectID()
		mockRepo.On("FindByID", mock.Anything, id).Return(nil, nil).Once()

		req := httptest.NewRequest("GET", "/game_types/"+utils.IdToCode(id), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestUpdateGameType(t *testing.T) {
	// Setup superadmin for tests
	restore := auth.SetSuperAdminsForTesting([]string{testSuperAdminID})
	defer restore()

	mockRepo := new(mocks.MockGameTypeRepository)
	router := setupTestRouter(mockRepo)

	t.Run("Successfully update game type", func(t *testing.T) {
		id := primitive.NewObjectID()
		existingGameType := &models.GameType{
			ID:          id,
			Key:         "existing_game",
			Names:       map[string]string{"en": "Original Name"},
			Version:     1,
			ScoringType: models.ScoringTypeClassic,
		}

		updatedGameType := gameTypeAPI{
			Key:         "existing_game",
			Names:       map[string]string{"en": "Updated Game Type"},
			ScoringType: "classic",
		}

		// Mock FindByID call
		mockRepo.On("FindByID", mock.Anything, id).Return(existingGameType, nil).Once()
		// Mock Update call
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.GameType")).Return(nil).Once()

		reqBody, _ := json.Marshal(updatedGameType)
		req := httptest.NewRequest("PUT", "/game_types/"+utils.IdToCode(id), bytes.NewBuffer(reqBody))
		req = addSuperAdminContext(req)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Forbidden for non-superadmin", func(t *testing.T) {
		id := primitive.NewObjectID()

		req := httptest.NewRequest("PUT", "/game_types/"+utils.IdToCode(id), bytes.NewBuffer([]byte("{}")))
		profile := &user_profile.UserProfile{
			Code:        "test-code",
			ExternalIDs: []string{"non-admin-id"},
		}
		ctx := context.WithValue(req.Context(), "user", profile)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("Invalid game type ID", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/game_types/invalid-id", nil)
		req = addSuperAdminContext(req)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestDeleteGameType(t *testing.T) {
	// Setup superadmin for tests
	restore := auth.SetSuperAdminsForTesting([]string{testSuperAdminID})
	defer restore()

	mockRepo := new(mocks.MockGameTypeRepository)
	router := setupTestRouter(mockRepo)

	t.Run("Successfully delete game type", func(t *testing.T) {
		id := primitive.NewObjectID()
		code := utils.IdToCode(id)

		// Non-builtin game type
		gameType := &models.GameType{
			ID:      id,
			Key:     "custom_game",
			BuiltIn: false,
		}

		mockRepo.On("FindByID", mock.Anything, id).Return(gameType, nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequest("DELETE", "/game_types/"+code, nil)
		req = addSuperAdminContext(req)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("Cannot delete built-in game type", func(t *testing.T) {
		id := primitive.NewObjectID()
		code := utils.IdToCode(id)

		// Built-in game type
		gameType := &models.GameType{
			ID:      id,
			Key:     "mafia",
			BuiltIn: true,
		}

		mockRepo.On("FindByID", mock.Anything, id).Return(gameType, nil).Once()

		req := httptest.NewRequest("DELETE", "/game_types/"+code, nil)
		req = addSuperAdminContext(req)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("Forbidden for non-superadmin", func(t *testing.T) {
		id := primitive.NewObjectID()
		code := utils.IdToCode(id)

		req := httptest.NewRequest("DELETE", "/game_types/"+code, nil)
		profile := &user_profile.UserProfile{
			Code:        "test-code",
			ExternalIDs: []string{"non-admin-id"},
		}
		ctx := context.WithValue(req.Context(), "user", profile)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("Invalid game type ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/game_types/invalid-id", nil)
		req = addSuperAdminContext(req)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
