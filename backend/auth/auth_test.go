package auth

import (
	"context"
	"fmt"
	"github.com/andriyg76/bgl/asserts2"
	"github.com/andriyg76/bgl/repositories"
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

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
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

// MockExternalAuthProvider implements ExternalAuthProvider interface
type MockExternalAuthProvider struct {
	mock.Mock
	ExternalAuthProvider
}

func (m *MockExternalAuthProvider) BeginUserAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusPermanentRedirect)
	w.Header().Add("Location", fmt.Sprintf("http://google.com/auth?state=%s", r.URL.Query().Get("state")))
}

func (m *MockExternalAuthProvider) CompleteUserAuthHandler(_ http.ResponseWriter, r *http.Request) (ExternalUser, error) {
	args := m.Called(r.Context(), r)
	return args.Get(0).(ExternalUser), args.Error(1)
}

func (m *MockExternalAuthProvider) LogoutHandler(_ http.ResponseWriter, r *http.Request) error {
	args := m.Called(r.Context(), r)
	return args.Error(0)
}

func TestIsSuperAdmin(t *testing.T) {
	// Temporarily set superAdmins for testing
	originalSuperAdmins := superAdmins
	superAdmins = []string{"admin@example.com", "super@example.com"}
	defer func() { superAdmins = originalSuperAdmins }()

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"Super admin email should return true", "admin@example.com", true},
		{"Regular email should return false", "user@example.com", false},
		{"Empty email should return false", "", false},
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
	middleware := Middleware(mockRepo)

	tests := []struct {
		name           string
		setupAuth      func(*http.Request)
		expectedStatus int
	}{
		{
			name: "Valid token should pass",
			setupAuth: func(r *http.Request) {
				token, _ := user_profile.CreateAuthToken(
					"test@example.com",
					"Test User",
					"http://example.com/avatar.jpg",
				)
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
	beginFlowHandler := HandleBeginLoginFlow(mockProvider)
	handler := GoogleCallbackHandler(mockRepo, mockProvider)

	superAdminEmail := "superadmin@example.com"
	superAdmins = []string{superAdminEmail}

	notExistingEmail := "notexisting@example.com"

	regularEmail := "existing@example.com"
	existingUser := &models.User{
		ID:        primitive.NewObjectID(),
		Email:     regularEmail,
		Name:      "Existing User",
		Avatar:    "http://example.com/avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Alias:     "existing",
	}

	mockRepo.On("FindByEmail", mock.Anything, regularEmail).Return(existingUser, nil)
	mockRepo.On("FindByEmail", mock.Anything, superAdminEmail).Return(nil, nil)
	mockRepo.On("FindByEmail", mock.Anything, notExistingEmail).Return(nil, nil)
	mockRepo.On("AliasUnique", mock.Anything, mock.Anything).Return(true, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	superadminRequest := httptest.NewRequest("GET", "/auth/callback?state=somestate&provider=google", nil)
	mockProvider.On("CompleteUserAuthHandler", mock.Anything, superadminRequest).Return(ExternalUser{
		Email:  superAdminEmail,
		Name:   "Superadmin",
		Avatar: "http://example.com/avatar.jpg",
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
		beginFlowHandler.ServeHTTP(r1, req)

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
			Email:  regularEmail,
			Name:   "Existing User",
			Avatar: "http://example.com/avatar.jpg",
		}, nil)

		handler.ServeHTTP(rr, regularUserRequest)

		// We expect an error here because gothic.CompleteUserAuthHandler won't work in test
		// In a real scenario, you'd need to mock gothic.CompleteUserAuthHandler
		asserts.
			Equal(http.StatusOK, rr.Code).
			True(len(discord) == 0, "No new notifications to discord").
			True(strings.Contains(rr.Header().Get("Set-Cookie"), "auth_token="),
				"auth_token cookie should be set by end auth")
		for _, cookie := range rr.Result().Cookies() {
			if cookie.Name == "auth_token" {
				profile, err := user_profile.ParseProfile(cookie.Value)
				if err != nil {
					asserts.NoError(err, "Coud not parse cookie %v", cookie)
				}
				asserts.Equal(regularEmail, profile.Email, "Invalid user token set")
			}
		}

	})

	t.Run("Superamdin user login", func(t *testing.T) {
		const somestate = "somestate2"

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
		beginFlowHandler.ServeHTTP(r1, req)

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
		request := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?state=%s", somestate), nil)
		mockProvider.On("CompleteUserAuthHandler", mock.Anything, request).Return(ExternalUser{
			Email:  superAdminEmail,
			Name:   "superadmin",
			Avatar: "sinine.jpg",
		}, nil)

		handler.ServeHTTP(rr, request)

		// We expect an error here because gothic.CompleteUserAuthHandler won't work in test
		// In a real scenario, you'd need to mock gothic.CompleteUserAuthHandler
		asserts.
			Equal(http.StatusOK, rr.Code).
			True(len(discord) == 0, "No new notifications to discord").
			True(strings.Contains(rr.Header().Get("Set-Cookie"), "auth_token="),
				"auth_token cookie should be set by end auth")
		for _, cookie := range rr.Result().Cookies() {
			if cookie.Name == "auth_token" {
				profile, err := user_profile.ParseProfile(cookie.Value)
				if err != nil {
					asserts.NoError(err, "Coud not parse cookie %v", cookie)
				}
				asserts.Equal(superAdminEmail, profile.Email, "Invalid user token set")
			}
		}
	})
}

func TestLogoutHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockProvider := new(MockExternalAuthProvider)
	handler := LogoutHandler(mockRepo, mockProvider)

	mockProvider.On("LogoutHandler", mock.Anything, mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/logout", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

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
		Name:  "Test User",
		Email: "test@example.com",
	}

	req := httptest.NewRequest("POST", "/send", nil)
	// Add the encoded cookie to the request
	req.AddCookie(&http.Cookie{
		Name:  "auth",
		Value: encoded,
	})

	asserts2.Get(t).Nil(sendNewUserToDiscord(req, user))
}
