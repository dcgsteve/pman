package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/steve/pman/shared/auth"
	"github.com/steve/pman/shared/config"
	"github.com/steve/pman/shared/models"
)

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.ValidateLogin(req.Email, req.Password)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	expireDays := req.ExpireDays
	if expireDays <= 0 {
		envConfig := config.GetEnvConfig()
		expireDays = envConfig.DefaultExpireDays
	}

	token, err := auth.GenerateToken(user.Email, user.Role, expireDays)
	if err != nil {
		writeError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Store the token in the database for tracking/revocation
	expiresAt := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)
	if err := h.tokenService.StoreToken(token, user.Email, expiresAt); err != nil {
		writeError(w, "Failed to store token", http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	writeJSON(w, response)
}

func (h *Handlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, "Current password and new password are required", http.StatusBadRequest)
		return
	}

	_, err := h.userService.ValidateLogin(claims.Email, req.CurrentPassword)
	if err != nil {
		writeError(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	if err := h.userService.ChangePassword(claims.Email, req.NewPassword); err != nil {
		writeError(w, "Failed to change password", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"message": "Password changed successfully"})
}

func (h *Handlers) AdminChangePassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	var req struct {
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		writeError(w, "New password is required", http.StatusBadRequest)
		return
	}

	if err := h.userService.ChangePassword(email, req.NewPassword); err != nil {
		writeError(w, "Failed to change password", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"message": "Password changed successfully"})
}