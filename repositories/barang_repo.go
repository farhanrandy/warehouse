package repositories

import (
	"context"
	"database/sql"

	"warehouse/models"
)

// BarangRepo is a simple repository for the master_barang table.
// It uses the standard library database/sql with the lib/pq driver.
type BarangRepo struct {
    DB *sql.DB
}

// NewBarangRepo creates a new repository instance.
func NewBarangRepo(db *sql.DB) *BarangRepo { return &BarangRepo{DB: db} }

// GetAll returns all barang rows.
// Note: for real apps you will usually add pagination; here we keep it simple.
func (r *BarangRepo) GetAll(ctx context.Context) ([]models.Barang, error) {
    const q = `
        SELECT id, kode_barang, nama_barang, deskripsi, satuan, harga_beli, harga_jual
        FROM master_barang
        ORDER BY id DESC`

    rows, err := r.DB.QueryContext(ctx, q)
    if err != nil { return nil, err }
    defer rows.Close()

    items := []models.Barang{}
    for rows.Next() {
        var (
            b  models.Barang
            ds sql.NullString
        )
        if err := rows.Scan(&b.ID, &b.KodeBarang, &b.NamaBarang, &ds, &b.Satuan, &b.HargaBeli, &b.HargaJual); err != nil {
            return nil, err
        }
        if ds.Valid { v := ds.String; b.Deskripsi = &v }
        items = append(items, b)
    }
    if err := rows.Err(); err != nil { return nil, err }
    return items, nil
}

// GetByID returns a single barang by its id.
// If not found, it returns (nil, nil).
func (r *BarangRepo) GetByID(ctx context.Context, id int64) (*models.Barang, error) {
    const q = `
        SELECT id, kode_barang, nama_barang, deskripsi, satuan, harga_beli, harga_jual
        FROM master_barang WHERE id = $1`

    var (
        b  models.Barang
        ds sql.NullString
    )
    err := r.DB.QueryRowContext(ctx, q, id).
        Scan(&b.ID, &b.KodeBarang, &b.NamaBarang, &ds, &b.Satuan, &b.HargaBeli, &b.HargaJual)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil { return nil, err }
    if ds.Valid { v := ds.String; b.Deskripsi = &v }
    return &b, nil
}

// Create inserts a new barang and sets the generated ID on the struct.
func (r *BarangRepo) Create(ctx context.Context, b *models.Barang) error {
    const q = `
        INSERT INTO master_barang (kode_barang, nama_barang, deskripsi, satuan, harga_beli, harga_jual)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id`

    var ds interface{}
    if b.Deskripsi == nil { ds = nil } else { ds = *b.Deskripsi }

    return r.DB.QueryRowContext(ctx, q,
        b.KodeBarang,
        b.NamaBarang,
        ds,
        b.Satuan,
        b.HargaBeli,
        b.HargaJual,
    ).Scan(&b.ID)
}

// Update modifies an existing barang row by its ID.
// Returns sql.ErrNoRows if the ID does not exist.
func (r *BarangRepo) Update(ctx context.Context, b *models.Barang) error {
    const q = `
        UPDATE master_barang
        SET kode_barang=$1, nama_barang=$2, deskripsi=$3, satuan=$4, harga_beli=$5, harga_jual=$6
        WHERE id=$7`

    var ds interface{}
    if b.Deskripsi == nil { ds = nil } else { ds = *b.Deskripsi }

    res, err := r.DB.ExecContext(ctx, q,
        b.KodeBarang,
        b.NamaBarang,
        ds,
        b.Satuan,
        b.HargaBeli,
        b.HargaJual,
        b.ID,
    )
    if err != nil { return err }
    n, _ := res.RowsAffected()
    if n == 0 { return sql.ErrNoRows }
    return nil
}

// Delete removes a barang row by its ID.
// Returns sql.ErrNoRows if the ID does not exist.
func (r *BarangRepo) Delete(ctx context.Context, id int64) error {
    const q = `DELETE FROM master_barang WHERE id=$1`
    res, err := r.DB.ExecContext(ctx, q, id)
    if err != nil { return err }
    n, _ := res.RowsAffected()
    if n == 0 { return sql.ErrNoRows }
    return nil
}
