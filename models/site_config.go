package models

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type ConfigStore struct{ db *sqlx.DB }

func NewConfigStore(db *sqlx.DB) *ConfigStore { return &ConfigStore{db} }

func (s *ConfigStore) Get(key string) (string, error) {
	var val string
	err := s.db.Get(&val, `SELECT value FROM site_config WHERE key = ?`, key)
	return val, err
}

func (s *ConfigStore) GetDefault(key, def string) string {
	v, err := s.Get(key)
	if err != nil || v == "" {
		return def
	}
	return v
}

func (s *ConfigStore) Set(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO site_config (key, value, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, time.Now(),
	)
	return err
}

func (s *ConfigStore) All() (map[string]string, error) {
	rows, err := s.db.Queryx(`SELECT key, value FROM site_config`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, nil
}
