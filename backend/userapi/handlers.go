package userapi

import (
	"encoding/json"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	log "github.com/andriyg76/glog"
	"net/http"
)

func NewCheckAliasUniqueness(userRepository *repositories.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		alias := r.URL.Query().Get("alias")
		if alias == "" {
			http.Error(w, "Alias is required", http.StatusBadRequest)
			return
		}

		exists, err := userRepository.AliasExists(r.Context(), alias)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(map[string]bool{"isUnique": !exists}); err != nil {
			log.Info("Error response serialising %v", err)
			http.Error(w, "Write result problem", http.StatusInternalServerError)
		}
	}
}

func UpdateUser(userRepository *repositories.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if err := userRepository.Update(r.Context(), &user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
