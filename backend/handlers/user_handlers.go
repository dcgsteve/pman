package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/steve/pman/shared/models"
)

func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Role == "" || req.Groups == "" {
		writeError(w, "Email, role, and groups are required", http.StatusBadRequest)
		return
	}

	password, err := h.userService.CreateUser(req.Email, req.Role, req.Groups)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{
		"message":  "User created successfully",
		"password": password,
	})
}

func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListUsers()
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{"users": users})
}

func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	var req models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" || req.Groups == "" {
		writeError(w, "Role and groups are required", http.StatusBadRequest)
		return
	}

	err := h.userService.UpdateUser(email, req.Role, req.Groups)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"message": "User updated successfully"})
}

func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	err := h.userService.DeleteUser(email)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"message": "User deleted successfully"})
}

func (h *Handlers) EnableUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	err := h.userService.EnableUser(email)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"message": "User enabled successfully"})
}

func (h *Handlers) DisableUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	err := h.userService.DisableUser(email)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Revoke all tokens for the disabled user
	if err := h.tokenService.RevokeUserTokens(email); err != nil {
		writeError(w, "Failed to revoke user tokens", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"message": "User disabled successfully"})
}