package userapi

import (
	"encoding/json"
	"fmt"
	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	log "github.com/andriyg76/glog"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) CheckAliasUniquenessHandler(w http.ResponseWriter, r *http.Request) {
	alias := r.URL.Query().Get("alias")
	if alias == "" {
		http.Error(w, "Alias is required", http.StatusBadRequest)
		return
	}

	unique, err := h.userRepository.AliasUnique(r.Context(), alias)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]bool{"isUnique": unique}); err != nil {
		log.Info("Error response serialising %v", err)
		http.Error(w, "Write result problem", http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*user_profile.UserProfile)
	if !ok || claims == nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, fmt.Errorf("claims are null or bad %v", r.Context().Value("user")), "server error")
		return
	}

	user, err := h.userRepository.FindByExternalId(r.Context(), claims.ExternalIDs)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching user profile")
		return
	}
	if user == nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusNotFound, fmt.Errorf("user profile not found"), "user profile not found")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	user.UpdatedAt = time.Now()

	if err := h.userRepository.Update(r.Context(), user); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error updating user profile")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if claims, err := user_profile.GetUserProfile(r); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusUnauthorized, err,
			"unauthorised")
		return
	} else {
		user, err := h.userRepository.FindByExternalId(r.Context(), claims.ExternalIDs)
		if err != nil {
			utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching user profile")
			return
		}

		if user == nil {
			utils.LogAndWriteHTTPError(r, w, http.StatusUnauthorized, nil, "user profile not found")
			return
		}

		if err := json.NewEncoder(w).Encode(user_profile.UserResponse{
			Code:        utils.IdToCode(user.ID),
			ExternalIDs: user.ExternalIDs,
			Name:        user.Name,
			Names:       user.Names,
			Avatar:      user.Avatar,
			Avatars:     user.Avatars,
			Alias:       user.Alias,
			Roles:       auth.GetUserRoles(user),
		}); err != nil {
			_ = log.Error("serialising error %v", err)
			http.Error(w, "serialising error", http.StatusInternalServerError)
		}
	}
}
func (h *Handler) AdminCreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExternalIDs []string `json:"external_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(req.ExternalIDs) == 0 {
		http.Error(w, "At least one external ExternalIDs is required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	if existingUser, err := h.userRepository.FindByExternalId(r.Context(), req.ExternalIDs); err != nil {
		_ = log.Error("error checking user %v", err)
		http.Error(w, "Error checking user", http.StatusConflict)
		return
	} else if existingUser != nil {
		log.Info("User %v already have one of external ids: %v assinged", existingUser, req.ExternalIDs)
		return
	}

	// Create new user
	newUser := &models.User{
		ExternalIDs: req.ExternalIDs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if alias, err := utils.GetUniqueAlias(func(alias string) (bool, error) {
		return h.userRepository.AliasUnique(r.Context(), alias)
	}); err != nil {
		_ = log.Error("failed to create user %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	} else {
		newUser.Alias = alias
	}

	if err := h.userRepository.Create(r.Context(), newUser); err != nil {
		_ = log.Error("failed to create user %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(w, "User created successfully")
}

type SessionInfo struct {
	ID             string            `json:"id"`
	IPAddress      string            `json:"ip_address"`
	UserAgent      string            `json:"user_agent"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	LastRotationAt time.Time         `json:"last_rotation_at"`
	ExpiresAt      time.Time         `json:"expires_at"`
	IsCurrent      bool              `json:"is_current"`
	GeoInfo        *models.GeoIPInfo `json:"geo_info,omitempty"`
}

func (h *Handler) GetUserSessionsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*user_profile.UserProfile)
	if !ok || claims == nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusUnauthorized, fmt.Errorf("unauthorized"), "unauthorized")
		return
	}

	user, err := h.userRepository.FindByExternalId(r.Context(), claims.ExternalIDs)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching user")
		return
	}
	if user == nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusNotFound, fmt.Errorf("user not found"), "user not found")
		return
	}

	// Get current rotate token from query param or Authorization header (optional)
	currentRotateToken := r.URL.Query().Get("current")
	if currentRotateToken == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				currentRotateToken = parts[1]
			}
		}
	}

	// Get all sessions for user
	sessions, err := h.sessionRepository.FindByUserID(r.Context(), user.ID)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching sessions")
		return
	}

	// Convert to SessionInfo with geo lookup
	sessionInfos := make([]SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		isCurrent := currentRotateToken != "" && session.RotateToken == currentRotateToken

		sessionInfo := SessionInfo{
			ID:             session.ID.Hex(),
			IPAddress:      session.IPAddress,
			UserAgent:      session.UserAgent,
			CreatedAt:      session.CreatedAt,
			UpdatedAt:      session.UpdatedAt,
			LastRotationAt: session.LastRotationAt,
			ExpiresAt:      session.ExpiresAt,
			IsCurrent:      isCurrent,
		}

		// Try to get geo info (non-blocking - if it fails, just omit it)
		if session.IPAddress != "" && h.geoIPService != nil {
			if geoInfo, err := h.geoIPService.GetGeoIPInfo(session.IPAddress); err == nil {
				sessionInfo.GeoInfo = geoInfo
			}
		}

		sessionInfos = append(sessionInfos, sessionInfo)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sessionInfos); err != nil {
		_ = log.Error("serialising error %v", err)
		http.Error(w, "serialising error", http.StatusInternalServerError)
	}
}

type Handler struct {
	userRepository    repositories.UserRepository
	sessionRepository repositories.SessionRepository
	geoIPService      services.GeoIPService
}

func NewHandler(userRepository repositories.UserRepository) *Handler {
	return &Handler{
		userRepository: userRepository,
	}
}

func NewHandlerWithServices(userRepository repositories.UserRepository, sessionRepository repositories.SessionRepository, geoIPService services.GeoIPService) *Handler {
	return &Handler{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		geoIPService:      geoIPService,
	}
}
