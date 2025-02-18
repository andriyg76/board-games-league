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
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Load super admins from environment variable
var superAdmins = strings.Split(os.Getenv("SUPERADMINS"), ",")

func init() {
	glog.Info("Registered superadmins: %v", superAdmins)
}

var store = sessions.NewCookieStore(config.SessionSecret)

func init() {
	gothic.Store = store
}

func GoogleCallbackHandler(repository repositories.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ensureGothInit(r)

		session, _ := store.Get(r, "auth-session")
		state := r.URL.Query().Get("state")
		storedState := session.Values["state"]
		delete(session.Values, "state")

		if storedState != nil && state != storedState.(string) {
			_ = glog.Error("Auth completion failed: State token mismatch")
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		googleUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			_ = glog.Error("Auth completion failed: %v", err)
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		token, err := createAuthToken(googleUser)
		if err != nil {
			_ = glog.Error("Token creation failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var user *models.User

		// Check if googleUser exists in the collection
		if existingUser, err := repository.FindByEmail(r.Context(), googleUser.Email); err != nil {
			_ = glog.Error("error fetching user profile: %v", err)
			http.Error(w, "error fetching user profile", http.StatusInternalServerError)
			return
		} else if existingUser == nil {
			if isSuperAdmin(googleUser.Email) {
				user = &models.User{
					ID:        primitive.ObjectID{},
					Email:     googleUser.Email,
					Name:      googleUser.Name,
					Picture:   googleUser.AvatarURL,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Alias:     "",
				}

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
				if err := repository.CreateUser(r.Context(), user); err != nil {
					_ = glog.Error("failed to create user", err)
					http.Error(w, "Failed to create user", http.StatusInternalServerError)
					return
				}
			} else {
				// Send googleUser info to Discord webhook
				_ = sendNewUserToDiscord(r, user)
				glog.Info("User %s is not known", user.Email)
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
			user.Name = googleUser.Name
			user.Picture = googleUser.AvatarURL
			user.UpdatedAt = time.Now()
			if err := repository.Update(r.Context(), user); err != nil {
				utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err,
					"error updating user profile")
				return
			}
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

		if err := json.NewEncoder(w).Encode(map[string]string{
			"email":   user.Email,
			"name":    user.Name,
			"picture": user.Picture,
			"alias":   user.Alias,
		}); err != nil {
			_ = glog.Error("serialising error %v", err)
			http.Error(w, "serialising error", http.StatusInternalServerError)
		}
	}
}

func LogoutHandler(_ repositories.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ensureGothInit(r)

		// Clear the auth cookie
		clearCookies(w)

		w.WriteHeader(http.StatusOK)
	}
}

var gothInitOnce sync.Once

func ensureGothInit(r *http.Request) {
	gothInitOnce.Do(func() {
		glog.Info("Late goth init...")
		hostName := utils.GetHostUrl(r)

		callbackUrl := hostName + "/ui/auth-callback" // defined at frontend/src/router/index.ts

		glog.Info("Google auth callback url: %v", callbackUrl)

		goth.UseProviders(
			google.New(
				config.GoogleClientID,
				config.GoogleClientSecret,
				callbackUrl,
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			),
		)
	})
}
func HandleLogin(_ repositories.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ensureGothInit(r)

		_ = gothic.Logout(w, r)

		query := r.URL.Query()
		if state := query.Get("state"); state != "" {
			session, _ := store.Get(r, "auth-session")
			session.Values["state"] = state
			if err := session.Save(r, w); err != nil {
				_ = glog.Error("serialising session error %v", err)
				http.Error(w, "serialising error", http.StatusInternalServerError)
				return
			}
		}

		gothic.BeginAuthHandler(w, r)
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
	SessionSecret      []byte
	JwtSecret          []byte
}{
	GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	SessionSecret:      []byte(os.Getenv("SESSION_SECRET")),
	JwtSecret:          []byte(os.Getenv("JWT_SECRET")),
}

func createAuthToken(user goth.User) (string, error) {
	claims := user_profile.Claims{
		Email:   user.Email,
		Name:    user.Name,
		Picture: user.AvatarURL,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JwtSecret)
}

func Middleware(_ repositories.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				handleUnauthorized(w, r)
				return
			}

			token, err := jwt.ParseWithClaims(cookie.Value, &user_profile.Claims{}, func(token *jwt.Token) (interface{}, error) {
				return config.JwtSecret, nil
			})

			if err != nil || !token.Valid {
				handleUnauthorized(w, r)
				return
			}

			if claims, ok := token.Claims.(*user_profile.Claims); ok {
				ctx := context.WithValue(r.Context(), "user", claims)
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

func isSuperAdmin(email string) bool {
	for _, admin := range superAdmins {
		if admin == email {
			return true
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
	createUserLink := fmt.Sprintf("%s/ui/admin/create-user?email=%s", domain, user.Email) // defined at frontend/src/router/index.ts
	content := fmt.Sprintf("New user login: %s (%s). Click [%s] to create the user.", user.Name, user.Email, createUserLink)

	return utils.SendToDiscord(content)
}
