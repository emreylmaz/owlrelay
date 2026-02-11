// Package store provides data access operations
package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/emreylmaz/owlrelay/relay/internal/database"
	"github.com/emreylmaz/owlrelay/relay/internal/models"
)

// TokenStore handles token-related database operations
type TokenStore struct {
	db *database.DB
}

// NewTokenStore creates a new TokenStore
func NewTokenStore(db *database.DB) *TokenStore {
	return &TokenStore{db: db}
}

// GenerateToken creates a new random token
func GenerateToken() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "owl_" + hex.EncodeToString(bytes), nil
}

// HashToken creates a SHA-256 hash of the token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Create stores a new token in the database
func (s *TokenStore) Create(name string, rateLimit int) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	hash := HashToken(token)

	_, err = s.db.Exec(
		"INSERT INTO tokens (hash, name, rate_limit, created_at) VALUES (?, ?, ?, ?)",
		hash, name, rateLimit, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert token: %w", err)
	}

	return token, nil
}

// Validate checks if a token is valid and returns its metadata
func (s *TokenStore) Validate(token string) (*models.Token, error) {
	hash := HashToken(token)

	var t models.Token
	var createdAt, lastUsedAt, revokedAt sql.NullString

	err := s.db.QueryRow(
		"SELECT id, hash, name, rate_limit, created_at, last_used_at, revoked_at FROM tokens WHERE hash = ?",
		hash,
	).Scan(&t.ID, &t.Hash, &t.Name, &t.RateLimit, &createdAt, &lastUsedAt, &revokedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Token not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query token: %w", err)
	}

	// Check if revoked
	if revokedAt.Valid {
		return nil, nil // Token is revoked
	}

	// Parse timestamps
	if createdAt.Valid {
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}
	if lastUsedAt.Valid {
		parsed, _ := time.Parse(time.RFC3339, lastUsedAt.String)
		t.LastUsedAt = &parsed
	}

	// Update last used
	go func() {
		_, _ = s.db.Exec(
			"UPDATE tokens SET last_used_at = ? WHERE id = ?",
			time.Now().UTC().Format(time.RFC3339), t.ID,
		)
	}()

	return &t, nil
}

// List returns all tokens (without hashes)
func (s *TokenStore) List() ([]*models.Token, error) {
	rows, err := s.db.Query(
		"SELECT id, name, rate_limit, created_at, last_used_at, revoked_at FROM tokens ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*models.Token
	for rows.Next() {
		var t models.Token
		var createdAt, lastUsedAt, revokedAt sql.NullString

		if err := rows.Scan(&t.ID, &t.Name, &t.RateLimit, &createdAt, &lastUsedAt, &revokedAt); err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}

		if createdAt.Valid {
			t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
		}
		if lastUsedAt.Valid {
			parsed, _ := time.Parse(time.RFC3339, lastUsedAt.String)
			t.LastUsedAt = &parsed
		}
		if revokedAt.Valid {
			parsed, _ := time.Parse(time.RFC3339, revokedAt.String)
			t.RevokedAt = &parsed
		}

		tokens = append(tokens, &t)
	}

	return tokens, rows.Err()
}

// Revoke marks a token as revoked
func (s *TokenStore) Revoke(id int64) error {
	result, err := s.db.Exec(
		"UPDATE tokens SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL",
		time.Now().UTC().Format(time.RFC3339), id,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("token not found or already revoked")
	}

	return nil
}
