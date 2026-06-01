package models

import (
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Guest struct {
	ID          string     `db:"id"`
	Code        string     `db:"code"`
	Name         string  `db:"name"`
	Nickname     string  `db:"nickname"`
	InternalNote *string `db:"internal_note"`
	Status       string  `db:"status"`
	Email       *string    `db:"email"`
	PhoneE164   *string    `db:"phone_e164"`
	PlusOne     bool       `db:"plus_one"`
	PlusOneName *string    `db:"plus_one_name"`
	Children    int        `db:"children"`
	Song        *string    `db:"song"`
	Comment           *string    `db:"comment"`
	Newsletter        bool       `db:"newsletter"`
	RSVPAt            *time.Time `db:"rsvp_at"`
	LoginCount        int        `db:"login_count"`
	ViewCount         int        `db:"view_count"`
	InteractionCount  int        `db:"interaction_count"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// DisplayName returns the name shown to the guest (nickname, falling back to name).
func (g *Guest) DisplayName() string {
	if n := strings.TrimSpace(g.Nickname); n != "" {
		return n
	}
	return g.Name
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

// ByNoResponse returns guests who have not RSVP'd accepted, declined, or tentative.
func (s *GuestStore) ByNoResponse() ([]Guest, error) {
	var gs []Guest
	err := s.db.Select(&gs,
		`SELECT * FROM guests WHERE status NOT IN ('accepted', 'declined', 'tentative') ORDER BY name`)
	return gs, err
}

var validGuestStatuses = map[string]struct{}{
	"added":     {},
	"invited":   {},
	"accepted":  {},
	"declined":  {},
	"tentative": {},
}

func ValidGuestStatus(status string) bool {
	_, ok := validGuestStatuses[strings.ToLower(strings.TrimSpace(status))]
	return ok
}

func (s *GuestStore) Create(name string) (*Guest, error) {
	g := &Guest{
		ID:        uuid.New().String(),
		Code:      generateCode(6),
		Name:      name,
		Nickname:  name,
		Status:    "added",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := s.db.Exec(`
		INSERT INTO guests (id, code, name, nickname, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		g.ID, g.Code, g.Name, g.Nickname, g.Status, g.CreatedAt, g.UpdatedAt,
	)
	return g, err
}

func (s *GuestStore) Update(g *Guest) error {
	g.UpdatedAt = time.Now()
	_, err := s.db.Exec(`
		UPDATE guests SET
			name=?, nickname=?, internal_note=?, status=?, email=?, phone_e164=?,
			plus_one=?, plus_one_name=?, children=?, song=?, comment=?, newsletter=?,
			rsvp_at=?, updated_at=?
		WHERE id=?`,
		g.Name, g.Nickname, g.InternalNote, g.Status, g.Email, g.PhoneE164,
		g.PlusOne, g.PlusOneName, g.Children, g.Song, g.Comment, g.Newsletter,
		g.RSVPAt, g.UpdatedAt, g.ID,
	)
	return err
}

func (s *GuestStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM guests WHERE id = ?`, id)
	return err
}

type Stats struct {
	Added      int
	Invited    int
	Accepted   int
	Declined   int
	Tentative  int
	NoResponse int // computed: guests without an RSVP (added + invited)
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
		case "added":
			st.Added = n
		case "invited":
			st.Invited = n
		case "accepted":
			st.Accepted = n
		case "declined":
			st.Declined = n
		case "tentative":
			st.Tentative = n
		}
	}
	st.NoResponse = st.Added + st.Invited
	return st, nil
}

func (s *GuestStore) NewsletterRecipients() ([]Guest, error) {
	var gs []Guest
	err := s.db.Select(&gs,
		`SELECT * FROM guests WHERE newsletter = 1 AND email IS NOT NULL AND email != ''`)
	return gs, err
}

func (s *GuestStore) WithComments() ([]Guest, error) {
	var gs []Guest
	err := s.db.Select(&gs, `
		SELECT * FROM guests
		WHERE comment IS NOT NULL AND trim(comment) != ''
		ORDER BY COALESCE(rsvp_at, updated_at) DESC`)
	return gs, err
}

func (s *GuestStore) IncrementLoginCount(id string) error {
	_, err := s.db.Exec(
		`UPDATE guests SET login_count = login_count + 1, updated_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

func (s *GuestStore) IncrementViewCount(id string) error {
	_, err := s.db.Exec(
		`UPDATE guests SET view_count = view_count + 1, updated_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

func (s *GuestStore) IncrementInteractionCount(id string) error {
	_, err := s.db.Exec(
		`UPDATE guests SET interaction_count = interaction_count + 1, updated_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

func (s *GuestStore) UnsubscribeByID(id string) error {
	_, err := s.db.Exec(
		`UPDATE guests SET newsletter = 0, updated_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

func generateCode(n int) string {
	const consonants = "BCDFGHJKLMNPQRSTVWXYZ"
	const vowels = "AEIOU"

	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		chars := consonants
		if i%2 == 1 {
			chars = vowels
		}
		b.WriteByte(randomChar(chars))
	}
	return b.String()
}

func randomChar(chars string) byte {
	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	if err != nil {
		panic(err)
	}
	return chars[idx.Int64()]
}
