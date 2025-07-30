package services

import (
	"database/sql"
	"time"

	"github.com/steve/pman/backend/database"
	"github.com/steve/pman/shared/auth"
)

type TokenService struct {
	db *database.DB
}

func NewTokenService(db *database.DB) *TokenService {
	return &TokenService{db: db}
}

func (s *TokenService) StoreToken(token, userEmail string, expiresAt time.Time) error {
	tokenHash := auth.HashToken(token)
	
	_, err := s.db.Exec(`
		INSERT INTO tokens (token_hash, user_email, expires_at, revoked)
		VALUES (?, ?, ?, false)
	`, tokenHash, userEmail, expiresAt)
	
	return err
}

func (s *TokenService) IsTokenRevoked(token string) (bool, error) {
	tokenHash := auth.HashToken(token)
	
	var revoked bool
	err := s.db.QueryRow(`
		SELECT revoked FROM tokens 
		WHERE token_hash = ? AND expires_at > CURRENT_TIMESTAMP
	`, tokenHash).Scan(&revoked)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// Token not found in database, could be an old token from before tracking
			return false, nil
		}
		return false, err
	}
	
	return revoked, nil
}

func (s *TokenService) RevokeUserTokens(userEmail string) error {
	_, err := s.db.Exec(`
		UPDATE tokens 
		SET revoked = true 
		WHERE user_email = ? AND revoked = false AND expires_at > CURRENT_TIMESTAMP
	`, userEmail)
	
	return err
}

func (s *TokenService) CleanupExpiredTokens() error {
	_, err := s.db.Exec(`
		DELETE FROM tokens 
		WHERE expires_at <= CURRENT_TIMESTAMP
	`)
	
	return err
}