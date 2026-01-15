package middleware

import (
	"context"
	"net/http"

	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
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
// and loads the league and membership objects into the request context
func (m *LeagueMiddleware) RequireLeagueMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user profile from context (set by authentication middleware)
		profile, ok := r.Context().Value("user").(*user_profile.UserProfile)
		if !ok || profile == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Convert user code to ObjectID
		userID, err := utils.CodeToID(profile.Code)
		if err != nil {
			http.Error(w, "Invalid user code", http.StatusBadRequest)
			return
		}

		// Get league code from URL parameter
		leagueCode := chi.URLParam(r, "code")
		if leagueCode == "" {
			http.Error(w, "League code is required", http.StatusBadRequest)
			return
		}

		// Parse league code using CodeToID (base64 encoded, not hex)
		leagueID, err := utils.CodeToID(leagueCode)
		if err != nil {
			http.Error(w, "Invalid league code", http.StatusBadRequest)
			return
		}

		// Load league object
		league, err := m.leagueService.GetLeague(r.Context(), leagueID)
		if err != nil {
			http.Error(w, "League not found", http.StatusNotFound)
			return
		}

		// Check if user is superadmin
		isSuperAdmin := auth.IsSuperAdminByExternalIDs(profile.ExternalIDs)

		// Load membership object
		membership, err := m.leagueService.GetMembershipByLeagueAndUser(r.Context(), leagueID, userID)
		if err != nil {
			// User is not a member - check if superadmin
			if !isSuperAdmin {
				http.Error(w, "You are not a member of this league", http.StatusForbidden)
				return
			}
			// Superadmin can access without membership
			membership = nil
		}

		// Verify membership is active (if not superadmin)
		if membership != nil && membership.Status != "active" && !isSuperAdmin {
			http.Error(w, "Your membership is not active", http.StatusForbidden)
			return
		}

		// Add league, membership, and leagueID to context for use in handlers
		ctx := context.WithValue(r.Context(), "league", league)
		ctx = context.WithValue(ctx, "membership", membership)
		ctx = context.WithValue(ctx, "leagueID", leagueID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireSuperAdmin verifies that the authenticated user has superadmin privileges
func (m *LeagueMiddleware) RequireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user profile from context (set by authentication middleware)
		profile, ok := r.Context().Value("user").(*user_profile.UserProfile)
		if !ok || profile == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user is superadmin
		if !auth.IsSuperAdminByExternalIDs(profile.ExternalIDs) {
			http.Error(w, "Superadmin privileges required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireLeagueMembershipByToken verifies league membership for endpoints that use token instead of code
func (m *LeagueMiddleware) RequireLeagueMembershipByToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user profile from context (set by authentication middleware)
		profile, ok := r.Context().Value("user").(*user_profile.UserProfile)
		if !ok || profile == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Convert user code to ObjectID
		userID, err := utils.CodeToID(profile.Code)
		if err != nil {
			http.Error(w, "Invalid user code", http.StatusBadRequest)
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
		isMember, err := m.leagueService.IsUserMember(r.Context(), invitation.LeagueID, userID)
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
