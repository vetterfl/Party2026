# Party 2026

## Dev

```bash
cp .env.example .env
# edit .env — set SESSION_SECRET, BASE_URL, ADMIN_USER, ADMIN_PASSWORD

mkdir -p data
go run ./cmd/app/
```

App runs on `http://localhost:3000`.  
Admin: `http://localhost:3000/admin/login`

## Production

```bash
docker compose up -d
```

Requires `.env` with real SMTP + SESSION_SECRET + BASE_URL.

## Email (SMTP)

Used for:

- RSVP confirmation emails (when a guest submits an email on the RSVP form)
- Admin newsletter sends

### Configuration

Set these in `.env`:

| Variable | Description |
|----------|-------------|
| `SMTP_HOST` | Mail server hostname |
| `SMTP_PORT` | Usually `465` (SSL) or `587` (STARTTLS) |
| `SMTP_TLS` | Optional. Set to `ssl` to force implicit TLS (also auto-enabled for port `465`) |
| `SMTP_USER` | Login username (leave empty if the server needs no auth) |
| `SMTP_PASS` | Login password |
| `SMTP_FROM` | Envelope / From address (required) |
| `SMTP_FROM_NAME` | Display name in the From header (fallback if not set in admin config) |

The admin **Site Config** field “Email From Name” (`smtp_from_name`) overrides `SMTP_FROM_NAME` for outgoing mail when set.

### SSL on port 465

Many providers (including typical shared hosting) expect **implicit SSL** on port **465**. Use:

```env
SMTP_HOST=mail.example.com
SMTP_PORT=465
SMTP_FROM=you@example.com
SMTP_USER=you@example.com
SMTP_PASS=your-password
```

Port `465` uses TLS from the first byte. Port `587` uses plain TCP first, then STARTTLS if the server supports it.

### Test SMTP

Send a test message without starting the web server:

```bash
go run ./cmd/app mailtest you@example.com
```

The command loads `.env`, prints the configured host/port/from, sends a short HTML test email, and exits with `mail test OK` or an error message.

Use this to verify credentials and TLS before testing RSVP or newsletter in the admin UI.
