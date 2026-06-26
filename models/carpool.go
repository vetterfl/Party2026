package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// CarpoolPost is a single entry on the carpool message board. The author's
// display name is denormalised into AuthorName on read for convenient listing.
type CarpoolPost struct {
	ID         string    `db:"id"`
	GuestID    string    `db:"guest_id"`
	Kind       string    `db:"kind"` // "offer" or "request"
	Origin     string    `db:"origin"`
	TravelTime string    `db:"travel_time"`
	Seats      int       `db:"seats"`
	Note       string    `db:"note"`
	Contact    string    `db:"contact"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`

	AuthorName string `db:"author_name"`
}

// IsOffer reports whether the post is a ride offer (vs. a request).
func (p *CarpoolPost) IsOffer() bool { return p.Kind == "offer" }

func ValidCarpoolKind(kind string) bool {
	k := strings.ToLower(strings.TrimSpace(kind))
	return k == "offer" || k == "request"
}

type CarpoolStore struct{ db *sqlx.DB }

func NewCarpoolStore(db *sqlx.DB) *CarpoolStore { return &CarpoolStore{db} }

// List returns every carpool post with its author's display name, offers first
// then newest first.
func (s *CarpoolStore) List() ([]CarpoolPost, error) {
	var posts []CarpoolPost
	err := s.db.Select(&posts, `
		SELECT
			p.*,
			COALESCE(NULLIF(trim(g.nickname), ''), g.name) AS author_name
		FROM carpool_posts p
		JOIN guests g ON g.id = p.guest_id
		ORDER BY p.kind = 'offer' DESC, p.created_at DESC`)
	return posts, err
}

// ByGuest returns the posts authored by a single guest, newest first.
func (s *CarpoolStore) ByGuest(guestID string) ([]CarpoolPost, error) {
	var posts []CarpoolPost
	err := s.db.Select(&posts, `
		SELECT
			p.*,
			COALESCE(NULLIF(trim(g.nickname), ''), g.name) AS author_name
		FROM carpool_posts p
		JOIN guests g ON g.id = p.guest_id
		WHERE p.guest_id = ?
		ORDER BY p.created_at DESC`, guestID)
	return posts, err
}

func (s *CarpoolStore) Create(p *CarpoolPost) error {
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	_, err := s.db.Exec(`
		INSERT INTO carpool_posts
			(id, guest_id, kind, origin, travel_time, seats, note, contact, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.GuestID, p.Kind, p.Origin, p.TravelTime, p.Seats, p.Note, p.Contact,
		p.CreatedAt, p.UpdatedAt,
	)
	return err
}

// Delete removes a post, scoped to its owner so a guest can only delete posts
// they authored.
func (s *CarpoolStore) Delete(id, guestID string) error {
	_, err := s.db.Exec(`DELETE FROM carpool_posts WHERE id = ? AND guest_id = ?`, id, guestID)
	return err
}
