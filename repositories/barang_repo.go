package repositories

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/lib/pq"
    "warehouse/models"
)

type BarangRepo struct {
    DB *sql.DB
}

func NewBarangRepo(db *sql.DB) *BarangRepo { return &BarangRepo{DB: db} }

// ErrBarangInUse is returned when a barang cannot be deleted due to FK references
var ErrBarangInUse = errors.New("barang in use")

func (r *BarangRepo) GetAll(ctx context.Context, search string, page, limit int) ([]models.Barang, int, error) {
    if page < 1 { page = 1 }
    if limit < 1 { limit = 10 }
    offset := (page - 1) * limit

    var (
        countQ string
        listQ  string
        rows   *sql.Rows
        err    error
        total  int
    )

    if search != "" {
        countQ = `SELECT COUNT(*) FROM master_barang
                  WHERE nama_barang ILIKE $1 OR kode_barang ILIKE $1`
        listQ = `SELECT id, kode_barang, nama_barang, deskripsi, satuan, harga_beli, harga_jual
                 FROM master_barang
                 WHERE nama_barang ILIKE $1 OR kode_barang ILIKE $1
                 ORDER BY id DESC
                 LIMIT $2 OFFSET $3`
        pattern := "%" + search + "%"
        if err = r.DB.QueryRowContext(ctx, countQ, pattern).Scan(&total); err != nil {
            return nil, 0, err
        }
        rows, err = r.DB.QueryContext(ctx, listQ, pattern, limit, offset)
        if err != nil { return nil, 0, err }
    } else {
        countQ = `SELECT COUNT(*) FROM master_barang`
        listQ = `SELECT id, kode_barang, nama_barang, deskripsi, satuan, harga_beli, harga_jual
                 FROM master_barang
                 ORDER BY id DESC
                 LIMIT $1 OFFSET $2`

        if err = r.DB.QueryRowContext(ctx, countQ).Scan(&total); err != nil {
            return nil, 0, err
        }
        rows, err = r.DB.QueryContext(ctx, listQ, limit, offset)
        if err != nil { return nil, 0, err }
    }
    defer rows.Close()

    items := make([]models.Barang, 0)
    for rows.Next() {
        var b models.Barang
        var ds sql.NullString
        if err := rows.Scan(&b.ID, &b.KodeBarang, &b.NamaBarang, &ds, &b.Satuan, &b.HargaBeli, &b.HargaJual); err != nil {
            return nil, 0, err
        }
        if ds.Valid { v := ds.String; b.Deskripsi = &v }
        items = append(items, b)
    }
    if err := rows.Err(); err != nil { return nil, 0, err }
    return items, total, nil
}

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

func (r *BarangRepo) Update(ctx context.Context, b *models.Barang) error {
    const q = `
        UPDATE master_barang
        SET nama_barang=$1, deskripsi=$2, satuan=$3, harga_beli=$4, harga_jual=$5
        WHERE id=$6`

    var ds interface{}
    if b.Deskripsi == nil { ds = nil } else { ds = *b.Deskripsi }

    res, err := r.DB.ExecContext(ctx, q,
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

func (r *BarangRepo) Delete(ctx context.Context, id int64) error {
    const q = `DELETE FROM master_barang WHERE id=$1`
    res, err := r.DB.ExecContext(ctx, q, id)
    if err != nil {
        if pqErr, ok := err.(*pq.Error); ok {
            // 23503 = foreign_key_violation
            if string(pqErr.Code) == "23503" {
                return ErrBarangInUse
            }
        }
        return err
    }
    n, _ := res.RowsAffected()
    if n == 0 { return sql.ErrNoRows }
    return nil
}

// GenerateKodeBarang generates a new kode_barang with format BRG-0001, BRG-0002, ...
// It finds the highest numeric suffix among existing codes with prefix BRG- and increments it.
func (r *BarangRepo) GenerateKodeBarang(ctx context.Context) (string, error) {
    const q = `
        SELECT COALESCE(MAX(CAST(SUBSTRING(kode_barang FROM '[0-9]+') AS INTEGER)), 0)
        FROM master_barang
        WHERE kode_barang LIKE 'BRG-%'`

    var maxNum int
    if err := r.DB.QueryRowContext(ctx, q).Scan(&maxNum); err != nil {
        return "", err
    }
    next := maxNum + 1
    return fmt.Sprintf("BRG-%04d", next), nil
}

func (r *BarangRepo) GetAllWithStok(ctx context.Context) ([]models.BarangWithStok, error) {
    const q = `SELECT b.id, b.kode_barang, b.nama_barang, b.deskripsi, b.satuan, b.harga_beli, b.harga_jual,
        COALESCE(s.stok_akhir,0) AS stok_akhir
        FROM master_barang b
        LEFT JOIN mstok s ON s.barang_id = b.id
        ORDER BY b.id DESC`
    rows, err := r.DB.QueryContext(ctx, q)
    if err != nil { return nil, err }
    defer rows.Close()
    list := make([]models.BarangWithStok, 0)
    for rows.Next() {
        var ds sql.NullString
        var item models.BarangWithStok
        if err := rows.Scan(&item.ID, &item.KodeBarang, &item.NamaBarang, &ds, &item.Satuan, &item.HargaBeli, &item.HargaJual, &item.StokAkhir); err != nil {
            return nil, err
        }
        if ds.Valid { v := ds.String; item.Deskripsi = &v }
        list = append(list, item)
    }
    if err := rows.Err(); err != nil { return nil, err }
    return list, nil
}

func (r *BarangRepo) GetWithStokByID(ctx context.Context, id int64) (*models.BarangWithStok, error) {
    const q = `SELECT b.id, b.kode_barang, b.nama_barang, b.deskripsi, b.satuan, b.harga_beli, b.harga_jual,
        COALESCE(s.stok_akhir,0) AS stok_akhir
        FROM master_barang b
        LEFT JOIN mstok s ON s.barang_id = b.id
        WHERE b.id = $1`
    var ds sql.NullString
    var item models.BarangWithStok
    err := r.DB.QueryRowContext(ctx, q, id).Scan(&item.ID, &item.KodeBarang, &item.NamaBarang, &ds, &item.Satuan, &item.HargaBeli, &item.HargaJual, &item.StokAkhir)
    if err != nil {
        if err == sql.ErrNoRows { return nil, nil }
        return nil, err
    }
    if ds.Valid { v := ds.String; item.Deskripsi = &v }
    return &item, nil
}
