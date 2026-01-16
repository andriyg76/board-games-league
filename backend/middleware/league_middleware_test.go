package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockLeagueService is a mock implementation of LeagueService
type MockLeagueService struct {
	mock.Mock
}

func (m *MockLeagueService) GetLeague(ctx context.Context, leagueID primitive.ObjectID) (*models.League, error) {
	args := m.Called(ctx, leagueID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.League), args.Error(1)
}

func (m *MockLeagueService) GetMembershipByLeagueAndUser(ctx context.Context, leagueID, userID primitive.ObjectID) (*models.LeagueMembership, error) {
	args := m.Called(ctx, leagueID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LeagueMembership), args.Error(1)
}

func (m *MockLeagueService) GetInvitationByToken(ctx context.Context, token string) (*models.LeagueInvitation, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LeagueInvitation), args.Error(1)
}

func (m *MockLeagueService) IsUserMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error) {
	args := m.Called(ctx, leagueID, userID)
	return args.Bool(0), args.Error(1)
}

// Stub implementations for other interface methods
func (m *MockLeagueService) CreateLeague(ctx context.Context, name string) (*models.League, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) ListLeagues(ctx context.Context) ([]*models.League, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) ListActiveLeagues(ctx context.Context) ([]*models.League, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) ArchiveLeague(ctx context.Context, leagueID primitive.ObjectID) error {
	return errors.New("not implemented")
}

func (m *MockLeagueService) UnarchiveLeague(ctx context.Context, leagueID primitive.ObjectID) error {
	return errors.New("not implemented")
}

func (m *MockLeagueService) GetLeagueMembers(ctx context.Context, leagueID primitive.ObjectID) ([]*models.User, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) GetLeagueMemberships(ctx context.Context, leagueID primitive.ObjectID) ([]*services.LeagueMemberInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) GetMemberByID(ctx context.Context, membershipID primitive.ObjectID) (*models.LeagueMembership, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) BanUserFromLeague(ctx context.Context, leagueID, userID primitive.ObjectID) error {
	return errors.New("not implemented")
}

func (m *MockLeagueService) UnbanUserFromLeague(ctx context.Context, leagueID, userID primitive.ObjectID) error {
	return errors.New("not implemented")
}

func (m *MockLeagueService) CreateMembershipForSuperAdmin(ctx context.Context, leagueID, userID primitive.ObjectID, alias string) (*models.LeagueMembership, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) CreateInvitation(ctx context.Context, leagueID, createdBy primitive.ObjectID, playerAlias string) (*models.LeagueInvitation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) AcceptInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*models.League, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) PreviewInvitation(ctx context.Context, token string) (*services.InvitationPreview, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) ListMyInvitations(ctx context.Context, leagueID, userID primitive.ObjectID) ([]*models.LeagueInvitation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) ListMyExpiredInvitations(ctx context.Context, leagueID, userID primitive.ObjectID) ([]*models.LeagueInvitation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) CancelInvitation(ctx context.Context, token string, userID primitive.ObjectID) error {
	return errors.New("not implemented")
}

func (m *MockLeagueService) ExtendInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*models.LeagueInvitation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) UpdatePendingMemberAlias(ctx context.Context, membershipID primitive.ObjectID, userID primitive.ObjectID, newAlias string) error {
	return errors.New("not implemented")
}

func (m *MockLeagueService) GetLeagueStandings(ctx context.Context, leagueID primitive.ObjectID) ([]*services.LeagueStanding, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLeagueService) UpdatePlayersAfterGame(ctx context.Context, playerMembershipIDs []primitive.ObjectID) error {
	return errors.New("not implemented")
}

func (m *MockLeagueService) GetSuggestedPlayers(ctx context.Context, leagueID primitive.ObjectID, userID primitive.ObjectID, isSuperAdmin bool) (*services.SuggestedPlayersResponse, error) {
	return nil, errors.New("not implemented")
}

// MockIdAndCodeCache is a mock implementation of IdAndCodeCache
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

func createTestUserProfile(code string, externalIDs []string) *user_profile.UserProfile {
	return &user_profile.UserProfile{
		Code:        code,
		ExternalIDs: externalIDs,
		Name:        "Test User",
		Picture:     "",
	}
}

func createTestLeague(id primitive.ObjectID) *models.League {
	return &models.League{
		ID:        id,
		Version:   1,
		Name:      "Test League",
		Status:    models.LeagueActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestMembership(leagueID, userID primitive.ObjectID, status models.LeagueMembershipStatus) *models.LeagueMembership {
	return &models.LeagueMembership{
		ID:        primitive.NewObjectID(),
		Version:   1,
		LeagueID:  leagueID,
		UserID:    userID,
		Status:    status,
		JoinedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestRequireLeagueMembership_Unauthorized(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	handler := middleware.RequireLeagueMembership(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/leagues/testcode", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Unauthorized")
}

func TestRequireLeagueMembership_MissingLeagueCode(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userID := primitive.NewObjectID()
	userCode := utils.IdToCode(userID)
	userProfile := createTestUserProfile(userCode, []string{"test@example.com"})
	handler := middleware.RequireLeagueMembership(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/leagues/", nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", "") // Empty code
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	userIdAndCode := models.NewIdAndCode(userID)
	mockIdCodeCache.On("GetByCode", userCode).Return(userIdAndCode, nil)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "League code is required")
	mockIdCodeCache.AssertExpectations(t)
}

func TestRequireLeagueMembership_InvalidLeagueCode(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userID := primitive.NewObjectID()
	userCode := utils.IdToCode(userID)
	userProfile := createTestUserProfile(userCode, []string{"test@example.com"})
	handler := middleware.RequireLeagueMembership(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/leagues/invalidcode", nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", "invalidcode")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	userIdAndCode := models.NewIdAndCode(userID)
	mockIdCodeCache.On("GetByCode", userCode).Return(userIdAndCode, nil)
	mockIdCodeCache.On("GetByCode", "invalidcode").Return(nil, errors.New("invalid code"))

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid league code")
	mockIdCodeCache.AssertExpectations(t)
}

func TestRequireLeagueMembership_LeagueNotFound(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userID := primitive.NewObjectID()
	leagueID := primitive.NewObjectID()
	userCode := utils.IdToCode(userID)
	leagueCode := utils.IdToCode(leagueID)

	userProfile := createTestUserProfile(userCode, []string{"test@example.com"})
	handler := middleware.RequireLeagueMembership(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/leagues/"+leagueCode, nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", leagueCode)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	userIdAndCode := models.NewIdAndCode(userID)
	leagueIdAndCode := models.NewIdAndCode(leagueID)

	mockIdCodeCache.On("GetByCode", userCode).Return(userIdAndCode, nil)
	mockIdCodeCache.On("GetByCode", leagueCode).Return(leagueIdAndCode, nil)
	mockLeagueService.On("GetLeague", mock.Anything, leagueID).Return(nil, errors.New("not found"))

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "League not found")
	mockIdCodeCache.AssertExpectations(t)
	mockLeagueService.AssertExpectations(t)
}

func TestRequireLeagueMembership_NotMember(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userID := primitive.NewObjectID()
	leagueID := primitive.NewObjectID()
	userCode := utils.IdToCode(userID)
	leagueCode := utils.IdToCode(leagueID)

	userProfile := createTestUserProfile(userCode, []string{"test@example.com"})
	handler := middleware.RequireLeagueMembership(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/leagues/"+leagueCode, nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", leagueCode)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	userIdAndCode := models.NewIdAndCode(userID)
	leagueIdAndCode := models.NewIdAndCode(leagueID)
	league := createTestLeague(leagueID)

	mockIdCodeCache.On("GetByCode", userCode).Return(userIdAndCode, nil)
	mockIdCodeCache.On("GetByCode", leagueCode).Return(leagueIdAndCode, nil)
	mockLeagueService.On("GetLeague", mock.Anything, leagueID).Return(league, nil)
	mockLeagueService.On("GetMembershipByLeagueAndUser", mock.Anything, leagueID, userID).Return(nil, errors.New("not found"))

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "not a member")
	mockIdCodeCache.AssertExpectations(t)
	mockLeagueService.AssertExpectations(t)
}

func TestRequireLeagueMembership_Success(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userID := primitive.NewObjectID()
	leagueID := primitive.NewObjectID()
	userCode := utils.IdToCode(userID)
	leagueCode := utils.IdToCode(leagueID)

	userProfile := createTestUserProfile(userCode, []string{"test@example.com"})
	handler := middleware.RequireLeagueMembership(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context values are set
		league := r.Context().Value("league").(*models.League)
		membership := r.Context().Value("membership").(*models.LeagueMembership)
		ctxLeagueID := r.Context().Value("leagueID").(primitive.ObjectID)

		assert.Equal(t, leagueID, league.ID)
		assert.Equal(t, leagueID, ctxLeagueID)
		assert.Equal(t, userID, membership.UserID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/leagues/"+leagueCode, nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", leagueCode)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	userIdAndCode := models.NewIdAndCode(userID)
	leagueIdAndCode := models.NewIdAndCode(leagueID)
	league := createTestLeague(leagueID)
	membership := createTestMembership(leagueID, userID, models.MembershipActive)

	mockIdCodeCache.On("GetByCode", userCode).Return(userIdAndCode, nil)
	mockIdCodeCache.On("GetByCode", leagueCode).Return(leagueIdAndCode, nil)
	mockLeagueService.On("GetLeague", mock.Anything, leagueID).Return(league, nil)
	mockLeagueService.On("GetMembershipByLeagueAndUser", mock.Anything, leagueID, userID).Return(membership, nil)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockIdCodeCache.AssertExpectations(t)
	mockLeagueService.AssertExpectations(t)
}

func TestRequireLeagueMembership_SuperAdminAccess(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userID := primitive.NewObjectID()
	leagueID := primitive.NewObjectID()
	userCode := utils.IdToCode(userID)
	leagueCode := utils.IdToCode(leagueID)

	// Use superadmin email - set up test superadmin
	restore := auth.SetSuperAdminsForTesting([]string{"admin@test.com"})
	defer restore()
	userProfile := createTestUserProfile(userCode, []string{"admin@test.com"})
	handler := middleware.RequireLeagueMembership(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context values are set
		league := r.Context().Value("league").(*models.League)
		membership := r.Context().Value("membership")
		ctxLeagueID := r.Context().Value("leagueID").(primitive.ObjectID)

		assert.Equal(t, leagueID, league.ID)
		assert.Equal(t, leagueID, ctxLeagueID)
		assert.Nil(t, membership) // Superadmin has no membership
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/leagues/"+leagueCode, nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", leagueCode)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	userIdAndCode := models.NewIdAndCode(userID)
	leagueIdAndCode := models.NewIdAndCode(leagueID)
	league := createTestLeague(leagueID)

	mockIdCodeCache.On("GetByCode", userCode).Return(userIdAndCode, nil)
	mockIdCodeCache.On("GetByCode", leagueCode).Return(leagueIdAndCode, nil)
	mockLeagueService.On("GetLeague", mock.Anything, leagueID).Return(league, nil)
	mockLeagueService.On("GetMembershipByLeagueAndUser", mock.Anything, leagueID, userID).Return(nil, errors.New("not found"))

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockIdCodeCache.AssertExpectations(t)
	mockLeagueService.AssertExpectations(t)
}

func TestRequireLeagueMembership_InactiveMembership(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userID := primitive.NewObjectID()
	leagueID := primitive.NewObjectID()
	userCode := utils.IdToCode(userID)
	leagueCode := utils.IdToCode(leagueID)

	userProfile := createTestUserProfile(userCode, []string{"test@example.com"})
	handler := middleware.RequireLeagueMembership(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/leagues/"+leagueCode, nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", leagueCode)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	userIdAndCode := models.NewIdAndCode(userID)
	leagueIdAndCode := models.NewIdAndCode(leagueID)
	league := createTestLeague(leagueID)
	membership := createTestMembership(leagueID, userID, models.MembershipBanned)

	mockIdCodeCache.On("GetByCode", userCode).Return(userIdAndCode, nil)
	mockIdCodeCache.On("GetByCode", leagueCode).Return(leagueIdAndCode, nil)
	mockLeagueService.On("GetLeague", mock.Anything, leagueID).Return(league, nil)
	mockLeagueService.On("GetMembershipByLeagueAndUser", mock.Anything, leagueID, userID).Return(membership, nil)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "not active")
	mockIdCodeCache.AssertExpectations(t)
	mockLeagueService.AssertExpectations(t)
}

func TestRequireSuperAdmin_Unauthorized(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	handler := middleware.RequireSuperAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Unauthorized")
}

func TestRequireSuperAdmin_NotSuperAdmin(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userProfile := createTestUserProfile("usercode", []string{"test@example.com"})
	handler := middleware.RequireSuperAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Superadmin privileges required")
}

func TestRequireSuperAdmin_Success(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	// Set up test superadmin
	restore := auth.SetSuperAdminsForTesting([]string{"admin@test.com"})
	defer restore()
	userProfile := createTestUserProfile("usercode", []string{"admin@test.com"})
	handler := middleware.RequireSuperAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireLeagueMembershipByToken_Unauthorized(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	handler := middleware.RequireLeagueMembershipByToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/invitations/token", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireLeagueMembershipByToken_MissingToken(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userID := primitive.NewObjectID()
	userCode := utils.IdToCode(userID)
	userProfile := createTestUserProfile(userCode, []string{"test@example.com"})
	handler := middleware.RequireLeagueMembershipByToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/invitations/", nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "") // Empty token
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	userIdAndCode := models.NewIdAndCode(userID)
	mockIdCodeCache.On("GetByCode", userCode).Return(userIdAndCode, nil)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invitation token is required")
	mockIdCodeCache.AssertExpectations(t)
}

func TestRequireLeagueMembershipByToken_InvalidToken(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userID := primitive.NewObjectID()
	userCode := utils.IdToCode(userID)

	userProfile := createTestUserProfile(userCode, []string{"test@example.com"})
	handler := middleware.RequireLeagueMembershipByToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/invitations/invalidtoken", nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "invalidtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	userIdAndCode := models.NewIdAndCode(userID)

	mockIdCodeCache.On("GetByCode", userCode).Return(userIdAndCode, nil)
	mockLeagueService.On("GetInvitationByToken", mock.Anything, "invalidtoken").Return(nil, errors.New("not found"))

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired invitation")
	mockIdCodeCache.AssertExpectations(t)
	mockLeagueService.AssertExpectations(t)
}

func TestRequireLeagueMembershipByToken_Success(t *testing.T) {
	mockLeagueService := new(MockLeagueService)
	mockIdCodeCache := new(MockIdAndCodeCache)
	middleware := NewLeagueMiddleware(mockLeagueService, mockIdCodeCache)

	userID := primitive.NewObjectID()
	leagueID := primitive.NewObjectID()
	userCode := utils.IdToCode(userID)
	token := "validtoken"

	userProfile := createTestUserProfile(userCode, []string{"test@example.com"})
	handler := middleware.RequireLeagueMembershipByToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context values are set
		isMember := r.Context().Value("isMember").(bool)
		invitation := r.Context().Value("invitation").(*models.LeagueInvitation)

		assert.False(t, isMember)
		assert.Equal(t, leagueID, invitation.LeagueID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/invitations/"+token, nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", token)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	userIdAndCode := models.NewIdAndCode(userID)
	invitation := &models.LeagueInvitation{
		ID:       primitive.NewObjectID(),
		LeagueID: leagueID,
		Token:    token,
	}

	mockIdCodeCache.On("GetByCode", userCode).Return(userIdAndCode, nil)
	mockLeagueService.On("GetInvitationByToken", mock.Anything, token).Return(invitation, nil)
	mockLeagueService.On("IsUserMember", mock.Anything, leagueID, userID).Return(false, nil)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockIdCodeCache.AssertExpectations(t)
	mockLeagueService.AssertExpectations(t)
}
