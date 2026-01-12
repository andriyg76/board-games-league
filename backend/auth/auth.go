package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/andriyg76/glog"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LogSuperAdmins logs the registered superadmins - call from main after all init() functions complete
func LogSuperAdmins() {
	glog.Info("Registered superadmins: %v", GetSuperAdmins())
}

var store = sessions.NewCookieStore(func() []byte {
	var secret = []byte(os.Getenv("SESSION_SECRET"))
	if len(secret) == 0 {
		glog.Warn("SESSION_SECRET is empty, generating session secret")
		secret = utils.GenerateRandomKey(32)
	} else {
		glog.Info("SESSION_SECRET is set with %d-th value", len(secret))
	}
	return secret
}())

func (h *Handler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth-session")
	state := r.URL.Query().Get("state")
	storedState := session.Values["state"]
	delete(session.Values, "state")

	if storedState != nil && state != storedState.(string) {
		_ = glog.Error("Auth completion failed: State token mismatch")
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	externalUser, err := h.provider.CompleteUserAuthHandler(w, r)
	if err != nil {
		_ = glog.Error("Auth completion failed: %v", err)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	var user *models.User
	var updateProfile bool

	// Check if googleUser exists in the collection
	if existingUser, err := h.userRepository.FindByExternalId(r.Context(), externalUser.ExternalIDs); err != nil {
		_ = glog.Error("error fetching user profile: %v", err)
		http.Error(w, "error fetching user profile", http.StatusInternalServerError)
		return
	} else if existingUser == nil {
		user = &models.User{
			ID:          primitive.ObjectID{},
			ExternalIDs: externalUser.ExternalIDs,
			Name:        externalUser.Name,
			Avatar:      externalUser.Avatar,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Alias:       "",
			Names:       []string{externalUser.Name},
			Avatars:     []string{externalUser.Avatar},
		}
		if isSuperAdmin(externalUser.ExternalIDs) {

			if alias, err := utils.GetUniqueAlias(func(alias string) (bool, error) {
				return h.userRepository.AliasUnique(r.Context(), alias)
			}); err != nil {
				utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err,
					"error checking alias uniqueness")
				return
			} else {
				user.Alias = alias
			}

			// Create googleUser in the collection
			if err := h.userRepository.Create(r.Context(), user); err != nil {
				_ = glog.Error("failed to create user", err)
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
				return
			}
		} else {
			// Send googleUser info to Discord webhook
			_ = sendNewUserToDiscord(r, user)
			glog.Info("User with externalID %v is not known", user.ExternalIDs)
			http.Error(w, "Unauthorised", http.StatusUnauthorized)
			return
		}
	} else {
		user = existingUser

		if user.Alias == "" {
			if alias, err := utils.GetUniqueAlias(func(alias string) (bool, error) {
				return h.userRepository.AliasUnique(r.Context(), alias)
			}); err != nil {
				utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err,
					"error fetching user profile")
				return
			} else {
				user.Alias = alias
				updateProfile = true
			}
		}

		if externalUser.Name != "" && !slices.Contains(user.Names, externalUser.Name) {
			user.Names = append(user.Names, externalUser.Name)
			updateProfile = true
		}
		if externalUser.Avatar != "" && !slices.Contains(user.Avatars, externalUser.Avatar) {
			user.Avatars = append(user.Avatars, externalUser.Avatar)
			updateProfile = true
		}

		for _, id := range externalUser.ExternalIDs {
			if !slices.Contains(user.ExternalIDs, id) {
				user.ExternalIDs = append(user.ExternalIDs, id)
				updateProfile = true
			}
		}
	}

	if updateProfile {
		user.UpdatedAt = time.Now()
		if err := h.userRepository.Update(r.Context(), user); err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err,
				"error updating user profile")
			return
		}
	}

	// Parse request info
	reqInfo := h.requestService.ParseRequest(r, nil)
	userCode := utils.IdToCode(user.ID)

	rotateToken, actionToken, err := h.sessionService.CreateSession(
		r.Context(),
		user.ID,
		userCode,
		user.ExternalIDs,
		user.Name,
		user.Avatar,
		reqInfo.ClientIP(),
		reqInfo.UserAgent(),
	)
	if err != nil {
		_ = glog.Error("Session creation failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set action token cookie (1 hour expiry)
	http.SetCookie(w, reqInfo.NewCookie(authCookieName, actionToken, 60*60))

	// Return user data with rotate token (for client to store in localStorage)
	response := struct {
		user_profile.UserResponse
		RotateToken string `json:"rotateToken"`
	}{
		UserResponse: user_profile.UserResponse{
			Code:        userCode,
			ExternalIDs: user.ExternalIDs,
			Name:        user.Name,
			Avatar:      user.Avatar,
			Alias:       user.Alias,
			Avatars:     user.Avatars,
			Names:       user.Names,
		},
		RotateToken: rotateToken,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		_ = glog.Error("serialising error %v", err)
		http.Error(w, "serialising error", http.StatusInternalServerError)
	}
}

func (h *Handler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Extract rotate token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Expect "Bearer <token>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}
	rotateToken := parts[1]

	// Parse request info
	reqInfo := h.requestService.ParseRequest(r, nil)

	// Refresh action token (and rotate token if needed)
	newRotateToken, actionToken, err := h.sessionService.RefreshActionToken(r.Context(), rotateToken, reqInfo.ClientIP(), reqInfo.UserAgent())
	if err != nil {
		_ = glog.Error("Token refresh failed: %v", err)
		http.Error(w, "Token refresh failed", http.StatusUnauthorized)
		return
	}

	// Set new action token cookie
	http.SetCookie(w, reqInfo.NewCookie(authCookieName, actionToken, 60*60))

	// Return new rotate token if it was rotated
	response := make(map[string]interface{})
	if newRotateToken != "" {
		response["rotateToken"] = newRotateToken
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		_ = glog.Error("Failed to encode refresh response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	_ = h.provider.LogoutHandler(w, r)

	// Try to get rotate token from request body or Authorization header
	var rotateToken string
	if r.Header.Get("Content-Type") == "application/json" {
		var body struct {
			RotateToken string `json:"rotateToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			rotateToken = body.RotateToken
		}
	}

	// Fallback to Authorization header
	if rotateToken == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				rotateToken = parts[1]
			}
		}
	}

	// Invalidate session if rotate token provided
	if rotateToken != "" {
		if err := h.sessionService.InvalidateSession(r.Context(), rotateToken); err != nil {
			glog.Warn("Failed to invalidate session: %v", err)
		}
	}

	// Clear the auth cookie
	reqInfo := h.requestService.ParseRequest(r, nil)
	http.SetCookie(w, reqInfo.ClearCookie(authCookieName))

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleBeginLoginFlow(w http.ResponseWriter, r *http.Request) {
	_ = h.provider.LogoutHandler(w, r)

	query := r.URL.Query()
	if state := query.Get("state"); state != "" {
		session, err := store.Get(r, "auth-session")
		if err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error handling session context")
			return
		}
		session.Values["state"] = state
		if err := session.Save(r, w); err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "serialising session error")
			return
		}
	}

	h.provider.BeginUserAuthHandler(w, r)
}

// clearCookies clears the auth cookie (without request context, uses safe defaults)
func clearCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

var config = struct {
	GoogleClientID      string
	GoogleClientSecret  string
	DiscordClientID     string
	DiscordClientSecret string
}{
	GoogleClientID:      os.Getenv("GOOGLE_CLIENT_ID"),
	GoogleClientSecret:  os.Getenv("GOOGLE_CLIENT_SECRET"),
	DiscordClientID:     os.Getenv("DISCORD_CLIENT_ID"),
	DiscordClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
}

const authCookieName = "auth_token"

func (h *Handler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(authCookieName)
		if err != nil {
			handleUnauthorized(w, r)
			return
		}

		// Parse and validate JWT (expiration is checked automatically by jwt library)
		profile, err := user_profile.ParseProfile(cookie.Value)
		if err != nil {
			// Token is invalid or expired - client should refresh
			handleUnauthorized(w, r)
			return
		}

		if len(profile.ExternalIDs) != 0 {
			ctx := context.WithValue(r.Context(), "user", profile)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			handleUnauthorized(w, r)
		}
	})
}

func handleUnauthorized(w http.ResponseWriter, _ *http.Request) {
	clearCookies(w)

	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func isSuperAdmin(ids []string) bool {
	return IsSuperAdminByExternalIDs(ids)
}

func sendNewUserToDiscord(r *http.Request, user *models.User) error {
	if user == nil {
		buf := make([]byte, 4096)
		length := runtime.Stack(buf, false)
		stackTrace := string(buf[:length])

		_ = glog.Error("sendNewUserToDiscord: User is not set. Stack trace: %s, request: %V", stackTrace, r)

		_ = utils.SendToDiscord(fmt.Sprintf("User is not set. Stack trace: %s, request: %v", stackTrace, r))

		return glog.Error("User is not set")
	}
	domain := utils.GetHostUrl(r)
	createUserLink := fmt.Sprintf("%s/ui/admin/create-user?external_ids=%s", domain, strings.Join(user.ExternalIDs, ",")) // domain defined at frontend/src/router/index.ts
	content := fmt.Sprintf("New user login: %s (%s). Click [%s] to create the user.", user.Name, strings.Join(user.ExternalIDs, ","), createUserLink)

	return utils.SendToDiscord(content)
}

type Handler struct {
	userRepository repositories.UserRepository
	sessionService services.SessionService
	requestService services.RequestService
	provider       ExternalAuthProvider
}

func NewHandler(repository repositories.UserRepository, sessionService services.SessionService, requestService services.RequestService, provider ExternalAuthProvider) Handler {
	return Handler{
		provider:       provider,
		userRepository: repository,
		sessionService: sessionService,
		requestService: requestService,
	}
}

func NewDefaultHandler(repository repositories.UserRepository, sessionService services.SessionService, requestService services.RequestService) Handler {
	return Handler{
		provider:       &authProviderInstance{},
		userRepository: repository,
		sessionService: sessionService,
		requestService: requestService,
	}
}
