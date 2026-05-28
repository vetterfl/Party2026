# Party 2026 — Specification

**Status:** Final Draft · April 2026  
**Party date:** 01.08.2026 · 16:00 open end

---

## 1. Overview

Summer party (pool party). ~30–50 guests, ages 30–50 + kids 4–10. Website = mysterious spell-login + personal invitation + RSVP. Admin manages guestlist + editable content (CMS). Every guest has a unique personal code ("Zauberspruch").

---

## 2. Tech Stack


| Layer          | Choice                     | Notes                                              |
| -------------- | -------------------------- | -------------------------------------------------- |
| **Framework**  | Echo v4                    | Actively maintained, groups+middleware, clean API  |
| **Database**   | SQLite                     | File-based, fine for this scale                    |
| **DB layer**   | sqlx                       | Minimal SQL extension, already a transitive dep    |
| **Migrations** | goose                      | Plain `.sql` files, no DSL magic                   |
| **Templates**  | `html/template` (stdlib)   | No deps, safe, fast                                |
| **Sessions**   | gorilla/sessions           | Cookie-based                                       |
| **i18n**       | simple map in `locales.go` | DE default, EN via cookie — no external dep needed |
| **QR codes**   | go-qrcode                  | Server-side QR PNG generation                      |
| **Markdown**   | goldmark                   | CMS content rendering                              |
| **Email**      | net/smtp (stdlib)          | Self-hosted SMTP                                   |


**Why not Buffalo:** unmaintained since ~2023, heavy dep tree, Plush non-standard.  
**Why Echo over Chi:** built-in group routing (clean `/admin` group with middleware), template interface, session middleware.  
**Why sqlx over GORM:** 3 tables, no magic needed, already in deps.

---

## 3. Database Schema

### `guests`

```sql
CREATE TABLE guests (
    id            TEXT PRIMARY KEY,       -- UUID
    code          TEXT NOT NULL UNIQUE,   -- the "spell" / login credential
    name          TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'invited',
                                          -- invited|accepted|declined|tentative|no_response
    email         TEXT,
    plus_one      INTEGER NOT NULL DEFAULT 0,
    plus_one_name TEXT,
    children      INTEGER NOT NULL DEFAULT 0,
    song          TEXT,
    comment       TEXT,
    newsletter    INTEGER NOT NULL DEFAULT 0,
    rsvp_at       DATETIME,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL
);
```

### `content_blocks`

```sql
CREATE TABLE content_blocks (
    key        TEXT PRIMARY KEY,
    label      TEXT NOT NULL,
    body_de    TEXT NOT NULL DEFAULT '',
    body_en    TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL
);
```

### `site_config`

```sql
CREATE TABLE site_config (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at DATETIME NOT NULL
);
```

**Default `site_config` seed:**


| key                | default value  |
| ------------------ | -------------- |
| `party_date`       | `2026-08-01`   |
| `party_time_start` | `16:00`        |
| `party_name_de`    | `Summer Party` |
| `party_name_en`    | `Summer Party` |
| `rsvp_deadline`    | `2026-07-15`   |
| `charity_name`     | ``             |
| `charity_url`      | ``             |
| `smtp_from_name`   | `Florian`      |


---

## 4. Guest Flow — "The Spell"

Landing page is **not** a public invitation. It's a mysterious login page. No party name visible.

```
GET  /   →  Spell input page (dark, atmospheric, single input field)
            URL may include ?spell=XXXXX (from QR scan) → prefills input
POST /login → lookup code in guests table
            Found    → set session (guest_id) → redirect /me
            Not found → error: "Das war kein Zauber." / "That was no spell."
```

After login, all guest routes are session-protected. Code never appears in the URL again.

```
/me            → Personal invitation page
/me/rsvp       → RSVP form (GET prefilled + POST)
/me/confirmed  → Thank-you / confirmation
/logout        → Clear session → redirect /
```

---

## 5. URL Structure

```
GET  /                       → Spell login page
POST /login                  → Authenticate → session → /me
GET  /logout                 → Clear session
GET  /unsubscribe?token=...  → Newsletter opt-out (HMAC token)

GET  /me                     → Personal invite (session required)
GET  /me/rsvp                → RSVP form prefilled (session required)
POST /me/rsvp                → Submit RSVP
GET  /me/confirmed           → Confirmation page

GET  /admin/login            → Admin login form
POST /admin/login            → Admin auth → session → /admin
GET  /admin                  → Dashboard
GET  /admin/guests           → Guestlist
POST /admin/guests           → Create guest
GET  /admin/guests/:id/edit  → Edit guest form
PUT  /admin/guests/:id       → Update guest
DELETE /admin/guests/:id     → Delete guest
GET  /admin/guests/:id/qr    → QR code PNG download

GET  /admin/content          → List content blocks
GET  /admin/content/:key     → Edit block (DE + EN)
POST /admin/content/:key     → Save block

GET  /admin/config           → Site config form
POST /admin/config           → Save config

GET  /admin/newsletter       → Compose newsletter
POST /admin/newsletter       → Send newsletter

GET  /admin/export/csv       → Guestlist CSV download
```

---

## 6. CMS Content Blocks

Editable in admin (DE + EN), Markdown, rendered server-side via `goldmark`.  
Template function: `{{ content "key" .Lang }}`.


| Key                    | Label                     | Used on          | Seed content                                                                                               |
| ---------------------- | ------------------------- | ---------------- | ---------------------------------------------------------------------------------------------------------- |
| `hero_tagline`         | Tagline                   | `/me` hero       | *(empty, set by admin)*                                                                                    |
| `event_description`    | About the event           | `/me`            | *(empty)*                                                                                                  |
| `what_to_expect`       | Food, drinks, pool, tent  | `/me`            | *(empty)*                                                                                                  |
| `kids_section`         | Kids activities           | `/me`            | DE: "Maltisch, Plantschbecken und jede Menge Bälle!" · EN: "Craft table, paddling pool and lots of balls!" |
| `charity_name`         | Charity name              | `/me`, confirmed | *(empty — set when chosen)*                                                                                |
| `charity_description`  | Charity info + donate CTA | `/me`            | *(empty)*                                                                                                  |
| `special_thing`        | Teaser for the surprise   | `/me`            | *(empty — fill when decided)*                                                                              |
| `rsvp_note`            | Text near RSVP CTA        | `/me`            | *(empty)*                                                                                                  |
| `confirmation_message` | Post-RSVP message         | `/me/confirmed`  | *(empty)*                                                                                                  |
| `footer_note`          | Closing remark            | all guest pages  | *(empty)*                                                                                                  |


`charity_url` lives in `site_config` (link only, not markdown).

---

## 7. RSVP Form (`POST /me/rsvp`)

Fields:

- **Attending?** Yes / No (required, radio)
- **Bringing a +1?** Yes / No → name field appears via JS (everyone gets +1)
- **How many kids?** 0–5 select (optional)
- **Favorite song** text (optional)
- **Message for Florian** textarea (optional)
- **Email** text (optional) → if filled: newsletter checkbox appears via JS

Behavior:

- Prefilled if already RSVPed (update flow)
- POST → upsert guest → if email present: send confirmation email → redirect `/me/confirmed`

---

## 8. Language Toggle

- Default: **German**
- Button "EN" / "DE" in top corner → sets `lang` cookie → reload
- Templates check cookie: `de` → `body_de`, else `body_en`
- Backend stays English throughout
- UI strings in `locales/locales.go` — two maps, no external dep

---

## 9. Design — "Midnight Pool"

**Vibe:** Deep navy · electric teal · atmospheric · premium · pool at night.

### Palette


| Token          | Value     | Usage                   |
| -------------- | --------- | ----------------------- |
| `--bg`         | `#050d1a` | Page background         |
| `--surface`    | `#0c1829` | Cards, forms            |
| `--accent`     | `#00d4aa` | Primary CTA, highlights |
| `--accent-2`   | `#7cfcff` | Ice blue glow, shimmer  |
| `--text`       | `#e8f4f8` | Body text               |
| `--text-muted` | `#6b8fa3` | Secondary text          |
| `--glow`       | `#3a7bd5` | Pool-water glow effects |


### Typography

- Display / name: Space Grotesk or Inter, light weight, very large
- Body: Inter or system-ui, regular

### Animations (CSS only, no heavy JS)

- Spell page: slow undulating mesh gradient — looks like pool water at night
- Personal page: hero name fades in large, sections slide up on scroll
- Subtle shimmer on accent elements

### Spell Page Specific

Full-screen dark. No party name. No context. Just:

- Animated pool-water background
- Small centered sigil/icon (waves or similar)
- Single input: placeholder `"Dein Zauberspruch…"` (DE) / `"Your spell…"` (EN)
- Enter key or button submits
- Error: `"Das war kein Zauber."` / `"That was no spell."` — cryptic, not technical

---

## 10. Invite Cards — Concert Ticket Style

Rendered as styled HTML per guest in admin → browser print → A6 or 2-up on A4. No PDF library.

**Layout (horizontal):**

```
┌─────────────────────────────────┬──────────────┐
│  SUMMER PARTY 2026              │  ░░░░░░░░░░  │
│  01. August 2026 · 16:00        │  ░░ QR  ░░░  │
│                                 │  ░░░░░░░░░░  │
│  [GUEST NAME large]             │              │
│                                 │  [NAME]      │
│  party.example.com              │  ADMIT ONE   │
└─────────────────────────────────┴──────────────┘
```

- Left panel: party branding, date, guest name large, URL
- Right stub: QR code + name + "ADMIT ONE"  
- Decorative dashed perforation line between panels
- Midnight Pool color scheme
- QR encodes `https://party.example.com/?spell=XXXXX` — scanning prefills spell input

---

## 11. Admin Panel

Functional, minimal CSS (Bootstrap 5 or plain).

### Dashboard

- Accepted / Declined / Pending / No-response counts
- Total headcount: guests + +1s + kids
- Days until party

### Guestlist

- Columns: Name · Code · Status · +1 · Kids · Song · Newsletter · RSVP date · Actions
- Filter tabs by status
- Quick status toggle (inline select or buttons)
- "Copy link" button (copies `https://…/?spell=CODE`)
- QR download button per guest
- "Print card" link → opens card HTML in new tab

### Newsletter Composer

- Subject + markdown body textarea
- Recipient count shown before send
- Send button with confirmation dialog

### Content Editor

- List all blocks with label + last updated
- Edit: DE textarea + EN textarea (markdown), save

### Site Config

- Key-value form: date, party name, RSVP deadline, charity URL, etc.

---

## 12. Email

### Confirmation Email (on RSVP, if email provided)

- Subject: `"Deine Anmeldung – Summer Party 2026"` / `"Your RSVP – Summer Party 2026"`
- Body: name, status, song if given, charity link if set
- Template: `templates/email/confirmation.html`
- Sent immediately after POST /me/rsvp

### Newsletter (manual send from admin)

- Compose in admin: subject + markdown body
- Recipients: `newsletter = 1 AND email IS NOT NULL`
- Unsubscribe link in footer → `GET /unsubscribe?token=...` → `newsletter = 0`
- Token: HMAC-SHA256 of guest ID (no extra table)

---

## 13. Admin Auth

Single admin user. Credentials in `site_config` (`admin_user`, `admin_password_hash` bcrypt). No user table. or env vars

---

## 14. Infrastructure

### docker-compose.yml

```yaml
services:
  app:
    build: .
    ports:
      - "3000:3000"
    env_file: .env
    volumes:
      - ./data:/app/data
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
      - ./data/certs:/etc/letsencrypt:ro
    depends_on:
      - app
    restart: unless-stopped
```

### .env.example

```
PORT=3000
DATABASE_PATH=./data/party.db
SESSION_SECRET=<random-32-bytes>
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
SMTP_FROM=you@example.com
SMTP_FROM_NAME=Florian
BASE_URL=https://party.example.com
GO_ENV=production
```

### Dockerfile (multi-stage)

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o party ./cmd/app

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/party .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/migrations ./migrations
RUN mkdir -p /app/data
EXPOSE 3000
CMD ["./party"]
```

---

## 15. Project Structure

```
/
├── cmd/app/main.go
├── handlers/
│   ├── guest.go          ← spell login, /me, rsvp
│   ├── admin.go          ← dashboard, guestlist
│   ├── content.go        ← CMS editor
│   └── newsletter.go     ← compose + send
├── models/
│   ├── guest.go
│   ├── content_block.go
│   └── site_config.go
├── middleware/
│   ├── session.go        ← guest session check
│   └── admin.go          ← admin auth check
├── locales/
│   └── locales.go        ← DE + EN string maps
├── templates/
│   ├── layout.html
│   ├── spell.html        ← the mysterious login page
│   ├── me/
│   │   ├── index.html
│   │   ├── rsvp.html
│   │   └── confirmed.html
│   ├── admin/
│   │   ├── login.html
│   │   ├── dashboard.html
│   │   ├── guests.html
│   │   ├── guest_edit.html
│   │   ├── content.html
│   │   ├── config.html
│   │   └── newsletter.html
│   └── email/
│       └── confirmation.html
├── migrations/
│   ├── 001_create_guests.sql
│   ├── 002_create_content_blocks.sql
│   └── 003_create_site_config.sql
├── assets/
│   ├── css/main.css
│   └── js/main.js
├── docker-compose.yml
├── Dockerfile
└── .env.example
```

---

## 16. Decisions Log


| #   | Decision                                                                            |
| --- | ----------------------------------------------------------------------------------- |
| 1   | Design theme: **Midnight Pool** (deep navy + electric teal)                         |
| 2   | Invite card style: **Concert Ticket** (horizontal, perforation stub, QR right)      |
| 3   | Charity: CMS placeholder + `charity_url` in site_config — fill when chosen          |
| 4   | Confirmation email: **yes**, sent on RSVP if email provided                         |
| 5   | Special thing 2026: TBD — `special_thing` CMS block ready, fill later               |
| 6   | Kids activities: Maltisch · Plantschbecken · Bälle (seeded in `kids_section` block) |
| 7   | RSVP deadline: **2026-07-15**                                                       |
| 8   | +1 policy: **everyone** gets one                                                    |
| 9   | QR encodes `/?spell=XXXXX` → prefills spell input on scan                           |


