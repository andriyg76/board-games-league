package main

import (
	"context"
	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/repositories/mocks"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/userapi"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type noopSessionService struct{}

func (n *noopSessionService) CreateSession(ctx context.Context, userID primitive.ObjectID, userCode string, externalIDs []string, name, avatar string, ipAddress, userAgent string) (rotateToken, actionToken string, err error) {
	return "", "", nil
}
func (n *noopSessionService) RefreshActionToken(ctx context.Context, rotateToken, ipAddress, userAgent string) (newRotateToken, actionToken string, err error) {
	return "", "", nil
}
func (n *noopSessionService) InvalidateSession(ctx context.Context, rotateToken string) error { return nil }
func (n *noopSessionService) CleanupExpiredSessions(ctx context.Context) error                 { return nil }

func setupTestRouter(mockUserRepo repositories.UserRepository, provider auth.ExternalAuthProvider) *chi.Mux {
	r := chi.NewRouter()
	authHandler := auth.NewHandler(mockUserRepo, services.SessionService(&noopSessionService{}), provider)
	userProfileHandler := userapi.NewHandler(mockUserRepo)

	r.Route("/api", func(r chi.Router) {
		r.Get("/auth/google", authHandler.HandleBeginLoginFlow)
		r.Post("/auth/google/callback", authHandler.GoogleCallbackHandler)
		r.Post("/auth/logout", authHandler.LogoutHandler)

		r.Group(func(r chi.Router) {
			r.Use(authHandler.Middleware)
			r.Get("/user", userProfileHandler.GetUserHandler)
			// Add other routes as needed
		})
	})

	return r
}

func TestProtectedEndpoint(t *testing.T) {
	router := setupTestRouter(new(mocks.MockUserRepository), nil)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/api/user", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d for unauthorized access, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestProtectedEndpointWithAuth(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)

	// Set up mock expectations
	mockUser := &models.User{
		ID:          primitive.NewObjectID(),
		ExternalIDs: []string{"test"},
		Name:        "Test User",
		Avatar:      "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Alias:       "test-user",
	}

	// Expect FindByExternalId to be called with any context and the test external ID
	mockUserRepo.On("FindByExternalId", mock.Anything, []string{"test"}).Return(mockUser, nil)

	router := setupTestRouter(mockUserRepo, nil)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Create a test token
	token, _ := user_profile.CreateAuthToken([]string{"test"}, "testid", "Test User", "")

	req, _ := http.NewRequest("GET", ts.URL+"/api/user", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d for authorized access, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestNotFoundHandler(t *testing.T) {
	router := setupTestRouter(new(mocks.MockUserRepository), nil)
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/nonexistent")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d for non-existent route, got %d", http.StatusNotFound, resp.StatusCode)
	}
}
