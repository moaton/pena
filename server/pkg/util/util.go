package util

import (
	"encoding/json"
	"net/http"
)

func ResponseOk(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(body)
}

func ResponseError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	var body = make(map[string]string, 0)
	body = map[string]string{
		"error": message,
	}

	_ = json.NewEncoder(w).Encode(body)
}
