package main

import (
	"context"
	"encoding/json"
	log "github.com/andriyg76/glog"
	"github.com/golang-jwt/jwt"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"net/http"
	"os"
	"time"
)

func init() {
	config := loadConfig()
	goth.UseProviders(
		google.New(
			config.GoogleClientID,
			config.GoogleClientSecret,
			config.CallbackURL,
		),
	)
}

func googleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Error("Auth completion failed: %v", err)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	token, err := createAuthToken(user)
	if err != nil {
		log.Error("Token creation failed: %v", err)
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

	json.NewEncoder(w).Encode(map[string]string{
		"email":   user.Email,
		"name":    user.Name,
		"picture": user.AvatarURL,
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the auth cookie
	clearCookies(w)

	w.WriteHeader(http.StatusOK)
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

type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	CallbackURL        string
	// Add other config fields as needed
}

func loadConfig() Config {
	return Config{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		CallbackURL:        os.Getenv("AUTH_CALLBACK_URL"),
	}
}

type Claims struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	jwt.StandardClaims
}

func createAuthToken(user goth.User) (string, error) {
	claims := Claims{
		Email:   user.Email,
		Name:    user.Name,
		Picture: user.AvatarURL,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			handleUnauthorized(w, r)
			return
		}

		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			handleUnauthorized(w, r)
			return
		}

		if claims, ok := token.Claims.(*Claims); ok {
			ctx := context.WithValue(r.Context(), "user", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			handleUnauthorized(w, r)
		}
	})
}

func handleUnauthorized(w http.ResponseWriter, r *http.Request) {
	clearCookies(w)

	// Create a new request with the provider specified
	authReq := r.WithContext(r.Context())
	q := authReq.URL.Query()
	q.Add("provider", "google")
	authReq.URL.RawQuery = q.Encode()

	url, err := gothic.GetAuthURL(w, authReq)
	if err != nil {
		log.Error("Failed to get auth URL: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Auth-URL", url)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	if claims, ok := r.Context().Value("user").(*Claims); !ok {
		log.Error("Failed to parse user claims %v", r.Context().Value("user"))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else {
		json.NewEncoder(w).Encode(map[string]string{
			"email":   claims.Email,
			"name":    claims.Name,
			"picture": claims.Picture,
		})
	}
}

func googleAuthHandler(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}
