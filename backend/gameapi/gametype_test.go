package gameapi

import (
	"bytes"
	"encoding/json"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories/mocks"
	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestRouter(repo *mocks.MockGameTypeRepository) *chi.Mux {
	r := chi.NewRouter()
	handler := &Handler{
		gameTypeRepository: repo,
		leagueService:      nil, // Not needed for game type tests
	}
	handler.RegisterRoutes(r)
	return r
}

func TestListGameTypes(t *testing.T) {
	mockRepo := new(mocks.MockGameTypeRepository)
	router := setupTestRouter(mockRepo)

	t.Run("Successfully list game types", func(t *testing.T) {
		gameTypes := []*models.GameType{
			{
				ID:          primitive.NewObjectID(),
				Name:        "Test Game Type",
				ScoringType: string(models.ScoringTypeClassic),
			},
		}

		mockRepo.On("FindAll", mock.Anything).Return(gameTypes, nil)

		req := httptest.NewRequest("GET", "/game_types", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response []*models.GameType
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, gameTypes[0].Name, response[0].Name)
	})
}

func TestCreateGameType(t *testing.T) {
	mockRepo := new(mocks.MockGameTypeRepository)
	router := setupTestRouter(mockRepo)

	t.Run("Successfully create game type", func(t *testing.T) {
		gameType := &models.GameType{
			Name:       "New Game Type",
			MinPlayers: 2,
			MaxPlayers: 10,
		}

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.GameType")).Return(nil)

		reqBody, _ := json.Marshal(gameType)
		req := httptest.NewRequest("POST", "/game_types", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Invalid request payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/game_types", bytes.NewBuffer([]byte("invalid json")))
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
			ID:   id,
			Name: "Test Game Type",
		}

		mockRepo.On("FindByID", mock.Anything, id).Return(gameType, nil)

		req := httptest.NewRequest("GET", "/game_types/"+utils.IdToCode(id), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.GameType
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, gameType.Name, response.Name)
	})

	t.Run("Game type not found", func(t *testing.T) {
		id := primitive.NewObjectID()
		mockRepo.On("FindByID", mock.Anything, id).Return(nil, nil)

		req := httptest.NewRequest("GET", "/game_types/"+utils.IdToCode(id), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestUpdateGameType(t *testing.T) {
	mockRepo := new(mocks.MockGameTypeRepository)
	router := setupTestRouter(mockRepo)

	t.Run("Successfully update game type", func(t *testing.T) {
		id := primitive.NewObjectID()
		existingGameType := &models.GameType{
			ID:          id,
			Name:        "Original Name",
			Version:     1,
			ScoringType: string(models.ScoringTypeClassic),
		}

		updatedGameType := &gameType{
			Name:        "Updated Game Type",
			ScoringType: string(models.ScoringTypeClassic),
		}

		// Mock FindByID call
		mockRepo.On("FindByID", mock.Anything, id).Return(existingGameType, nil)
		// Mock Update call
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.GameType")).Return(nil)

		reqBody, _ := json.Marshal(updatedGameType)
		req := httptest.NewRequest("PUT", "/game_types/"+utils.IdToCode(id), bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid game type ID", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/game_types/invalid-id", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestDeleteGameType(t *testing.T) {
	mockRepo := new(mocks.MockGameTypeRepository)
	router := setupTestRouter(mockRepo)

	t.Run("Successfully delete game type", func(t *testing.T) {
		id := primitive.NewObjectID()
		code := utils.IdToCode(id)
		mockRepo.On("Delete", mock.Anything, id).Return(nil)

		req := httptest.NewRequest("DELETE", "/game_types/"+code, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("Invalid game type ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/game_types/invalid-id", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
