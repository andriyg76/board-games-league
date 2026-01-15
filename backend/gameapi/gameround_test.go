package gameapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"time"
)

// MockIdAndCodeCache is a mock implementation of IdAndCodeCache for testing
type MockIdAndCodeCache struct {
	mock.Mock
}

func (m *MockIdAndCodeCache) GetByID(id primitive.ObjectID) *models.IdAndCode {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.IdAndCode)
}

func (m *MockIdAndCodeCache) GetByCode(code string) (*models.IdAndCode, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IdAndCode), args.Error(1)
}

// leagueIDMiddleware creates a middleware that adds league ID to context
func leagueIDMiddleware(leagueID primitive.ObjectID) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "leagueID", leagueID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TestStartGame(t *testing.T) {
	mockGameRoundRepo := new(mocks.MockGameRoundRepository)
	mockGameTypeRepo := new(mocks.MockGameTypeRepository)
	mockIdCodeCache := new(MockIdAndCodeCache)
	
	leagueID := primitive.NewObjectID()
	
	handler := &Handler{
		gameRoundRepository: mockGameRoundRepo,
		gameTypeRepository:  mockGameTypeRepo,
		idCodeCache:         mockIdCodeCache,
		leagueService:       nil,
	}

	router := chi.NewRouter()
	router.Use(leagueIDMiddleware(leagueID))
	router.Post("/games", handler.startGame)

	t.Run("Start game with valid team assignments", func(t *testing.T) {
		gameTypeID := primitive.NewObjectID()
		membership1ID := primitive.NewObjectID()
		membership2ID := primitive.NewObjectID()

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
				{MembershipCode: "player1-code", Position: 1, TeamName: "team_a"},
				{MembershipCode: "player2-code", Position: 2, TeamName: "team_b"},
			},
		}

		// Mock idCodeCache.GetByCode for membership codes
		mockIdCodeCache.On("GetByCode", "player1-code").Return(&models.IdAndCode{ID: membership1ID, Code: "player1-code"}, nil).Once()
		mockIdCodeCache.On("GetByCode", "player2-code").Return(&models.IdAndCode{ID: membership2ID, Code: "player2-code"}, nil).Once()
		
		mockGameTypeRepo.On("FindByKey", mock.Anything, "test_game").Return(gameType, nil).Once()
		mockGameRoundRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.GameRound")).Return(nil).Once()

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Start game with missing team assignments", func(t *testing.T) {
		gameTypeID := primitive.NewObjectID()
		membership1ID := primitive.NewObjectID()

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
				{MembershipCode: "player1-code-2", Position: 1, TeamName: "team_a"},
			},
		}

		// Mock idCodeCache.GetByCode for membership codes
		mockIdCodeCache.On("GetByCode", "player1-code-2").Return(&models.IdAndCode{ID: membership1ID, Code: "player1-code-2"}, nil).Once()
		
		mockGameTypeRepo.On("FindByKey", mock.Anything, "test_game2").Return(gameType, nil).Once()

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
	mockIdCodeCache := new(MockIdAndCodeCache)
	
	handler := &Handler{
		gameRoundRepository: mockRepo,
		idCodeCache:         mockIdCodeCache,
		leagueService:       nil,
	}

	router := chi.NewRouter()
	// Note: The handler expects URL params {gameRoundCode} and {playerCode}
	router.Put("/games/{gameRoundCode}/players/{playerCode}/score", handler.updatePlayerScore)

	t.Run("Successfully update player score", func(t *testing.T) {
		gameID := primitive.NewObjectID()
		membershipID := primitive.NewObjectID()
		
		gameCode := utils.IdToCode(gameID)
		playerCode := utils.IdToCode(membershipID)

		gameRound := &models.GameRound{
			ID: gameID,
			Players: []models.GameRoundPlayer{
				{MembershipID: membershipID, Score: 0},
			},
		}

		// Mock idCodeCache calls
		mockIdCodeCache.On("GetByCode", gameCode).Return(&models.IdAndCode{ID: gameID, Code: gameCode}, nil).Once()
		mockIdCodeCache.On("GetByCode", playerCode).Return(&models.IdAndCode{ID: membershipID, Code: playerCode}, nil).Once()
		
		mockRepo.On("FindByID", mock.Anything, gameID).Return(gameRound, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.GameRound")).Return(nil).Once()

		req := updateScoreRequest{Score: 100}
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/games/"+gameCode+"/players/"+playerCode+"/score", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Player not found", func(t *testing.T) {
		gameID := primitive.NewObjectID()
		membershipID := primitive.NewObjectID()
		
		gameCode := utils.IdToCode(gameID)
		playerCode := utils.IdToCode(membershipID)

		gameRound := &models.GameRound{
			ID:      gameID,
			Players: []models.GameRoundPlayer{},
		}

		// Mock idCodeCache calls
		mockIdCodeCache.On("GetByCode", gameCode).Return(&models.IdAndCode{ID: gameID, Code: gameCode}, nil).Once()
		mockIdCodeCache.On("GetByCode", playerCode).Return(&models.IdAndCode{ID: membershipID, Code: playerCode}, nil).Once()
		
		mockRepo.On("FindByID", mock.Anything, gameID).Return(gameRound, nil).Once()

		req := updateScoreRequest{Score: 100}
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/games/"+gameCode+"/players/"+playerCode+"/score", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestFinalizeGame(t *testing.T) {
	mockRepo := new(mocks.MockGameRoundRepository)
	mockIdCodeCache := new(MockIdAndCodeCache)
	
	handler := &Handler{
		gameRoundRepository: mockRepo,
		idCodeCache:         mockIdCodeCache,
		leagueService:       nil,
	}

	router := chi.NewRouter()
	// Note: The handler uses GetIDFromChiURL(r, "code")
	router.Put("/games/{code}/finalize", handler.finalizeGame)

	t.Run("Successfully finalize game", func(t *testing.T) {
		gameID := primitive.NewObjectID()
		membership1ID := primitive.NewObjectID()
		membership2ID := primitive.NewObjectID()
		
		gameCode := utils.IdToCode(gameID)
		player1Code := utils.IdToCode(membership1ID)
		player2Code := utils.IdToCode(membership2ID)

		gameRound := &models.GameRound{
			ID: gameID,
			Players: []models.GameRoundPlayer{
				{PlayerID: membership1ID, MembershipID: membership1ID, TeamName: "team_a"},
				{PlayerID: membership2ID, MembershipID: membership2ID, TeamName: "team_b"},
			},
			TeamScores: []models.TeamScore{
				{Name: "team_a"},
				{Name: "team_b"},
			},
		}

		// Mock idCodeCache - for the URL code parsing 
		mockIdCodeCache.On("GetByCode", gameCode).Return(&models.IdAndCode{ID: gameID, Code: gameCode}, nil).Once()
		
		mockRepo.On("FindByID", mock.Anything, gameID).Return(gameRound, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.GameRound")).Return(nil).Once()

		req := finalizeGameRequest{
			PlayerScores: map[string]int64{
				player1Code: 100,
				player2Code: 50,
			},
			TeamScores: map[string]int64{
				"team_a": 100,
				"team_b": 50,
			},
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/games/"+gameCode+"/finalize", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, httpReq)

		// Print error for debugging if status is not OK
		if rr.Code != http.StatusOK {
			fmt.Printf("Response body: %s\n", rr.Body.String())
		}
		
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
