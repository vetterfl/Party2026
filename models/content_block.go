package models

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type ContentBlock struct {
	Key       string    `db:"key"`
	Label     string    `db:"label"`
	BodyDE    string    `db:"body_de"`
	BodyEN    string    `db:"body_en"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ContentStore struct{ db *sqlx.DB }

func NewContentStore(db *sqlx.DB) *ContentStore { return &ContentStore{db} }

func (s *ContentStore) All() ([]ContentBlock, error) {
	var blocks []ContentBlock
	err := s.db.Select(&blocks, `SELECT * FROM content_blocks ORDER BY key`)
	return blocks, err
}

func (s *ContentStore) Get(key string) (ContentBlock, error) {
	var b ContentBlock
	err := s.db.Get(&b, `SELECT * FROM content_blocks WHERE key = ?`, key)
	return b, err
}

func (s *ContentStore) Save(key, bodyDE, bodyEN string) error {
	_, err := s.db.Exec(`
		UPDATE content_blocks SET body_de = ?, body_en = ?, updated_at = ? WHERE key = ?`,
		bodyDE, bodyEN, time.Now(), key,
	)
	return err
}

// Body returns the appropriate language body, falling back to DE.
func (b ContentBlock) Body(lang string) string {
	if lang == "en" && b.BodyEN != "" {
		return b.BodyEN
	}
	return b.BodyDE
}
