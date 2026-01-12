package auth

import (
	"context"
	"fmt"
	"github.com/andriyg76/bgl/asserts2"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/gorilla/securecookie"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/andriyg76/bgl/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserRepository implements repositories.UserRepository interface
type MockUserRepository struct {
	mock.Mock
	repositories.UserRepository
}

func (m *MockUserRepository) FindByExternalId(ctx context.Context, externalIDs []string) (*models.User, error) {
	args := m.Called(ctx, externalIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) AliasUnique(ctx context.Context, alias string) (bool, error) {
	args := m.Called(ctx, alias)
	return args.Bool(0), args.Error(1)
}

type testSessionService struct{}

func (s *testSessionService) CreateSession(ctx context.Context, userID primitive.ObjectID, userCode string, externalIDs []string, name, avatar string, ipAddress, userAgent string) (rotateToken, actionToken string, err error) {
	token, err := user_profile.CreateAuthTokenWithExpiry(externalIDs, userCode, name, avatar, 1*time.Hour)
	if err != nil {
		return "", "", err
	}
	return "test-rotate-token", token, nil
}
func (s *testSessionService) RefreshActionToken(ctx context.Context, rotateToken, ipAddress, userAgent string) (newRotateToken, actionToken string, err error) {
	return "", "", nil
}
func (s *testSessionService) InvalidateSession(ctx context.Context, rotateToken string) error {
	return nil
}
func (s *testSessionService) CleanupExpiredSessions(ctx context.Context) error { return nil }

func TestIsSuperAdmin(t *testing.T) {
	// Temporarily set superAdmins for testing
	restore := SetSuperAdminsForTesting([]string{"admin@example.com", "super@example.com"})
	defer restore()

	tests := []struct {
		name     string
		email    []string
		expected bool
	}{
		{"Super admin email should return true", []string{"admin@example.com"}, true},
		{"Regular email should return false", []string{"user@example.com"}, false},
		{"Empty email should return false", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSuperAdmin(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMiddleware(t *testing.T) {
	// Set up test JWT secret
	originalConfig := config
	defer func() { config = originalConfig }()

	mockRepo := new(MockUserRepository)
	middleware := (&Handler{
		userRepository: mockRepo,
		provider:       new(MockExternalAuthProvider),
		requestService: services.NewRequestService(),
	}).Middleware

	tests := []struct {
		name           string
		setupAuth      func(*http.Request)
		expectedStatus int
	}{
		{
			name: "Valid token should pass",
			setupAuth: func(r *http.Request) {
				token, _ := user_profile.CreateAuthToken([]string{"test@example.com"}, "00", "Test User", "http://example.com/avatar.jpg")
				r.AddCookie(&http.Cookie{
					Name:  "auth_token",
					Value: token,
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing token should fail",
			setupAuth:      func(r *http.Request) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler that will be wrapped by middleware
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			tt.setupAuth(req)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Run middleware with test handler
			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			asserts2.Get(t).Equal(tt.expectedStatus, rr.Code)
		})
	}
}

func TestGoogleCallbackHandler(t *testing.T) {

	mockRepo := new(MockUserRepository)
	mockProvider := new(MockExternalAuthProvider)
	handler := Handler{
		userRepository: mockRepo,
		sessionService: services.SessionService(&testSessionService{}),
		requestService: services.NewRequestService(),
		provider:       mockProvider,
	}
	beginFlowHandler := handler.HandleBeginLoginFlow
	finalHandler := handler.GoogleCallbackHandler

	superAdminEmail := "superadmin@example.com"
	restoreSuperAdmins := SetSuperAdminsForTesting([]string{superAdminEmail})
	defer restoreSuperAdmins()

	notExistingEmail := "notexisting@example.com"

	regularEmail := "existing@example.com"
	existingUser := &models.User{
		ID:          primitive.NewObjectID(),
		ExternalIDs: []string{regularEmail},
		Name:        "Existing User",
		Avatar:      "http://example.com/avatar.jpg",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Alias:       "existing",
	}

	mockRepo.On("FindByExternalId", mock.Anything, []string{regularEmail}).Return(existingUser, nil)
	mockRepo.On("FindByExternalId", mock.Anything, []string{superAdminEmail}).Return(nil, nil)
	mockRepo.On("FindByExternalId", mock.Anything, []string{notExistingEmail}).Return(nil, nil)
	mockRepo.On("AliasUnique", mock.Anything, mock.Anything).Return(true, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	superadminRequest := httptest.NewRequest("GET", "/auth/callback?state=somestate&provider=google", nil)
	mockProvider.On("CompleteUserAuthHandler", mock.Anything, superadminRequest).Return(ExternalUser{
		ExternalIDs: []string{superAdminEmail},
		Name:        "Superadmin",
		Avatar:      "http://example.com/avatar.jpg",
	}, nil)

	mockProvider.On("LogoutHandler", mock.Anything, mock.Anything).Return(nil)

	// Test existing user flow
	t.Run("Existing user login", func(t *testing.T) {
		const somestate = "somestate"

		var discord = []string{}
		utils.AddDiscordSendCapturer(func(s string) {
			discord = append(discord, s)
		})
		asserts := asserts2.Get(t)

		var r1 = httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/auth/start?type=google&state="+somestate, nil)
		if err != nil {
			asserts.NoError(err)
			return
		}
		beginFlowHandler(r1, req)

		s, e := store.Get(req, "auth-session")
		if e != nil {
			asserts.NoError(e)
			return
		}
		asserts.Equal(somestate, s.Values["state"])

		location := r1.Header().Get("Location")
		asserts.NotEmpty(location)

		url1, e := url.Parse(location)
		if e != nil {
			asserts.NoError(e)
		}
		url1.Query().Get("state")
		asserts.NotEmpty("state", somestate)

		rr := httptest.NewRecorder()
		regularUserRequest := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?state=%s", somestate), nil)
		mockProvider.On("CompleteUserAuthHandler", mock.Anything, regularUserRequest).Return(ExternalUser{
			ExternalIDs: []string{regularEmail},
			Name:        "Existing User",
			Avatar:      "http://example.com/avatar.jpg",
		}, nil)

		finalHandler(rr, regularUserRequest)

		asserts.
			Equal(http.StatusOK, rr.Code).
			True(len(discord) == 0, "No new notifications to discord").
			True(strings.Contains(rr.Header().Get("Set-Cookie"), "auth_token="),
				"auth_token cookie should be set by end auth")
		for _, cookie := range rr.Result().Cookies() {
			if cookie.Name == "auth_token" {
				profile, err := user_profile.ParseProfile(cookie.Value)
				if err != nil {
					asserts.NoError(err, "Could not parse cookie %v", cookie)
				}
				asserts.Equal([]string{regularEmail}, profile.ExternalIDs, "Invalid user token set")
			}
		}
	})
}

func TestLogoutHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	provider := new(MockExternalAuthProvider)
	handler := Handler{
		userRepository: mockRepo,
		sessionService: services.SessionService(&testSessionService{}),
		requestService: services.NewRequestService(),
		provider:       provider,
	}

	provider.On("LogoutHandler", mock.Anything, mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/logout", nil)
	rr := httptest.NewRecorder()

	handler.LogoutHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Check if auth cookie was cleared
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" {
			assert.Equal(t, "", cookie.Value)
			assert.True(t, cookie.MaxAge < 0)
		}
	}
}

func TestSendNewUserToDiscord_UserNil(t *testing.T) {
	hashKey := []byte("very-secret")
	s := securecookie.New(hashKey, nil)

	// Create a value to store in the cookie
	value := map[string]string{
		"email": "test@example.com",
	}

	// Encode the value
	encoded, err := s.Encode("auth", value)
	if err != nil {
		t.Fatalf("Failed to encode cookie: %v", err)
	}

	req := httptest.NewRequest("POST", "/send", nil)
	// Add the encoded cookie to the request
	req.AddCookie(&http.Cookie{
		Name:  "auth",
		Value: encoded,
	})

	asserts2.Get(t).NotNil(sendNewUserToDiscord(req, nil))
}

func TestSendNewUserToDiscord_UserNotNil(t *testing.T) {
	hashKey := []byte("very-secret")
	s := securecookie.New(hashKey, nil)

	// Create a value to store in the cookie
	value := map[string]string{
		"email": "test@example.com",
	}

	// Encode the value
	encoded, err := s.Encode("auth", value)
	if err != nil {
		t.Fatalf("Failed to encode cookie: %v", err)
	}

	user := &models.User{
		Name:        "Test User",
		ID:          primitive.ObjectID([12]byte{}),
		ExternalIDs: []string{"test@example.com"},
	}

	req := httptest.NewRequest("POST", "/send", nil)
	// Add the encoded cookie to the request
	req.AddCookie(&http.Cookie{
		Name:  "auth",
		Value: encoded,
	})

	asserts2.Get(t).Nil(sendNewUserToDiscord(req, user))
}
