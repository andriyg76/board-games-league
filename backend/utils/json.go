package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(r *http.Request, w http.ResponseWriter, object interface{}, httpsStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpsStatus)
	if err := json.NewEncoder(w).Encode(object); err != nil {
		LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error serialising of object %v", object)
	}
}
