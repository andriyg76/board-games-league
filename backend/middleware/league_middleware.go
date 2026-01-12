package middleware

import (
	"context"
	"net/http"

	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/services"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LeagueMiddleware provides middleware functions for league access control
type LeagueMiddleware struct {
	leagueService services.LeagueService
}

// NewLeagueMiddleware creates a new league middleware instance
func NewLeagueMiddleware(leagueService services.LeagueService) *LeagueMiddleware {
	return &LeagueMiddleware{
		leagueService: leagueService,
	}
}

// RequireLeagueMembership verifies that the authenticated user is an active member of the league
func (m *LeagueMiddleware) RequireLeagueMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (set by authentication middleware)
		user, ok := r.Context().Value("user").(*models.User)
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get league code from URL parameter
		leagueCode := chi.URLParam(r, "code")
		if leagueCode == "" {
			http.Error(w, "League code is required", http.StatusBadRequest)
			return
		}

		// Parse league code as ObjectID
		leagueID, err := primitive.ObjectIDFromHex(leagueCode)
		if err != nil {
			http.Error(w, "Invalid league code", http.StatusBadRequest)
			return
		}

		// Check if user is a member of the league
		isMember, err := m.leagueService.IsUserMember(r.Context(), leagueID, user.ID)
		if err != nil {
			http.Error(w, "Failed to check league membership", http.StatusInternalServerError)
			return
		}

		if !isMember {
			http.Error(w, "You are not a member of this league", http.StatusForbidden)
			return
		}

		// Add league ID to context for use in handlers
		ctx := context.WithValue(r.Context(), "leagueID", leagueID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireSuperAdmin verifies that the authenticated user has superadmin privileges
func (m *LeagueMiddleware) RequireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (set by authentication middleware)
		user, ok := r.Context().Value("user").(*models.User)
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user is superadmin
		if !auth.IsSuperAdmin(user) {
			http.Error(w, "Superadmin privileges required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireLeagueMembershipByToken verifies league membership for endpoints that use token instead of code
func (m *LeagueMiddleware) RequireLeagueMembershipByToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (set by authentication middleware)
		user, ok := r.Context().Value("user").(*models.User)
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get token from URL parameter
		token := chi.URLParam(r, "token")
		if token == "" {
			http.Error(w, "Invitation token is required", http.StatusBadRequest)
			return
		}

		// Get invitation to extract league ID
		invitation, err := m.leagueService.GetInvitationByToken(r.Context(), token)
		if err != nil {
			http.Error(w, "Invalid or expired invitation", http.StatusNotFound)
			return
		}

		// Check if user is already a member
		isMember, err := m.leagueService.IsUserMember(r.Context(), invitation.LeagueID, user.ID)
		if err != nil {
			http.Error(w, "Failed to check league membership", http.StatusInternalServerError)
			return
		}

		// Add membership status to context
		ctx := context.WithValue(r.Context(), "isMember", isMember)
		ctx = context.WithValue(ctx, "invitation", invitation)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
