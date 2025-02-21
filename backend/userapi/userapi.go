package userapi

import (
	"encoding/json"
	"fmt"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	log "github.com/andriyg76/glog"
	"net/http"
	"time"
)

func CheckAliasUniquenessHandler(userRepository repositories.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		alias := r.URL.Query().Get("alias")
		if alias == "" {
			http.Error(w, "Alias is required", http.StatusBadRequest)
			return
		}

		unique, err := userRepository.AliasUnique(r.Context(), alias)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(map[string]bool{"isUnique": unique}); err != nil {
			log.Info("Error response serialising %v", err)
			http.Error(w, "Write result problem", http.StatusInternalServerError)
		}
	}
}

func UpdateUser(userRepository repositories.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value("user").(*user_profile.UserProfile)
		if !ok || claims == nil {
			utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, fmt.Errorf("claims are null or bad %v", r.Context().Value("user")), "server error")
			return
		}

		user, err := userRepository.FindByExternalId(r.Context(), claims.IDs)
		if err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error fetching user profile")
			return
		}
		if user == nil {
			utils.LogAndWriteHTTPError(w, http.StatusNotFound, fmt.Errorf("user profile not found"), "user profile not found")
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		user.UpdatedAt = time.Now()

		if err := userRepository.Update(r.Context(), user); err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error updating user profile")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type userResponse struct {
	IDs     []string
	Name    string
	Picture string
	Alias   string
}

func GetUserHandler(userRepository repositories.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if claims, err := user_profile.GetUserProfile(r); err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusUnauthorized, err,
				"unauthorised")
			return
		} else {
			user, err := userRepository.FindByExternalId(r.Context(), claims.IDs)
			if err != nil {
				utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error fetching user profile")
			}

			if err := json.NewEncoder(w).Encode(userResponse{
				IDs:     user.ExternalID,
				Name:    user.Name,
				Picture: user.Avatar,
				Alias:   user.Alias,
			}); err != nil {
				_ = log.Error("serialising error %v", err)
				http.Error(w, "serialising error", http.StatusInternalServerError)
			}
		}
	}
}

func AdminCreateUserHandler(userRepository repositories.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ExternalIDs []string `json:"external_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if len(req.ExternalIDs) == 0 {
			http.Error(w, "At leas one external IDs is required", http.StatusBadRequest)
			return
		}

		// Check if user already exists
		if existingUser, err := userRepository.FindByExternalId(r.Context(), req.ExternalIDs); err != nil {
			_ = log.Error("error checking user %v", err)
			http.Error(w, "Error checking user", http.StatusConflict)
			return
		} else if existingUser != nil {
			log.Info("User %v already have one of external ids: %v assinged", existingUser, req.ExternalIDs)
			return
		}

		// Create new user
		newUser := &models.User{
			ExternalID: req.ExternalIDs,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if alias, err := utils.GetUniqueAlias(func(alias string) (bool, error) {
			return userRepository.AliasUnique(r.Context(), alias)
		}); err != nil {
			_ = log.Error("failed to create user %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		} else {
			newUser.Alias = alias
		}

		if err := userRepository.Create(r.Context(), newUser); err != nil {
			_ = log.Error("failed to create user %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, "User created successfully")
	}

}
