package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/steve/pman/backend/database"
	"github.com/steve/pman/backend/services"
	"github.com/steve/pman/shared/auth"
)

type Handlers struct {
	userService     *services.UserService
	passwordService *services.PasswordService
	tokenService    *services.TokenService
}

func SetupRoutes(r *mux.Router, db *database.DB) {
	h := &Handlers{
		userService:     services.NewUserService(db),
		passwordService: services.NewPasswordService(db),
		tokenService:    services.NewTokenService(db),
	}

	r.HandleFunc("/auth/login", h.Login).Methods("POST")
	r.HandleFunc("/health", h.Health).Methods("GET")

	protected := r.PathPrefix("").Subrouter()
	protected.Use(auth.AuthMiddlewareWithTokenService(h.tokenService))

	protected.HandleFunc("/passwords", h.CreatePassword).Methods("POST")
	protected.HandleFunc("/passwords/{group}/{path:.*}/info", h.GetPasswordInfo).Methods("GET")
	protected.HandleFunc("/passwords/{group}/{path:.*}", h.GetPassword).Methods("GET")
	protected.HandleFunc("/passwords/{group}/{path:.*}", h.UpdatePassword).Methods("PUT")
	protected.HandleFunc("/passwords/{group}/{path:.*}", h.DeletePassword).Methods("DELETE")
	protected.HandleFunc("/passwords/{group}", h.ListPasswords).Methods("GET")

	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(auth.AdminRequired)

	admin.HandleFunc("/users", h.CreateUser).Methods("POST")
	admin.HandleFunc("/users", h.ListUsers).Methods("GET")
	admin.HandleFunc("/users/{email}", h.UpdateUser).Methods("PUT")
	admin.HandleFunc("/users/{email}", h.DeleteUser).Methods("DELETE")
	admin.HandleFunc("/users/{email}/enable", h.EnableUser).Methods("POST")
	admin.HandleFunc("/users/{email}/disable", h.DisableUser).Methods("POST")

	protected.HandleFunc("/auth/passwd", h.ChangePassword).Methods("POST")
	admin.HandleFunc("/users/{email}/passwd", h.AdminChangePassword).Methods("POST")
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"status": "healthy",
		"service": "pman-server",
	})
}