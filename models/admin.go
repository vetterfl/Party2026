package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Admin struct {
	ID           string    `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type AdminStore struct{ db *sqlx.DB }

func NewAdminStore(db *sqlx.DB) *AdminStore { return &AdminStore{db} }

func (s *AdminStore) FindByUsername(username string) (*Admin, error) {
	a := &Admin{}
	err := s.db.Get(a,
		`SELECT * FROM admins WHERE username = ? COLLATE NOCASE`,
		strings.TrimSpace(username),
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AdminStore) FindByID(id string) (*Admin, error) {
	a := &Admin{}
	err := s.db.Get(a, `SELECT * FROM admins WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AdminStore) All() ([]Admin, error) {
	var admins []Admin
	err := s.db.Select(&admins, `SELECT id, username, created_at, updated_at FROM admins ORDER BY username`)
	return admins, err
}

func (s *AdminStore) Count() (int, error) {
	var n int
	err := s.db.Get(&n, `SELECT COUNT(*) FROM admins`)
	return n, err
}

func (s *AdminStore) Create(username, passwordHash string) (*Admin, error) {
	now := time.Now()
	a := &Admin{
		ID:           uuid.New().String(),
		Username:     strings.TrimSpace(username),
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err := s.db.Exec(`
		INSERT INTO admins (id, username, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		a.ID, a.Username, a.PasswordHash, a.CreatedAt, a.UpdatedAt,
	)
	return a, err
}

func (s *AdminStore) UpdatePassword(id, passwordHash string) error {
	_, err := s.db.Exec(
		`UPDATE admins SET password_hash = ?, updated_at = ? WHERE id = ?`,
		passwordHash, time.Now(), id,
	)
	return err
}

func (s *AdminStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM admins WHERE id = ?`, id)
	return err
}
