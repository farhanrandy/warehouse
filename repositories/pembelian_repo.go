package repositories

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "time"

    "warehouse/models"
)

// PembelianRepo handles purchase transaction persistence.
type PembelianRepo struct {
    DB *sql.DB
}

func NewPembelianRepo(db *sql.DB) *PembelianRepo { return &PembelianRepo{DB: db} }

// CreatePembelianTx performs a purchase (pembelian) transaction atomically.
// Steps (ALL inside a single DB transaction):
//  1. Validate each barang exists.
//  2. Compute subtotal for each detail & accumulate total header.
//  3. Insert beli_header (returning its id).
//  4. For each detail:
//       a. Lock current stock row (mstok) for that barang (SELECT ... FOR UPDATE).
//       b. Insert beli_detail.
//       c. Update / insert mstok (add qty to stok_akhir).
//       d. Insert history_stok capturing before/after.
//  5. If any error occurs -> ROLLBACK.
//  6. If all succeed -> COMMIT.
// On success BeliHeader.ID is populated; Details keep their inserted IDs.
func (r *PembelianRepo) CreatePembelianTx(ctx context.Context, hdr *models.BeliHeader) error {
    if hdr == nil { return errors.New("header is nil") }
    if len(hdr.Details) == 0 { return errors.New("details empty") }

    // Begin a database transaction.
    tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{})
    if err != nil { return fmt.Errorf("begin tx: %w", err) }

    // Helper to rollback on failure.
    rollback := func(e error) error {
        _ = tx.Rollback()
        return e
    }

    // 1. Validate barang exists for every detail.
    for i, d := range hdr.Details {
        var exists bool
        // Use SELECT 1 pattern for existence check.
        if err := tx.QueryRowContext(ctx, "SELECT 1 FROM master_barang WHERE id=$1", d.BarangID).Scan(&exists); err != nil {
            if err == sql.ErrNoRows {
                return rollback(fmt.Errorf("barang id %d not found (detail index %d)", d.BarangID, i))
            }
            return rollback(fmt.Errorf("validate barang: %w", err))
        }
    }

    // 2. Compute subtotal & total if not already set.
    var total int64
    for i := range hdr.Details {
        d := &hdr.Details[i]
        if d.Qty <= 0 { return rollback(fmt.Errorf("qty must be > 0 for barang %d", d.BarangID)) }
        if d.Harga < 0 { return rollback(fmt.Errorf("harga must be >= 0 for barang %d", d.BarangID)) }
        if d.Subtotal == 0 { d.Subtotal = d.Qty * d.Harga }
        total += d.Subtotal
    }
    hdr.Total = total
    if hdr.Status == "" { hdr.Status = "completed" }

    // 3. Insert beli_header.
    // created_at defaults in table, but we capture it via RETURNING.
    err = tx.QueryRowContext(ctx, `INSERT INTO beli_header (no_faktur, supplier, total, user_id, status)
            VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
        hdr.NoFaktur, hdr.Supplier, hdr.Total, hdr.UserID, hdr.Status,
    ).Scan(&hdr.ID, &hdr.CreatedAt)
    if err != nil { return rollback(fmt.Errorf("insert header: %w", err)) }

    // 4. Process each detail.
    for i := range hdr.Details {
        d := &hdr.Details[i]

        // 4a. Lock current stock for barang (SELECT ... FOR UPDATE).
        var stokBefore sql.NullInt64
        err = tx.QueryRowContext(ctx, "SELECT stok_akhir FROM mstok WHERE barang_id=$1 FOR UPDATE", d.BarangID).Scan(&stokBefore)
        if err != nil && err != sql.ErrNoRows {
            return rollback(fmt.Errorf("lock stock: %w", err))
        }
        var before int64
        if stokBefore.Valid { before = stokBefore.Int64 }
        after := before + d.Qty

        // 4b. Insert beli_detail.
        err = tx.QueryRowContext(ctx, `INSERT INTO beli_detail (beli_header_id, barang_id, qty, harga, subtotal)
                VALUES ($1,$2,$3,$4,$5) RETURNING id`,
            hdr.ID, d.BarangID, d.Qty, d.Harga, d.Subtotal,
        ).Scan(&d.ID)
        if err != nil { return rollback(fmt.Errorf("insert detail: %w", err)) }

        // 4c. Update or Insert mstok.
        if stokBefore.Valid {
            // Row exists -> update stok_akhir by adding qty.
            res, uErr := tx.ExecContext(ctx, "UPDATE mstok SET stok_akhir=$1 WHERE barang_id=$2", after, d.BarangID)
            if uErr != nil { return rollback(fmt.Errorf("update mstok: %w", uErr)) }
            if rows, _ := res.RowsAffected(); rows == 0 {
                return rollback(errors.New("expected mstok update to affect 1 row"))
            }
        } else {
            // No row -> insert new stock record.
            _, iErr := tx.ExecContext(ctx, "INSERT INTO mstok (barang_id, stok_akhir) VALUES ($1,$2)", d.BarangID, after)
            if iErr != nil { return rollback(fmt.Errorf("insert mstok: %w", iErr)) }
        }

        // 4d. Insert history_stok (recording before & after).
        _, hErr := tx.ExecContext(ctx, `INSERT INTO history_stok (barang_id, user_id, jenis_transaksi, jumlah, stok_sebelum, stok_sesudah)
                VALUES ($1,$2,$3,$4,$5,$6)`,
            d.BarangID, hdr.UserID, "pembelian", d.Qty, before, after,
        )
        if hErr != nil { return rollback(fmt.Errorf("insert history: %w", hErr)) }
    }

    // 5. Commit transaction.
    if err = tx.Commit(); err != nil {
        return fmt.Errorf("commit tx: %w", err)
    }

    // 6. Optionally set CreatedAt if not returned (fallback).
    if hdr.CreatedAt.IsZero() { hdr.CreatedAt = time.Now() }
    return nil
}
