package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/steve/pman/backend/database"
	"github.com/steve/pman/shared/crypto"
	"github.com/steve/pman/shared/models"
	"github.com/steve/pman/shared/permissions"
)

type PasswordService struct {
	db *database.DB
}

func NewPasswordService(db *database.DB) *PasswordService {
	return &PasswordService{db: db}
}

func (s *PasswordService) CreatePassword(path, value, groupName, userEmail string, userGroups string) error {
	if !permissions.HasGroupAccess(userGroups, groupName, true) {
		return fmt.Errorf("insufficient permissions to write to group '%s'", groupName)
	}

	encryptedValue, err := crypto.Encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO passwords (path, encrypted_value, group_name, created_by, updated_by, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, path, encryptedValue, groupName, userEmail, userEmail)

	return err
}

func (s *PasswordService) GetPassword(path, groupName string, userGroups string) (string, error) {
	if !permissions.HasGroupAccess(userGroups, groupName, false) {
		return "", fmt.Errorf("insufficient permissions to read from group '%s'", groupName)
	}

	var encryptedValue string
	err := s.db.QueryRow(`
		SELECT encrypted_value FROM passwords 
		WHERE path = ? AND group_name = ?
	`, path, groupName).Scan(&encryptedValue)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("password not found")
		}
		return "", err
	}

	value, err := crypto.Decrypt(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %w", err)
	}

	return value, nil
}

func (s *PasswordService) GetPasswordInfo(path, groupName string, userGroups string) (*models.PasswordInfo, error) {
	if !permissions.HasGroupAccess(userGroups, groupName, false) {
		return nil, fmt.Errorf("insufficient permissions to read from group '%s'", groupName)
	}

	info := &models.PasswordInfo{}
	err := s.db.QueryRow(`
		SELECT path, created_by, updated_by, created_at, updated_at
		FROM passwords WHERE path = ? AND group_name = ?
	`, path, groupName).Scan(&info.Path, &info.CreatedBy, &info.UpdatedBy, &info.CreatedAt, &info.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("password not found")
		}
		return nil, err
	}

	return info, nil
}

func (s *PasswordService) UpdatePassword(path, value, groupName, userEmail string, userGroups string) error {
	if !permissions.HasGroupAccess(userGroups, groupName, true) {
		return fmt.Errorf("insufficient permissions to write to group '%s'", groupName)
	}

	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM passwords WHERE path = ? AND group_name = ?)
	`, path, groupName).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("password not found")
	}

	encryptedValue, err := crypto.Encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE passwords 
		SET encrypted_value = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE path = ? AND group_name = ?
	`, encryptedValue, userEmail, path, groupName)

	return err
}

func (s *PasswordService) DeletePassword(path, groupName string, userGroups string) error {
	if !permissions.HasGroupAccess(userGroups, groupName, true) {
		return fmt.Errorf("insufficient permissions to write to group '%s'", groupName)
	}

	result, err := s.db.Exec(`
		DELETE FROM passwords WHERE path = ? AND group_name = ?
	`, path, groupName)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("password not found")
	}

	// Clean up empty parent folders
	s.cleanupEmptyFolders(path, groupName)

	return nil
}

func (s *PasswordService) ListPasswords(groupName string, pathPrefix string, userGroups string) ([]string, error) {
	if !permissions.HasGroupAccess(userGroups, groupName, false) {
		return nil, fmt.Errorf("insufficient permissions to read from group '%s'", groupName)
	}

	query := `SELECT path FROM passwords WHERE group_name = ?`
	args := []interface{}{groupName}

	if pathPrefix != "" {
		query += ` AND path LIKE ?`
		args = append(args, pathPrefix+"%")
	}

	query += ` ORDER BY path`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}

	return paths, nil
}

func (s *PasswordService) DeletePasswordRecursive(pathPrefix, groupName string, userGroups string) (int, error) {
	if !permissions.HasGroupAccess(userGroups, groupName, true) {
		return 0, fmt.Errorf("insufficient permissions to write to group '%s'", groupName)
	}

	result, err := s.db.Exec(`
		DELETE FROM passwords WHERE group_name = ? AND path LIKE ?
	`, groupName, pathPrefix+"%")
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	// Clean up empty parent folders after recursive deletion
	if rowsAffected > 0 {
		s.cleanupEmptyFolders(pathPrefix, groupName)
	}

	return int(rowsAffected), nil
}

// cleanupEmptyFolders recursively removes empty parent folders after password deletion
func (s *PasswordService) cleanupEmptyFolders(deletedPath, groupName string) {
	// Get parent folder path
	pathParts := strings.Split(deletedPath, "/")
	if len(pathParts) <= 1 {
		return // No parent folder to clean up
	}
	
	// Work backwards through parent folders
	for i := len(pathParts) - 1; i > 0; i-- {
		parentPath := strings.Join(pathParts[:i], "/")
		if parentPath == "" {
			break // Reached top level
		}
		
		// Check if this folder has any passwords
		hasPasswords := s.folderHasPasswords(parentPath, groupName)
		if hasPasswords {
			break // Stop cleanup, folder still contains passwords
		}
		
		// Folder is empty, continue to check parent
		// Note: We don't actually delete folder entries since we're using a flat structure
		// The tree display will automatically hide empty folders
	}
}

// folderHasPasswords checks if a folder path contains any passwords
func (s *PasswordService) folderHasPasswords(folderPath, groupName string) bool {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM passwords 
		WHERE group_name = ? AND path LIKE ?
	`, groupName, folderPath+"/%").Scan(&count)
	
	if err != nil {
		return true // Assume folder has passwords on error to be safe
	}
	
	return count > 0
}