package repositories

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "time"

    "warehouse/models"
)

type PenjualanRepo struct {
    DB *sql.DB
}

func NewPenjualanRepo(db *sql.DB) *PenjualanRepo { return &PenjualanRepo{DB: db} }

func (r *PenjualanRepo) CreatePenjualanTx(ctx context.Context, hdr *models.JualHeader) error {
    if hdr == nil { return errors.New("header is nil") }
    if len(hdr.Details) == 0 { return errors.New("details empty") }

    tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{})
    if err != nil { return fmt.Errorf("begin tx: %w", err) }

    rollback := func(e error) error {
        _ = tx.Rollback()
        return e
    }

    for i := range hdr.Details {
        d := &hdr.Details[i]
        var exists bool
        if err := tx.QueryRowContext(ctx, "SELECT 1 FROM master_barang WHERE id=$1", d.BarangID).Scan(&exists); err != nil {
            if err == sql.ErrNoRows {
                return rollback(fmt.Errorf("barang id %d not found (detail index %d)", d.BarangID, i))
            }
            return rollback(fmt.Errorf("validate barang: %w", err))
        }
    }

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

    if err := tx.QueryRowContext(ctx, `INSERT INTO jual_header (no_faktur, customer, total, user_id, status)
            VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
        hdr.NoFaktur, hdr.Customer, hdr.Total, hdr.UserID, hdr.Status,
    ).Scan(&hdr.ID, &hdr.CreatedAt); err != nil {
        return rollback(fmt.Errorf("insert header: %w", err))
    }

    for i := range hdr.Details {
        d := &hdr.Details[i]
        var stokBefore sql.NullInt64
        if err := tx.QueryRowContext(ctx, "SELECT stok_akhir FROM mstok WHERE barang_id=$1 FOR UPDATE", d.BarangID).Scan(&stokBefore); err != nil && err != sql.ErrNoRows {
            return rollback(fmt.Errorf("lock stock: %w", err))
        }
        var before int64
        if stokBefore.Valid { before = stokBefore.Int64 } else { before = 0 }
        if before < d.Qty {
            return rollback(fmt.Errorf("insufficient stock for barang %d: have %d, need %d (detail index %d)", d.BarangID, before, d.Qty, i))
        }
        after := before - d.Qty

        if err := tx.QueryRowContext(ctx, `INSERT INTO jual_detail (jual_header_id, barang_id, qty, harga, subtotal)
                VALUES ($1,$2,$3,$4,$5) RETURNING id`,
            hdr.ID, d.BarangID, d.Qty, d.Harga, d.Subtotal,
        ).Scan(&d.ID); err != nil {
            return rollback(fmt.Errorf("insert detail: %w", err))
        }

        res, uErr := tx.ExecContext(ctx, "UPDATE mstok SET stok_akhir=$1 WHERE barang_id=$2", after, d.BarangID)
        if uErr != nil { return rollback(fmt.Errorf("update mstok: %w", uErr)) }
        if rows, _ := res.RowsAffected(); rows == 0 {
            return rollback(errors.New("expected mstok update to affect 1 row"))
        }

        if _, hErr := tx.ExecContext(ctx, `INSERT INTO history_stok (barang_id, user_id, jenis_transaksi, jumlah, stok_sebelum, stok_sesudah)
                VALUES ($1,$2,$3,$4,$5,$6)`,
            d.BarangID, hdr.UserID, "penjualan", d.Qty, before, after,
        ); hErr != nil {
            return rollback(fmt.Errorf("insert history: %w", hErr))
        }
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit tx: %w", err)
    }
    if hdr.CreatedAt.IsZero() { hdr.CreatedAt = time.Now() }
    return nil
}

func (r *PenjualanRepo) GetAll(ctx context.Context) ([]models.JualHeader, error) {
    const q = `SELECT id, no_faktur, customer, total, user_id, status, created_at
               FROM jual_header
               ORDER BY created_at DESC`
    rows, err := r.DB.QueryContext(ctx, q)
    if err != nil { return nil, fmt.Errorf("query headers: %w", err) }
    defer rows.Close()

    list := make([]models.JualHeader, 0)
    for rows.Next() {
        var h models.JualHeader
        if err := rows.Scan(&h.ID, &h.NoFaktur, &h.Customer, &h.Total, &h.UserID, &h.Status, &h.CreatedAt); err != nil {
            return nil, fmt.Errorf("scan header: %w", err)
        }
        list = append(list, h)
    }
    if err := rows.Err(); err != nil { return nil, fmt.Errorf("rows err: %w", err) }
    return list, nil
}

func (r *PenjualanRepo) GetByID(ctx context.Context, id int64) (*models.JualHeader, error) {
    const qHeader = `SELECT h.id, h.no_faktur, h.customer, h.total, h.user_id, h.status, h.created_at,
                            u.id, u.username, u.password, u.email, u.full_name, u.role
                     FROM jual_header h
                     JOIN users u ON u.id = h.user_id
                     WHERE h.id = $1`
    var h models.JualHeader
    var u models.User
    if err := r.DB.QueryRowContext(ctx, qHeader, id).Scan(
        &h.ID, &h.NoFaktur, &h.Customer, &h.Total, &h.UserID, &h.Status, &h.CreatedAt,
        &u.ID, &u.Username, &u.Password, &u.Email, &u.FullName, &u.Role,
    ); err != nil {
        if err == sql.ErrNoRows { return nil, nil }
        return nil, fmt.Errorf("get header: %w", err)
    }
    h.UserDetail = &u

    const qDetail = `SELECT d.id, d.jual_header_id, d.barang_id, d.qty, d.harga, d.subtotal,
                            b.id, b.kode_barang, b.nama_barang, b.deskripsi, b.satuan, b.harga_beli, b.harga_jual
                     FROM jual_detail d
                     JOIN master_barang b ON b.id = d.barang_id
                     WHERE d.jual_header_id = $1 ORDER BY d.id ASC`
    rows, err := r.DB.QueryContext(ctx, qDetail, id)
    if err != nil { return nil, fmt.Errorf("query details: %w", err) }
    defer rows.Close()
    details := make([]models.JualDetail, 0)
    for rows.Next() {
        var d models.JualDetail
        var b models.Barang
        var desc sql.NullString
        if err := rows.Scan(
            &d.ID, &d.JualHeaderID, &d.BarangID, &d.Qty, &d.Harga, &d.Subtotal,
            &b.ID, &b.KodeBarang, &b.NamaBarang, &desc, &b.Satuan, &b.HargaBeli, &b.HargaJual,
        ); err != nil {
            return nil, fmt.Errorf("scan detail: %w", err)
        }
        if desc.Valid { v := desc.String; b.Deskripsi = &v }
        d.BarangDetail = &b
        details = append(details, d)
    }
    if err := rows.Err(); err != nil { return nil, fmt.Errorf("rows err: %w", err) }
    h.Details = details
    return &h, nil
}
