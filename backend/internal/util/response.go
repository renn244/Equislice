package util

import (
	"backend/internal/dto"
	"encoding/json"
	"net/http"
)

func WriteJson[T any](w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJson(w, status, dto.ErrorResponseDto{
		Message: message,
	})
}
