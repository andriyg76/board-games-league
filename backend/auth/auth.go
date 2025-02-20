package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/andriyg76/glog"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// Load super admins from environment variable
var superAdmins = strings.Split(os.Getenv("SUPERADMINS"), ",")

func init() {
	glog.Info("Registered superadmins: %v", superAdmins)
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

func GoogleCallbackHandler(repository repositories.UserRepository, provider ExternalAuthProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth-session")
		state := r.URL.Query().Get("state")
		storedState := session.Values["state"]
		delete(session.Values, "state")

		if storedState != nil && state != storedState.(string) {
			_ = glog.Error("Auth completion failed: State token mismatch")
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		externalUser, err := provider.CompleteUserAuthHandler(w, r)
		if err != nil {
			_ = glog.Error("Auth completion failed: %v", err)
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		var user *models.User

		// Check if googleUser exists in the collection
		if existingUser, err := repository.FindByExternalId(r.Context(), externalUser.ExternalIDs...); err != nil {
			_ = glog.Error("error fetching user profile: %v", err)
			http.Error(w, "error fetching user profile", http.StatusInternalServerError)
			return
		} else if existingUser == nil {
			user = &models.User{
				ID:         primitive.ObjectID{},
				ExternalID: externalUser.ExternalIDs,
				Name:       externalUser.Name,
				Avatar:     externalUser.Avatar,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				Alias:      "",
			}
			if isSuperAdmin(externalUser.ExternalIDs) {

				if alias, err := utils.GetUniqueAlias(func(alias string) (bool, error) {
					return repository.AliasUnique(r.Context(), alias)
				}); err != nil {
					utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err,
						"error fetching user profile")
					return
				} else {
					user.Alias = alias
				}

				// Create googleUser in the collection
				if err := repository.Create(r.Context(), user); err != nil {
					_ = glog.Error("failed to create user", err)
					http.Error(w, "Failed to create user", http.StatusInternalServerError)
					return
				}
			} else {
				// Send googleUser info to Discord webhook
				_ = sendNewUserToDiscord(r, user)
				glog.Info("User with externalID %v is not known", user.ExternalID)
				http.Error(w, "Unauthorised", http.StatusUnauthorized)
				return
			}
		} else {
			user = existingUser

			if user.Alias == "" {
				if alias, err := utils.GetUniqueAlias(func(alias string) (bool, error) {
					return repository.AliasUnique(r.Context(), alias)
				}); err != nil {
					utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err,
						"error fetching user profile")
					return
				} else {
					user.Alias = alias
				}
			}
			user.Name = externalUser.Name
			user.Avatar = externalUser.Avatar
			user.UpdatedAt = time.Now()
			if err := repository.Update(r.Context(), user); err != nil {
				utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err,
					"error updating user profile")
				return
			}
		}

		token, err := user_profile.CreateAuthToken(externalUser.ExternalIDs, user.ID.Hex(), externalUser.Name, externalUser.Avatar)
		if err != nil {
			_ = glog.Error("Token creation failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set secure cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   24 * 60 * 60, // 24 hours
		})

		if err := json.NewEncoder(w).Encode(userResponse{
			IDs:     user.ExternalID,
			Name:    user.Name,
			Picture: user.Avatar,
			Alias:   user.Alias,
		}); err != nil {
			_ = glog.Error("serialising error %v", err)
			http.Error(w, "serialising error", http.StatusInternalServerError)
		}
	}
}

type userResponse struct {
	IDs     []string `json:"emails"`
	Name    string   `json:"name"`
	Picture string   `json:"picture"`
	Alias   string   `json:"alias"`
}

func LogoutHandler(_ repositories.UserRepository, provider ExternalAuthProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = provider.LogoutHandler(w, r)

		// Clear the auth cookie
		clearCookies(w)

		w.WriteHeader(http.StatusOK)
	}
}

type ExternalUser struct {
	ExternalIDs []string
	Name        string
	Avatar      string
}

type ExternalAuthProvider interface {
	BeginUserAuthHandler(w http.ResponseWriter, r *http.Request)
	CompleteUserAuthHandler(w http.ResponseWriter, r *http.Request) (ExternalUser, error)
	LogoutHandler(w http.ResponseWriter, r *http.Request) error
}

func HandleBeginLoginFlow(provider ExternalAuthProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = provider.LogoutHandler(w, r)

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

		provider.BeginUserAuthHandler(w, r)
	}
}

func clearCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

var config = struct {
	GoogleClientID     string
	GoogleClientSecret string
}{
	GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
}

func Middleware(_ repositories.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				handleUnauthorized(w, r)
				return
			}

			profile, err := user_profile.ParseProfile(cookie.Value)

			if err == nil && len(profile.IDs) == 0 {
				ctx := context.WithValue(r.Context(), "user", profile)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				handleUnauthorized(w, r)
			}
		})
	}
}

func handleUnauthorized(w http.ResponseWriter, _ *http.Request) {
	clearCookies(w)

	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func isSuperAdmin(ids []string) bool {
	for _, admin := range superAdmins {
		for _, id := range ids {
			if admin == id {
				return true
			}
		}
	}
	return false
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
	createUserLink := fmt.Sprintf("%s/ui/admin/create-user?email=%s", domain, strings.Join(user.ExternalID, ",")) // domain defined at frontend/src/router/index.ts
	content := fmt.Sprintf("New user login: %s (%s). Click [%s] to create the user.", user.Name, strings.Join(user.ExternalID, ","), createUserLink)

	return utils.SendToDiscord(content)
}
