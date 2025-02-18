package auth

import (
	"context"
	"github.com/andriyg76/bgl/asserts2"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"net/http/httptest"
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

// MockExternalAuthProvider implements ExternalAuthProvider interface
type MockExternalAuthProvider struct {
	mock.Mock
	ExternalAuthProvider
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
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
	store = sessions.NewCookieStore(utils.GenerateRandomKey(32))
	mockRepo := new(MockUserRepository)
	mockProvider := new(MockExternalAuthProvider)
	handler := GoogleCallbackHandler(mockRepo, mockProvider)

	superAdmins = []string{"superadmin@example.com"}

	existingUser := &models.User{
		ID:        primitive.NewObjectID(),
		Email:     "existing@example.com",
		Name:      "Existing User",
		Avatar:    "http://example.com/avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Alias:     "existing",
	}

	mockRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	mockProvider.On("CompleteUserAuth", mock.Anything, mock.Anything).Return(ExternalUser{
		Email:  "existing@example.com",
		Name:   "Existing User",
		Avatar: "http://example.com/avatar.jpg",
	}, nil)

	// Test existing user flow
	t.Run("Existing user login", func(t *testing.T) {
		discord := []string{}
		utils.AddDiscordSendCapturer(func(s string) {
			discord = append(discord, s)
		})
		asserts := asserts2.Get(t)

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
		req := httptest.NewRequest("GET", "/auth/callback?state=somestate&provider=google", nil)
		// Add the encoded cookie to the request
		req.AddCookie(&http.Cookie{
			Name:  "auth",
			Value: encoded,
		})
		rr := httptest.NewRecorder()

		// Set up session state
		session, _ := store.Get(req, "auth-session")
		session.Values["state"] = "somestate"
		asserts.Nil(session.Save(req, rr))

		handler.ServeHTTP(rr, req)

		// We expect an error here because gothic.CompleteUserAuth won't work in test
		// In a real scenario, you'd need to mock gothic.CompleteUserAuth
		asserts.Equal(http.StatusOK, rr.Code)
		asserts.True(len(discord) == 0, "No new notifications to discord")
	})
}

func TestLogoutHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockProvider := new(MockExternalAuthProvider)
	handler := LogoutHandler(mockRepo, mockProvider)

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

func TestGoogleCallbackHandlerNesUser(t *testing.T) {
	req, err := http.NewRequest("GET", "/callback", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	mockRepo := new(MockUserRepository)
	handler := http.HandlerFunc(GoogleCallbackHandler(mockRepo, new(MockExternalAuthProvider)))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what you expect
	expected := `expected response`
	if rr.Body.String() != expected {
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
