package models

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Guest struct {
	ID          string     `db:"id"`
	Code        string     `db:"code"`
	Name        string     `db:"name"`
	Status      string     `db:"status"`
	Email       *string    `db:"email"`
	PlusOne     bool       `db:"plus_one"`
	PlusOneName *string    `db:"plus_one_name"`
	Children    int        `db:"children"`
	Song        *string    `db:"song"`
	Comment     *string    `db:"comment"`
	Newsletter  bool       `db:"newsletter"`
	RSVPAt      *time.Time `db:"rsvp_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type GuestStore struct{ db *sqlx.DB }

func NewGuestStore(db *sqlx.DB) *GuestStore { return &GuestStore{db} }

func (s *GuestStore) FindByCode(code string) (*Guest, error) {
	g := &Guest{}
	err := s.db.Get(g, `SELECT * FROM guests WHERE code = ?`, strings.ToUpper(code))
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *GuestStore) FindByID(id string) (*Guest, error) {
	g := &Guest{}
	err := s.db.Get(g, `SELECT * FROM guests WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *GuestStore) All() ([]Guest, error) {
	var gs []Guest
	err := s.db.Select(&gs, `SELECT * FROM guests ORDER BY name`)
	return gs, err
}

func (s *GuestStore) ByStatus(status string) ([]Guest, error) {
	var gs []Guest
	err := s.db.Select(&gs, `SELECT * FROM guests WHERE status = ? ORDER BY name`, status)
	return gs, err
}

func (s *GuestStore) Create(name string) (*Guest, error) {
	g := &Guest{
		ID:        uuid.New().String(),
		Code:      generateCode(6),
		Name:      name,
		Status:    "invited",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := s.db.Exec(`
		INSERT INTO guests (id, code, name, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		g.ID, g.Code, g.Name, g.Status, g.CreatedAt, g.UpdatedAt,
	)
	return g, err
}

func (s *GuestStore) Update(g *Guest) error {
	g.UpdatedAt = time.Now()
	_, err := s.db.Exec(`
		UPDATE guests SET
			name=?, status=?, email=?, plus_one=?, plus_one_name=?,
			children=?, song=?, comment=?, newsletter=?, rsvp_at=?, updated_at=?
		WHERE id=?`,
		g.Name, g.Status, g.Email, g.PlusOne, g.PlusOneName,
		g.Children, g.Song, g.Comment, g.Newsletter, g.RSVPAt, g.UpdatedAt,
		g.ID,
	)
	return err
}

func (s *GuestStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM guests WHERE id = ?`, id)
	return err
}

type Stats struct {
	Accepted   int
	Declined   int
	Pending    int
	NoResponse int
	TotalHeads int // guests + plus ones + children
}

func (s *GuestStore) Stats() (Stats, error) {
	rows, err := s.db.Queryx(`
		SELECT status, COUNT(*) as n,
		       SUM(plus_one) as po, SUM(children) as ch
		FROM guests GROUP BY status`)
	if err != nil {
		return Stats{}, err
	}
	defer rows.Close()

	var st Stats
	for rows.Next() {
		var status string
		var n, po, ch int
		if err := rows.Scan(&status, &n, &po, &ch); err != nil {
			return Stats{}, err
		}
		st.TotalHeads += n + po + ch
		switch status {
		case "accepted":
			st.Accepted = n
		case "declined":
			st.Declined = n
		case "invited", "tentative":
			st.Pending += n
		case "no_response":
			st.NoResponse = n
		}
	}
	return st, nil
}

func (s *GuestStore) NewsletterRecipients() ([]Guest, error) {
	var gs []Guest
	err := s.db.Select(&gs,
		`SELECT * FROM guests WHERE newsletter = 1 AND email IS NOT NULL AND email != ''`)
	return gs, err
}

func (s *GuestStore) UnsubscribeByID(id string) error {
	_, err := s.db.Exec(
		`UPDATE guests SET newsletter = 0, updated_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

func generateCode(n int) string {
	b := make([]byte, n*2)
	_, _ = rand.Read(b)
	enc := base32.StdEncoding.EncodeToString(b)
	enc = strings.NewReplacer("0", "", "O", "1", "=", "").Replace(enc)
	if len(enc) < n {
		return generateCode(n)
	}
	return enc[:n]
}
