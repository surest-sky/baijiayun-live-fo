package controller

import (
	"encoding/json"
	"net/http"
)

func response(w http.ResponseWriter, r *http.Request, data interface{}) interface{} {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
	return nil
}
