# Warehouse Inventory Management System API

Simple Golang + PostgreSQL REST API for managing warehouse master data (barang), stock, and transactions (pembelian / penjualan) with proper transactional updates and history logging.

## Tech Stack

- Golang
- PostgreSQL
- Router: chi
- DB: database/sql + lib/pq
- Env loader: godotenv
- Auth: JWT (golang-jwt) + bcrypt

## Prerequisites

- Go 1.20+
- PostgreSQL 13+
- Git

## Quick Start

1. Clone & install deps:

```bash
git clone <your-repo-url>
cd warehouse
go mod tidy
```

2. Create `.env` (adjust values):

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=warehouse_db
DB_SSLMODE=disable
```

3. Create DB & apply schema:

```bash
createdb -h localhost -U postgres warehouse_db
psql -h localhost -U postgres -d warehouse_db -f schema.sql
psql -h localhost -U postgres -d warehouse_db -f seed.sql   # optional seed
```

4. Run server:

```bash
go run .
# Server listening on :8080
```

5. Health check:

```bash
curl http://localhost:8080/health
```

## Authentication & Authorization

The API uses a pair of JWT tokens:

- Access Token: expires in 15 minutes
- Refresh Token: expires in 7 days

Public endpoints:

- `POST /api/login`
- `POST /api/refresh`
- `GET /health`

Protected endpoints: all other `/api/*` routes require header:

```
Authorization: Bearer <access_token>
```

### Login Flow

Request:

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<your-password>"}'
```

Response:

```json
{
  "success": true,
  "message": "Login success",
  "data": {
    "access_token": "<access_token>",
    "refresh_token": "<refresh_token>"
  }
}
```

### Refresh Flow

Request:

```bash
curl -X POST http://localhost:8080/api/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

Response:

```json
{
  "success": true,
  "message": "Token refreshed",
  "data": {
    "access_token": "<new_access_token>
    ", "refresh_token": "<new_refresh_token>"
  }
}
```

### Postman Usage (Recommended)

Environment variables (example):

```
BASE_URL = http://localhost:8080
ACCESS_TOKEN = (set by test script)
REFRESH_TOKEN = (set by test script)
```

Login request Test tab script:

```javascript
if (pm.response.code === 200) {
  const json = pm.response.json();
  if (json && json.data) {
    if (json.data.access_token)
      pm.environment.set("ACCESS_TOKEN", json.data.access_token);
    if (json.data.refresh_token)
      pm.environment.set("REFRESH_TOKEN", json.data.refresh_token);
  }
}
```

Using saved access token in other requests (Headers):

```
Authorization: Bearer {{ACCESS_TOKEN}}
```

Refreshing token Test tab script:

```javascript
if (pm.response.code === 200) {
  const json = pm.response.json();
  if (json && json.data) {
    if (json.data.access_token)
      pm.environment.set("ACCESS_TOKEN", json.data.access_token);
    if (json.data.refresh_token)
      pm.environment.set("REFRESH_TOKEN", json.data.refresh_token);
  }
}
```

### Roles

- Example roles: `admin`, `user` (or `staff`).
- Only `admin` can delete barang (`DELETE /api/barang/{id}`).
- Both `admin` and `staff` can create transactions (pembelian / penjualan).

### Seed Credentials

From `seed.sql`:

- Username: `admin` (role: admin)
- Username: `user1` (role: user)
  Passwords are stored as bcrypt hashes. If you do not know the plain password, update it:

```bash
go run tools/hash_password.go new-password
# Copy hash and UPDATE users SET password='<hash>' WHERE username='admin';
```

### Important Notes

- `user_id` for transactions is taken from JWT (not from request body).
- Access tokens are short (15m); always refresh before they expire using `/api/refresh`.
- Keep refresh tokens secret; they allow minting new access tokens.

## API Endpoints (Summary)

### Master Barang

`GET /api/barang?search=&page=&limit=` – List with search & pagination
`GET /api/barang/{id}` – Detail (includes stok if implemented)
`GET /api/barang/stok` – List barang + current stok
`POST /api/barang` – Create
`PUT /api/barang/{id}` – Update
`DELETE /api/barang/{id}` – Delete (admin only)

### Stok & History

`GET /api/stok` – All current stock
`GET /api/stok/{barang_id}` – Stock by barang
`GET /api/history-stok?page=&limit=` – Paginated stock history
`GET /api/history-stok/{barang_id}?page=&limit=` – History by barang

### Pembelian

`POST /api/pembelian` – Create (auto update stok + history)
`GET /api/pembelian?page=&limit=&from=&to=` – Paginated + date filter
`GET /api/pembelian/{id}` – Header + details
Body example:

```json
{
  "no_faktur": "PB-001",
  "supplier": "PT ABC",
  "details": [
    { "barang_id": 1, "qty": 5, "harga": 12000 },
    { "barang_id": 2, "qty": 3, "harga": 15000 }
  ]
}
```

### Penjualan

`POST /api/penjualan` – Create (validates stok, auto update + history)
`GET /api/penjualan?page=&limit=&from=&to=` – Paginated + date filter
`GET /api/penjualan/{id}` – Header + details
Body example:

```json
{
  "no_faktur": "SJ-001",
  "customer": "PT XYZ",
  "details": [{ "barang_id": 1, "qty": 2, "harga": 15000 }]
}
```

### Laporan

`GET /api/laporan/stok`
`GET /api/laporan/penjualan?from=&to=`
`GET /api/laporan/pembelian?from=&to=`

## Transactions & Stock Logic

Pembelian:

- Validate barang, compute totals
- Update/insert `mstok` (add qty)
- Insert `history_stok` (jenis_transaksi = pembelian)
- All inside one DB transaction

Penjualan:

- Validate barang + stok available
- Update `mstok` (subtract qty)
- Insert `history_stok` (jenis_transaksi = penjualan)
- Rollback on any error

## Pagination & Search

Responses can include:

```json
"meta": { "page": 1, "limit": 10, "total": 42 }
```

Search uses `ILIKE` on `nama_barang` and `kode_barang`.

## Utilities

Generate bcrypt hash:

```bash
go run tools/hash_password.go my-password
```

## Development Tips

- Keep access tokens short-lived for security.
- Refresh token rotation reduces risk of theft.
- Consider moving JWT secret to `.env` for production.

## License

See `LICENSE` file (if present).
