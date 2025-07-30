package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/steve/pman/shared/auth"
	"github.com/steve/pman/shared/models"
)

func (h *Handlers) CreatePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.PasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	groupName := r.URL.Query().Get("group")
	if groupName == "" {
		writeError(w, "Group parameter is required", http.StatusBadRequest)
		return
	}

	if req.Path == "" || req.Value == "" {
		writeError(w, "Path and value are required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserByEmail(claims.Email)
	if err != nil {
		writeError(w, "User not found", http.StatusInternalServerError)
		return
	}

	err = h.passwordService.CreatePassword(req.Path, req.Value, groupName, claims.Email, user.Groups)
	if err != nil {
		writeError(w, err.Error(), http.StatusForbidden)
		return
	}

	writeJSON(w, map[string]string{"message": "Password created successfully"})
}

func (h *Handlers) GetPassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupName := vars["group"]
	path := vars["path"]

	user, err := h.userService.GetUserByEmail(claims.Email)
	if err != nil {
		writeError(w, "User not found", http.StatusInternalServerError)
		return
	}

	value, err := h.passwordService.GetPassword(path, groupName, user.Groups)
	if err != nil {
		writeError(w, err.Error(), http.StatusForbidden)
		return
	}

	writeJSON(w, map[string]string{"value": value})
}

func (h *Handlers) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupName := vars["group"]
	path := vars["path"]

	var req models.PasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Value == "" {
		writeError(w, "Value is required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserByEmail(claims.Email)
	if err != nil {
		writeError(w, "User not found", http.StatusInternalServerError)
		return
	}

	err = h.passwordService.UpdatePassword(path, req.Value, groupName, claims.Email, user.Groups)
	if err != nil {
		writeError(w, err.Error(), http.StatusForbidden)
		return
	}

	writeJSON(w, map[string]string{"message": "Password updated successfully"})
}

func (h *Handlers) DeletePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupName := vars["group"]
	path := vars["path"]

	recursive := r.URL.Query().Get("recursive") == "true"

	user, err := h.userService.GetUserByEmail(claims.Email)
	if err != nil {
		writeError(w, "User not found", http.StatusInternalServerError)
		return
	}

	if recursive {
		count, err := h.passwordService.DeletePasswordRecursive(path, groupName, user.Groups)
		if err != nil {
			writeError(w, err.Error(), http.StatusForbidden)
			return
		}
		writeJSON(w, map[string]interface{}{"message": "Passwords deleted successfully", "count": count})
	} else {
		err = h.passwordService.DeletePassword(path, groupName, user.Groups)
		if err != nil {
			writeError(w, err.Error(), http.StatusForbidden)
			return
		}
		writeJSON(w, map[string]string{"message": "Password deleted successfully"})
	}
}

func (h *Handlers) ListPasswords(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupName := vars["group"]
	pathPrefix := r.URL.Query().Get("prefix")

	user, err := h.userService.GetUserByEmail(claims.Email)
	if err != nil {
		writeError(w, "User not found", http.StatusInternalServerError)
		return
	}

	paths, err := h.passwordService.ListPasswords(groupName, pathPrefix, user.Groups)
	if err != nil {
		writeError(w, err.Error(), http.StatusForbidden)
		return
	}

	writeJSON(w, map[string]interface{}{"paths": paths})
}

func (h *Handlers) GetPasswordInfo(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupName := vars["group"]
	path := vars["path"]

	user, err := h.userService.GetUserByEmail(claims.Email)
	if err != nil {
		writeError(w, "User not found", http.StatusInternalServerError)
		return
	}

	info, err := h.passwordService.GetPasswordInfo(path, groupName, user.Groups)
	if err != nil {
		writeError(w, err.Error(), http.StatusForbidden)
		return
	}

	writeJSON(w, info)
}
