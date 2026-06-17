# Party 2026

[![Docker Image](https://github.com/vetterfl/Party2026/actions/workflows/docker-image.yml/badge.svg?branch=main)](https://github.com/vetterfl/Party2026/actions/workflows/docker-image.yml)
[![License: MIT](https://img.shields.io/github/license/vetterfl/Party2026)](LICENSE)

Self-hosted invitation + RSVP web app for a personal summer party. Each guest
gets a unique "spell" code, lands on a personalised page, RSVPs, optionally
opts in to a newsletter. The host manages the guest list, content blocks,
themes, and newsletter through an admin UI.

- **Language:** Go 1.23 (Echo v4)
- **Storage:** SQLite (single file, WAL mode)
- **Migrations:** Goose (embedded in the binary)
- **Templates + assets:** embedded via `embed.FS`
- **Auth:** bcrypt + cookie sessions (gorilla/sessions)
- **Localisation:** German (default) + English
- **Deployment:** single binary or `docker compose` with bundled nginx

## Features

| Area | What it does |
|------|--------------|
| Spell login | Guests enter a 6-letter code (also linkable as `/?spell=CODE`) and land on a personalised page. |
| Personal page | Greeting, party info, configurable content blocks (Markdown). |
| RSVP | Accept/decline, plus-one + name, kids count, favourite song, comment, email + newsletter opt-in. |
| Confirmation email | Sent on RSVP submission. Includes recap and links to update RSVP / unsubscribe. |
| Calendar export | Optional `/me/calendar.ics` download for guests who accepted. |
| Themes | Multiple visual themes shipped under `assets/themes/<name>/`. Admin selects per page. |
| Content blocks | Markdown blocks edited in admin, rendered into guest pages. |
| Admin dashboard | Headcount stats, days-until-party, quick links. |
| Guest management | CRUD, status filters, CSV export, printable invite cards (A6 ticket), QR codes, WhatsApp/Signal share-message helpers. |
| Newsletter | Bulk send to opted-in guests with Markdown body, per-recipient `{name}` placeholder, automatic unsubscribe link. |
| Multi-admin | Multiple admin accounts via UI, with self-deletion / last-admin safeguards. |

## Quick start (development)

Requires **Go 1.23+** (CGO enabled for sqlite — needs a C toolchain).

```bash
cp .env.example .env
# Edit .env: set SESSION_SECRET, BASE_URL, ADMIN_USER, ADMIN_PASSWORD

mkdir -p data
go run ./cmd/app
```

App listens on `http://localhost:3000` by default.
Admin login: `http://localhost:3000/admin/login`.

After first start, the bootstrap `ADMIN_USER` / `ADMIN_PASSWORD` from `.env`
seed (or overwrite) the first admin account. **Remove those two lines from
`.env` after the first run** — they overwrite the stored password on every
restart.

## Production (Docker)

```bash
docker compose up -d
```

The compose file builds the app, mounts `./data` for the SQLite database, and
serves through nginx on ports 80/443.

Before going live:
- Edit `nginx.conf` — replace `party.example.com` with your real hostname (in
  both the `server_name` line and the certificate paths).
- Provide TLS certificates under `data/certs/live/<hostname>/` (or symlink to
  your existing Let's Encrypt directory).
- Set `GO_ENV=production` in `.env` so the app:
  - sets `Secure` on session + CSRF cookies,
  - refuses to start without a strong `SESSION_SECRET`.

## Environment variables

| Var | Required | Default | Description |
|-----|----------|---------|-------------|
| `PORT` | no | `3000` | App listen port |
| `DATABASE_PATH` | no | `./data/party.db` | SQLite file path |
| `SESSION_SECRET` | **yes in production** | — | ≥32 random bytes hex. Generate with `openssl rand -hex 32`. Also used to sign unsubscribe tokens. |
| `BASE_URL` | yes | `http://localhost:$PORT` | Public origin. Used in invite links, QR codes, emails. |
| `GO_ENV` | no | unset | Set to `production` to enable Secure cookies and the session-secret guard. |
| `ADMIN_USER` | no | — | Bootstrap admin username (applied on every start if set). |
| `ADMIN_PASSWORD` | no | — | Bootstrap admin password. |
| `SMTP_HOST` | yes for email | — | Mail server host |
| `SMTP_PORT` | no | `587` | `465` for implicit TLS, `587` for STARTTLS |
| `SMTP_TLS` | no | — | Set to `ssl` to force implicit TLS (also auto-enabled on port 465) |
| `SMTP_USER` | no | — | SMTP login user (empty if server allows unauth) |
| `SMTP_PASS` | no | — | SMTP password |
| `SMTP_FROM` | yes for email | — | Envelope/From address |
| `SMTP_FROM_NAME` | no | `Florian` | Display name; overridden by the admin "Email From Name" config field if set |

## Email (SMTP)

Used for:
- RSVP confirmation messages (sent when a guest submits an email on the form),
- Admin newsletter blasts.

### SSL on port 465

Many providers expect **implicit TLS** on port 465. Configure:

```env
SMTP_HOST=mail.example.com
SMTP_PORT=465
SMTP_FROM=you@example.com
SMTP_USER=you@example.com
SMTP_PASS=your-password
```

Port 465 → TLS from the first byte.
Port 587 → plain TCP, then STARTTLS if offered.

### Test SMTP without booting the web app

```bash
go run ./cmd/app mailtest you@example.com
```

Loads `.env`, prints the resolved SMTP settings, sends a short HTML test
message, exits with `mail test OK` or an error.

## Themes

Themes live under `assets/themes/<name>/` and bundle:
- `theme.css` (required) — guest-page styles
- `theme.js` (optional) — guest-page interactivity

The app discovers themes at startup by listing the directory. Admin → Config
exposes a dropdown per page (login + guest). Add a new theme by dropping a
folder with the two files and restarting.

## Content blocks

Three Markdown blocks live in the `content_blocks` table and are edited via
Admin → Content:

- `general` — main page body (the long text on `/me`)
- `rsvp_note` — note shown above the RSVP CTA
- `confirmation_message` — message shown on the confirmed page

Markdown is rendered with goldmark.

## Admin: multi-admin accounts

`Admin → Admins` lets you add and remove admin accounts, and change your own
password (current password required). Two guardrails apply:
- you cannot delete your own account,
- the last remaining admin cannot be deleted.

If you lock yourself out, re-add `ADMIN_USER`/`ADMIN_PASSWORD` to `.env` and
restart — the bootstrap step recreates or overwrites the account.

## Database

- SQLite, single file under `./data/` by default
- Schema is managed by `goose` migrations in `migrations/`, embedded into the
  binary; new migrations run automatically at startup
- WAL mode + foreign keys are enabled at open time
- To back up: stop the app and copy `data/party.db` (or use SQLite's online
  backup API / `.backup` command)

## Security notes

- CSRF tokens are required on every unsafe method (cookie + form field).
- Session cookies are HttpOnly, SameSite=Lax, and `Secure` in production.
- Login endpoints (`/login`, `/admin/login`) are rate-limited per IP.
- Echo's `Secure` middleware sets X-Frame-Options, X-Content-Type-Options,
  Referrer-Policy etc. HSTS is left to nginx (already adequate when fronted
  by TLS-terminating nginx).
- Failed admin login still runs a bcrypt comparison so existing usernames
  cannot be inferred from response timing.
- Access logs redact the `spell` query parameter so invite codes do not leak.
- Unsubscribe links are signed with `SESSION_SECRET` (HMAC-SHA256).

If you forked this and intend to run it publicly:
- generate a fresh `SESSION_SECRET` (`openssl rand -hex 32`),
- rotate `SMTP_PASS` after any commit/sharing that exposed it,
- keep `data/party.db` out of backups that leave your trust boundary (it
  contains guest names, emails, phone numbers).

## Project layout

```
cmd/app/main.go         entry point, routing, middleware setup
embed.go                embeds assets + templates
handlers/               HTTP handlers (admin, guest, mailer, calendar, …)
middleware/             session, CSRF helpers, auth gates
models/                 sqlx stores + types
migrations/             goose .sql files (embedded)
locales/                DE/EN UI strings
templates/              html/template files (admin, me, email, _partials)
assets/                 static files: css, js, fonts, themes
```

## License

Personal project — no license declared. Treat as "all rights reserved" unless
the repo owner says otherwise.
