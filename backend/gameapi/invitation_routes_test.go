package gameapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/services"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var errNotImplemented = errors.New("not implemented")

type stubLeagueService struct {
	preview    *services.InvitationPreview
	previewErr error
	lastToken  string
}

func (s *stubLeagueService) CreateLeague(ctx context.Context, name string) (*models.League, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) GetLeague(ctx context.Context, leagueID primitive.ObjectID) (*models.League, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) ListLeagues(ctx context.Context) ([]*models.League, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) ListActiveLeagues(ctx context.Context) ([]*models.League, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) ArchiveLeague(ctx context.Context, leagueID primitive.ObjectID) error {
	return errNotImplemented
}

func (s *stubLeagueService) UnarchiveLeague(ctx context.Context, leagueID primitive.ObjectID) error {
	return errNotImplemented
}

func (s *stubLeagueService) GetLeagueMembers(ctx context.Context, leagueID primitive.ObjectID) ([]*models.User, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) GetLeagueMemberships(ctx context.Context, leagueID primitive.ObjectID) ([]*services.LeagueMemberInfo, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) GetMemberByID(ctx context.Context, membershipID primitive.ObjectID) (*models.LeagueMembership, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) IsUserMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error) {
	return false, errNotImplemented
}

func (s *stubLeagueService) BanUserFromLeague(ctx context.Context, leagueID, userID primitive.ObjectID) error {
	return errNotImplemented
}

func (s *stubLeagueService) UnbanUserFromLeague(ctx context.Context, leagueID, userID primitive.ObjectID) error {
	return errNotImplemented
}

func (s *stubLeagueService) CreateMembershipForSuperAdmin(ctx context.Context, leagueID, userID primitive.ObjectID, alias string) (*models.LeagueMembership, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) CreateInvitation(ctx context.Context, leagueID, createdBy primitive.ObjectID, playerAlias string) (*models.LeagueInvitation, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) AcceptInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*models.League, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) PreviewInvitation(ctx context.Context, token string) (*services.InvitationPreview, error) {
	s.lastToken = token
	return s.preview, s.previewErr
}

func (s *stubLeagueService) GetInvitationByToken(ctx context.Context, token string) (*models.LeagueInvitation, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) ListMyInvitations(ctx context.Context, leagueID, userID primitive.ObjectID) ([]*models.LeagueInvitation, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) ListMyExpiredInvitations(ctx context.Context, leagueID, userID primitive.ObjectID) ([]*models.LeagueInvitation, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) CancelInvitation(ctx context.Context, token string, userID primitive.ObjectID) error {
	return errNotImplemented
}

func (s *stubLeagueService) ExtendInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*models.LeagueInvitation, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) UpdatePendingMemberAlias(ctx context.Context, membershipID primitive.ObjectID, userID primitive.ObjectID, newAlias string) error {
	return errNotImplemented
}

func (s *stubLeagueService) GetLeagueStandings(ctx context.Context, leagueID primitive.ObjectID) ([]*services.LeagueStanding, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) UpdatePlayersAfterGame(ctx context.Context, playerMembershipIDs []primitive.ObjectID) error {
	return errNotImplemented
}

func (s *stubLeagueService) GetSuggestedPlayers(ctx context.Context, leagueID, userID primitive.ObjectID, isSuperAdmin bool) (*services.SuggestedPlayersResponse, error) {
	return nil, errNotImplemented
}

func (s *stubLeagueService) GetMembershipByLeagueAndUser(ctx context.Context, leagueID, userID primitive.ObjectID) (*models.LeagueMembership, error) {
	return nil, errNotImplemented
}

var _ services.LeagueService = (*stubLeagueService)(nil)

func TestInvitationPreviewRouteIsPublic(t *testing.T) {
	expiresAt := time.Date(2026, 1, 18, 12, 30, 0, 0, time.UTC)
	service := &stubLeagueService{
		preview: &services.InvitationPreview{
			LeagueName:   "Test League",
			InviterAlias: "Alice",
			PlayerAlias:  "Bob",
			ExpiresAt:    expiresAt,
			Status:       "valid",
		},
	}
	handler := &Handler{leagueService: service}

	router := chi.NewRouter()
	var authHit bool
	router.Route("/api", func(r chi.Router) {
		handler.RegisterPublicRoutes(r)
		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					authHit = true
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				})
			})
			handler.RegisterRoutes(r, nil)
		})
	})

	req := httptest.NewRequest("GET", "/api/leagues/join/test-token/preview", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, authHit)
	assert.Equal(t, "test-token", service.lastToken)

	var response InvitationPreviewResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, service.preview.LeagueName, response.LeagueName)
	assert.Equal(t, service.preview.InviterAlias, response.InviterAlias)
	assert.Equal(t, service.preview.PlayerAlias, response.PlayerAlias)
	assert.Equal(t, service.preview.Status, response.Status)
	assert.Equal(t, expiresAt.Format("2006-01-02T15:04:05Z07:00"), response.ExpiresAt)
}
