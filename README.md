# Warehouse Inventory Management System API

Golang-based REST API for a simple Warehouse Inventory system. It includes Master Barang management, Stok monitoring, and transactional Pembelian/Penjualan with proper database transactions and stock validations.

## Tech Stack

- Golang
- PostgreSQL
- Router: chi (github.com/go-chi/chi/v5)
- DB: database/sql + lib/pq
- Env loader: godotenv
- Auth: JWT (github.com/golang-jwt/jwt/v5) + bcrypt (golang.org/x/crypto/bcrypt)

## Prerequisites

- Go (1.20+ recommended)
- PostgreSQL (13+ recommended)
- Git

## Getting Started

1. Clone repo and install dependencies

```bash
git clone <your-repo-url>
cd warehouse
go mod tidy
```

2. Create `.env`

```bash
cp .env.example .env  # if available, else create new file
```

Minimal `.env` example:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=warehouse_db
DB_SSLMODE=disable
```

3. Create database and apply schema

```bash
# create database (adjust credentials as needed)
createdb -h localhost -U postgres warehouse_db

# apply schema
psql -h localhost -U postgres -d warehouse_db -f schema.sql
```

4. Run the server

```bash
go run .
# Server listening on :8080
```

Health check:

```bash
curl http://localhost:8080/health
```

## Authentication

- Public endpoint: `POST /api/login`
- All other `/api/*` routes are protected with Bearer JWT.

Login request:

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<your-password>"}'
```

Successful response:

```json
{
  "success": true,
  "message": "Login success",
  "data": { "token": "<jwt>" }
}
```

Use the token:

```bash
curl http://localhost:8080/api/barang \
  -H "Authorization: Bearer <jwt>"
```

## API Endpoints

Base URL: `http://localhost:8080`

### Auth (Public)

- POST `/api/login` — Get JWT token

### Master Barang

- All routes below require `Authorization: Bearer <jwt>` header
- GET `/api/barang?search=&page=&limit=` — List with search + pagination
- GET `/api/barang/{id}` — Detail by ID
- POST `/api/barang` — Create
- PUT `/api/barang/{id}` — Update
- DELETE `/api/barang/{id}` — Delete (admin only)

### Stok Management

- Requires Bearer token
- GET `/api/stok` — List all current stock
- GET `/api/stok/{barang_id}` — Stock for a specific barang
- GET `/api/history-stok?page=&limit=` — List movement history (pagination supported; returns meta)
- GET `/api/history-stok/{barang_id}?page=&limit=` — Movement history by barang (pagination supported; returns meta)

### Transaksi Pembelian

- Requires Bearer token
- POST `/api/pembelian` — Create pembelian (transactional)
- GET `/api/pembelian?page=&limit=&from=&to=` — List headers (pagination + optional date range)
- GET `/api/pembelian/{id}` — Detail (header + details)

Contoh body pembelian:

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

### Transaksi Penjualan

- Requires Bearer token
- POST `/api/penjualan` — Create penjualan (transactional, with stock validation)
- GET `/api/penjualan?page=&limit=&from=&to=` — List headers (pagination + optional date range)
- GET `/api/penjualan/{id}` — Detail (header + details)

Contoh body penjualan:

```json
{
  "no_faktur": "SJ-001",
  "customer": "PT XYZ",
  "details": [{ "barang_id": 1, "qty": 2, "harga": 15000 }]
}
```

### Laporan

- Requires Bearer token
- GET `/api/laporan/stok` — Stock report (barang + stok_akhir)
- GET `/api/laporan/penjualan?from=&to=` — Sales report (headers; optional date range)
- GET `/api/laporan/pembelian?from=&to=` — Purchase report (headers; optional date range)

## Transactions & Stock Validation

- Pembelian (Create):

  - Runs inside a single database transaction.
  - Validates barang exists, computes subtotal/total.
  - Inserts header and details.
  - Updates or inserts stock in `mstok` (stok_akhir += qty).
  - Writes `history_stok` with jenis_transaksi = "pembelian".
  - Any error triggers ROLLBACK; otherwise COMMIT.

- Penjualan (Create):
  - Runs inside a single database transaction with row-level lock on stock.
  - Validates barang exists and stock availability (stok_akhir >= qty) BEFORE updating.
  - Inserts header and details.
  - Updates stock in `mstok` (stok_akhir -= qty).
  - Writes `history_stok` with jenis_transaksi = "penjualan".
  - Insufficient stock or any error triggers ROLLBACK; otherwise COMMIT.

## Notes

- Search uses PostgreSQL `ILIKE` on `nama_barang` and `kode_barang`.
- Pagination uses `LIMIT` and `OFFSET` with meta `{ page, limit, total }` in responses.
- Check `schema.sql` for full table definitions and constraints.
- JWT secret is a simple constant in code for demo purposes.
- Seed users use bcrypt-hashed passwords. To generate your own hash:

```bash
cd warehouse
go run tools/hash_password.go your-plain-password
```
