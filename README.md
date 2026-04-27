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
