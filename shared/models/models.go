package models

import "time"

type User struct {
	ID          int       `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	Password    string    `json:"-" db:"password_hash"`
	Role        string    `json:"role" db:"role"`
	Groups      string    `json:"groups" db:"groups"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Password struct {
	ID          int       `json:"id" db:"id"`
	Path        string    `json:"path" db:"path"`
	Value       string    `json:"-" db:"encrypted_value"`
	GroupName   string    `json:"group_name" db:"group_name"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Group struct {
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ExpireDays int  `json:"expire_days,omitempty"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type PasswordRequest struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

type PasswordInfo struct {
	Path      string    `json:"path"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRequest struct {
	Email    string `json:"email"`
	Role     string `json:"role"`
	Groups   string `json:"groups"`
	Password string `json:"password,omitempty"`
}