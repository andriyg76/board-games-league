package gameapi

import (
	"bytes"
	"encoding/json"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStartGame(t *testing.T) {
	mockGameRoundRepo := new(mocks.MockGameRoundRepository)
	mockGameTypeRepo := new(mocks.MockGameTypeRepository)
	mockUserService := new(mocks.MockUserService)
	handler := &Handler{
		gameRoundRepository: mockGameRoundRepo,
		gameTypeRepository:  mockGameTypeRepo,
		userService:         mockUserService,
		leagueService:       nil, // Not needed for game round tests
	}

	router := chi.NewRouter()
	router.Post("/games", handler.startGame)

	t.Run("Start game with valid team assignments", func(t *testing.T) {
		gameTypeID := primitive.NewObjectID()
		player1ID := primitive.NewObjectID()
		player2ID := primitive.NewObjectID()

		gameType := &models.GameType{
			ID:  gameTypeID,
			Key: "test_game",
			Names: map[string]string{
				"en": "Test Game",
			},
			Roles: []models.Role{
				{Key: "team_a", Names: map[string]string{"en": "Team A"}, RoleType: models.RoleTypeMultiple},
				{Key: "team_b", Names: map[string]string{"en": "Team B"}, RoleType: models.RoleTypeMultiple},
			},
		}

		req := startGameRequest{
			Name:      "Test Game Round",
			Type:      "test_game",
			StartTime: time.Now(),
			Players: []playerSetup{
				{UserID: player1ID, Position: 1, TeamName: "team_a"},
				{UserID: player2ID, Position: 2, TeamName: "team_b"},
			},
		}

		mockGameTypeRepo.On("FindByKey", mock.Anything, "test_game").Return(gameType, nil).Once()
		mockUserService.On("FindByID", mock.Anything, player1ID).Return(&models.User{ID: player1ID}, nil).Once()
		mockUserService.On("FindByID", mock.Anything, player2ID).Return(&models.User{ID: player2ID}, nil).Once()
		mockGameRoundRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.GameRound")).Return(nil).Once()

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Start game with missing team assignments", func(t *testing.T) {
		gameTypeID := primitive.NewObjectID()
		player1ID := primitive.NewObjectID()

		gameType := &models.GameType{
			ID:  gameTypeID,
			Key: "test_game2",
			Names: map[string]string{
				"en": "Test Game",
			},
			Roles: []models.Role{
				{Key: "team_a", Names: map[string]string{"en": "Team A"}, RoleType: models.RoleTypeMultiple},
				{Key: "team_b", Names: map[string]string{"en": "Team B"}, RoleType: models.RoleTypeMultiple},
			},
		}

		req := startGameRequest{
			Name:      "Test Game Round",
			Type:      "test_game2",
			StartTime: time.Now(),
			Players: []playerSetup{
				{UserID: player1ID, Position: 1, TeamName: "team_a"},
			},
		}

		mockGameTypeRepo.On("FindByKey", mock.Anything, "test_game2").Return(gameType, nil).Once()
		mockUserService.On("FindByID", mock.Anything, player1ID).Return(&models.User{ID: player1ID}, nil).Once()

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid team assignments")
	})
}

func TestUpdatePlayerScore(t *testing.T) {
	mockRepo := new(mocks.MockGameRoundRepository)
	handler := &Handler{
		gameRoundRepository: mockRepo,
		leagueService:       nil, // Not needed for this test
	}

	router := chi.NewRouter()
	router.Put("/games/{id}/players/{userId}", handler.updatePlayerScore)

	t.Run("Successfully update player score", func(t *testing.T) {
		gameID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		gameRound := &models.GameRound{
			ID: gameID,
			Players: []models.GameRoundPlayer{
				{PlayerID: userID, Score: 0},
			},
		}

		mockRepo.On("FindByID", mock.Anything, gameID).Return(gameRound, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.GameRound")).Return(nil).Once()

		req := updateScoreRequest{Score: 100}
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/games/"+gameID.Hex()+"/players/"+userID.Hex(), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Player not found", func(t *testing.T) {
		gameID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		gameRound := &models.GameRound{
			ID:      gameID,
			Players: []models.GameRoundPlayer{},
		}

		mockRepo.On("FindByID", mock.Anything, gameID).Return(gameRound, nil).Once()

		req := updateScoreRequest{Score: 100}
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/games/"+gameID.Hex()+"/players/"+userID.Hex(), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestFinalizeGame(t *testing.T) {
	mockRepo := new(mocks.MockGameRoundRepository)
	handler := &Handler{
		gameRoundRepository: mockRepo,
		leagueService:       nil, // Not needed for this test
	}

	router := chi.NewRouter()
	router.Put("/games/{id}/finalize", handler.finalizeGame)

	t.Run("Successfully finalize game", func(t *testing.T) {
		gameID := primitive.NewObjectID()
		player1ID := primitive.NewObjectID()
		player2ID := primitive.NewObjectID()

		gameRound := &models.GameRound{
			ID: gameID,
			Players: []models.GameRoundPlayer{
				{PlayerID: player1ID, TeamName: "team_a"},
				{PlayerID: player2ID, TeamName: "team_b"},
			},
			TeamScores: []models.TeamScore{
				{Name: "team_a"},
				{Name: "team_b"},
			},
		}

		mockRepo.On("FindByID", mock.Anything, gameID).Return(gameRound, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.GameRound")).Return(nil).Once()

		req := finalizeGameRequest{
			PlayerScores: map[string]int64{
				player1ID.Hex(): 100,
				player2ID.Hex(): 50,
			},
			TeamScores: map[string]int64{
				"team_a": 100,
				"team_b": 50,
			},
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/games/"+gameID.Hex()+"/finalize", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
