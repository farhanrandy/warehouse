# Warehouse Inventory Management System API

Golang-based REST API for a simple Warehouse Inventory system. It includes Master Barang management, Stok monitoring, and transactional Pembelian/Penjualan with proper database transactions and stock validations.

## Tech Stack

- Golang
- PostgreSQL
- Router: chi (github.com/go-chi/chi/v5)
- DB: database/sql + lib/pq
- Env loader: godotenv

## Prerequisites

- Go (1.20+ recommended)
- PostgreSQL (13+ recommended)
- Git

## Getting Started

1) Clone repo and install dependencies

```bash
git clone <your-repo-url>
cd warehouse
go mod tidy
```

2) Create `.env`

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

3) Create database and apply schema

```bash
# create database (adjust credentials as needed)
createdb -h localhost -U postgres warehouse_db

# apply schema
psql -h localhost -U postgres -d warehouse_db -f schema.sql
```

4) Run the server

```bash
go run .
# Server listening on :8080
```

Health check:

```bash
curl http://localhost:8080/health
```

## API Endpoints

Base URL: `http://localhost:8080`

### Master Barang

- GET `/api/barang?search=&page=&limit=` — List with search + pagination
- GET `/api/barang/{id}` — Detail by ID
- POST `/api/barang` — Create
- PUT `/api/barang/{id}` — Update
- DELETE `/api/barang/{id}` — Delete

### Stok Management

- GET `/api/stok` — List all current stock
- GET `/api/stok/{barang_id}` — Stock for a specific barang
- GET `/api/history-stok` — List movement history
- GET `/api/history-stok/{barang_id}` — Movement history by barang

### Transaksi Pembelian

- POST `/api/pembelian` — Create pembelian (transactional)
- GET `/api/pembelian` — List headers
- GET `/api/pembelian/{id}` — Detail (header + details)

Contoh body pembelian:

```json
{
	"no_faktur": "PB-001",
	"supplier": "PT ABC",
	"user_id": 1,
	"details": [
		{ "barang_id": 1, "qty": 5, "harga": 12000 },
		{ "barang_id": 2, "qty": 3, "harga": 15000 }
	]
}
```

### Transaksi Penjualan

- POST `/api/penjualan` — Create penjualan (transactional, with stock validation)
- GET `/api/penjualan` — List headers
- GET `/api/penjualan/{id}` — Detail (header + details)

Contoh body penjualan:

```json
{
	"no_faktur": "SJ-001",
	"customer": "PT XYZ",
	"user_id": 1,
	"details": [
		{ "barang_id": 1, "qty": 2, "harga": 15000 }
	]
}
```

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
