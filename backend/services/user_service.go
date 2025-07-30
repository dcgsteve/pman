package services

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/steve/pman/backend/database"
	"github.com/steve/pman/shared/crypto"
	"github.com/steve/pman/shared/models"
	"github.com/steve/pman/shared/permissions"
)

type UserService struct {
	db *database.DB
}

func NewUserService(db *database.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, role, groups, enabled, created_at, updated_at
		FROM users WHERE email = ?
	`, email).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.Groups, &user.Enabled, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (s *UserService) CreateUser(email, role, groupsStr string) (string, error) {
	if _, err := permissions.ParseGroups(groupsStr); err != nil {
		return "", fmt.Errorf("invalid groups format: %w", err)
	}

	password := generateRandomPassword()
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return "", err
	}

	_, err = s.db.Exec(`
		INSERT INTO users (email, password_hash, role, groups, enabled) 
		VALUES (?, ?, ?, ?, true)
	`, email, hashedPassword, role, groupsStr)
	
	if err != nil {
		return "", err
	}

	return password, nil
}

func (s *UserService) UpdateUser(email, role, groupsStr string) error {
	if _, err := permissions.ParseGroups(groupsStr); err != nil {
		return fmt.Errorf("invalid groups format: %w", err)
	}

	_, err := s.db.Exec(`
		UPDATE users SET role = ?, groups = ?, updated_at = CURRENT_TIMESTAMP
		WHERE email = ?
	`, role, groupsStr, email)
	
	return err
}

func (s *UserService) DeleteUser(email string) error {
	if email == "admin@pman.system" {
		return fmt.Errorf("cannot delete default admin user")
	}

	_, err := s.db.Exec("DELETE FROM users WHERE email = ?", email)
	return err
}

func (s *UserService) ListUsers() ([]models.User, error) {
	rows, err := s.db.Query(`
		SELECT id, email, password_hash, role, groups, enabled, created_at, updated_at
		FROM users ORDER BY email
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.Groups, &user.Enabled, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (s *UserService) EnableUser(email string) error {
	_, err := s.db.Exec(`
		UPDATE users SET enabled = true, updated_at = CURRENT_TIMESTAMP
		WHERE email = ?
	`, email)
	return err
}

func (s *UserService) DisableUser(email string) error {
	if email == "admin@pman.system" {
		return fmt.Errorf("cannot disable default admin user")
	}

	_, err := s.db.Exec(`
		UPDATE users SET enabled = false, updated_at = CURRENT_TIMESTAMP
		WHERE email = ?
	`, email)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE tokens SET revoked = true WHERE user_email = ?
	`, email)
	
	return err
}

func (s *UserService) ChangePassword(email, newPassword string) error {
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP
		WHERE email = ?
	`, hashedPassword, email)
	
	return err
}

func (s *UserService) ValidateLogin(email, password string) (*models.User, error) {
	user, err := s.GetUserByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, err
	}

	if !user.Enabled {
		return nil, fmt.Errorf("user account is disabled")
	}

	if !crypto.CheckPasswordHash(password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

func generateRandomPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var password strings.Builder
	
	for i := 0; i < 16; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		password.WriteByte(chars[num.Int64()])
	}
	
	return password.String()
}