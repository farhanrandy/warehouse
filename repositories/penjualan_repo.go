package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"warehouse/models"
)

// PenjualanRepo handles sales transaction persistence.
type PenjualanRepo struct {
    DB *sql.DB
}

func NewPenjualanRepo(db *sql.DB) *PenjualanRepo { return &PenjualanRepo{DB: db} }

// CreatePenjualanTx performs a sales (penjualan) transaction atomically.
// Steps (ALL inside a single DB transaction):
//  1. Validate each barang exists.
//  2. Compute subtotal for each detail & accumulate total header.
//  3. Insert jual_header (returning its id & created_at).
//  4. For each detail:
//       a. Lock current stock row (mstok) for that barang (SELECT ... FOR UPDATE).
//       b. Validate stock availability (stok_akhir >= qty); if not, rollback with clear error.
//       c. Insert jual_detail.
//       d. Update mstok by subtracting qty from stok_akhir.
//       e. Insert history_stok capturing before/after with jenis_transaksi="penjualan".
//  5. If any error occurs -> ROLLBACK.
//  6. If all succeed -> COMMIT.
// On success JualHeader.ID is populated; Details keep their inserted IDs.
func (r *PenjualanRepo) CreatePenjualanTx(ctx context.Context, hdr *models.JualHeader) error {
    if hdr == nil { return errors.New("header is nil") }
    if len(hdr.Details) == 0 { return errors.New("details empty") }

    // Begin transaction
    tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{})
    if err != nil { return fmt.Errorf("begin tx: %w", err) }

    rollback := func(e error) error {
        _ = tx.Rollback()
        return e
    }

    // 1. Validate barang exists for every detail
    for i, d := range hdr.Details {
        var exists bool
        if err := tx.QueryRowContext(ctx, "SELECT 1 FROM master_barang WHERE id=$1", d.BarangID).Scan(&exists); err != nil {
            if err == sql.ErrNoRows {
                return rollback(fmt.Errorf("barang id %d not found (detail index %d)", d.BarangID, i))
            }
            return rollback(fmt.Errorf("validate barang: %w", err))
        }
    }

    // 2. Compute subtotal & total
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

    // 3. Insert jual_header
    if err := tx.QueryRowContext(ctx, `INSERT INTO jual_header (no_faktur, customer, total, user_id, status)
            VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
        hdr.NoFaktur, hdr.Customer, hdr.Total, hdr.UserID, hdr.Status,
    ).Scan(&hdr.ID, &hdr.CreatedAt); err != nil {
        return rollback(fmt.Errorf("insert header: %w", err))
    }

    // 4. Process each detail with stock validation
    for i := range hdr.Details {
        d := &hdr.Details[i]

        // 4a. Lock current stock
        var stokBefore sql.NullInt64
        if err := tx.QueryRowContext(ctx, "SELECT stok_akhir FROM mstok WHERE barang_id=$1 FOR UPDATE", d.BarangID).Scan(&stokBefore); err != nil && err != sql.ErrNoRows {
            return rollback(fmt.Errorf("lock stock: %w", err))
        }

        // Treat missing stock row as zero available
        var before int64
        if stokBefore.Valid { before = stokBefore.Int64 } else { before = 0 }

        // 4b. Validate availability
        if before < d.Qty {
            return rollback(fmt.Errorf("insufficient stock for barang %d: have %d, need %d (detail index %d)", d.BarangID, before, d.Qty, i))
        }
        after := before - d.Qty

        // 4c. Insert jual_detail
        if err := tx.QueryRowContext(ctx, `INSERT INTO jual_detail (jual_header_id, barang_id, qty, harga, subtotal)
                VALUES ($1,$2,$3,$4,$5) RETURNING id`,
            hdr.ID, d.BarangID, d.Qty, d.Harga, d.Subtotal,
        ).Scan(&d.ID); err != nil {
            return rollback(fmt.Errorf("insert detail: %w", err))
        }

        // 4d. Update mstok (must exist because before >= qty implies existing row or zero)
        // If stok row didn't exist (before=0), this path would have returned earlier due to insufficient stock.
        res, uErr := tx.ExecContext(ctx, "UPDATE mstok SET stok_akhir=$1 WHERE barang_id=$2", after, d.BarangID)
        if uErr != nil { return rollback(fmt.Errorf("update mstok: %w", uErr)) }
        if rows, _ := res.RowsAffected(); rows == 0 {
            // Should not happen because we validated before from existing row; defensive check
            return rollback(errors.New("expected mstok update to affect 1 row"))
        }

        // 4e. Insert history_stok
        if _, hErr := tx.ExecContext(ctx, `INSERT INTO history_stok (barang_id, user_id, jenis_transaksi, jumlah, stok_sebelum, stok_sesudah)
                VALUES ($1,$2,$3,$4,$5,$6)`,
            d.BarangID, hdr.UserID, "penjualan", d.Qty, before, after,
        ); hErr != nil {
            return rollback(fmt.Errorf("insert history: %w", hErr))
        }
    }

    // Commit
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit tx: %w", err)
    }
    if hdr.CreatedAt.IsZero() { hdr.CreatedAt = time.Now() }
    return nil
}
