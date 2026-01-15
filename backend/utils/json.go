package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, object interface{}, httpsStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpsStatus)
	if err := json.NewEncoder(w).Encode(object); err != nil {
		LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error serialising of object %v", object)
	}
}
