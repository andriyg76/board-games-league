package gameapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/andriyg76/bgl/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) FindByID(ctx context.Context, ID primitive.ObjectID) (*models.User, error) {
	args := m.Called(ctx, ID)
	if user := args.Get(0); user != nil {
		return user.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestStartGame(t *testing.T) {
	mockGameRoundRepo := new(MockGameRoundRepository)
	mockGameTypeRepo := new(MockGameTypeRepository)
	mockUserService := new(MockUserService)
	handler := &Handler{
		gameRoundRepository: mockGameRoundRepo,
		gameTypeRepository:  mockGameTypeRepo,
		userService:         mockUserService,
	}

	router := chi.NewRouter()
	router.Post("/games", handler.startGame)

	t.Run("Start game with valid team assignments", func(t *testing.T) {
		gameTypeID := primitive.NewObjectID()
		player1ID := primitive.NewObjectID()
		player2ID := primitive.NewObjectID()

		gameType := &models.GameType{
			ID:   gameTypeID,
			Name: "Test Game",
			Teams: []models.Label{
				{Name: "Team A"},
				{Name: "Team B"},
			},
		}

		req := startGameRequest{
			Name:      "Test Game Round",
			Type:      "Test Game",
			StartTime: time.Now(),
			Players: []playerSetup{
				{UserID: player1ID, Order: 1, TeamName: "Team A"},
				{UserID: player2ID, Order: 2, TeamName: "Team B"},
			},
		}

		mockGameTypeRepo.On("FindByName", mock.Anything, "Test Game").Return(gameType, nil)
		mockUserService.On("FindByID", mock.Anything, player1ID).Return(&models.User{ID: player1ID}, nil)
		mockUserService.On("FindByID", mock.Anything, player2ID).Return(&models.User{ID: player2ID}, nil)
		mockGameRoundRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.GameRound")).Return(nil)

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
			ID:   gameTypeID,
			Name: "Test Game",
			Teams: []models.Label{
				{Name: "Team A"},
				{Name: "Team B"},
			},
		}

		req := startGameRequest{
			Name:      "Test Game Round",
			Type:      "Test Game",
			StartTime: time.Now(),
			Players: []playerSetup{
				{UserID: player1ID, Order: 1, TeamName: "Team A"},
			},
		}

		mockGameTypeRepo.On("FindByName", mock.Anything, "Test Game").Return(gameType, nil)
		mockUserService.On("FindByID", mock.Anything, player1ID).Return(&models.User{ID: player1ID}, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid team assignments")
	})
}

func TestUpdatePlayerScore(t *testing.T) {
	mockRepo := new(MockGameRoundRepository)
	handler := &Handler{gameRoundRepository: mockRepo}

	router := chi.NewRouter()
	router.Put("/games/{id}/players/{userId}", handler.updatePlayerScore)

	t.Run("Successfully update player score", func(t *testing.T) {
		gameID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		gameRound := &models.GameRound{
			ID: gameID,
			Players: []models.GameRoundPlayer{
				{UserID: userID, Score: 0},
			},
		}

		mockRepo.On("FindByID", mock.Anything, gameID).Return(gameRound, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.GameRound")).Return(nil)

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

		mockRepo.On("FindByID", mock.Anything, gameID).Return(gameRound, nil)

		req := updateScoreRequest{Score: 100}
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/games/"+gameID.Hex()+"/players/"+userID.Hex(), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestFinalizeGame(t *testing.T) {
	mockRepo := new(MockGameRoundRepository)
	handler := &Handler{gameRoundRepository: mockRepo}

	router := chi.NewRouter()
	router.Put("/games/{id}/finalize", handler.finalizeGame)

	t.Run("Successfully finalize game", func(t *testing.T) {
		gameID := primitive.NewObjectID()
		player1ID := primitive.NewObjectID()
		player2ID := primitive.NewObjectID()

		gameRound := &models.GameRound{
			ID: gameID,
			Players: []models.GameRoundPlayer{
				{UserID: player1ID, TeamName: "Team A"},
				{UserID: player2ID, TeamName: "Team B"},
			},
			TeamScores: []models.TeamScore{
				{Name: "Team A"},
				{Name: "Team B"},
			},
		}

		mockRepo.On("FindByID", mock.Anything, gameID).Return(gameRound, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.GameRound")).Return(nil)

		req := finalizeGameRequest{
			PlayerScores: map[string]int64{
				player1ID.Hex(): 100,
				player2ID.Hex(): 50,
			},
			TeamScores: map[string]int64{
				"Team A": 100,
				"Team B": 50,
			},
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/games/"+gameID.Hex()+"/finalize", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
