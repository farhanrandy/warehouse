package repositories

import (
	"context"
	"database/sql"
	"warehouse/models"
)

type StokRepo struct {
    DB *sql.DB
}

func NewStokRepo(db *sql.DB) *StokRepo { return &StokRepo{DB: db} }

func (r *StokRepo) GetStokAkhirAll(ctx context.Context) ([]models.Mstok, error) {
    const q = `SELECT s.id, s.barang_id, s.stok_akhir,
        b.id, b.kode_barang, b.nama_barang, b.deskripsi, b.satuan, b.harga_beli, b.harga_jual
        FROM mstok s
        JOIN master_barang b ON b.id = s.barang_id
        ORDER BY b.nama_barang ASC`

    rows, err := r.DB.QueryContext(ctx, q)
    if err != nil { return nil, err }
    defer rows.Close()

    list := []models.Mstok{}
    for rows.Next() {
        var m models.Mstok
        var b models.Barang
        var desc sql.NullString
        if err := rows.Scan(&m.ID, &m.BarangID, &m.StokAkhir,
            &b.ID, &b.KodeBarang, &b.NamaBarang, &desc, &b.Satuan, &b.HargaBeli, &b.HargaJual); err != nil {
            return nil, err
        }
        if desc.Valid { v := desc.String; b.Deskripsi = &v }
        m.Barang = &b
        list = append(list, m)
    }
    if err := rows.Err(); err != nil { return nil, err }
    return list, nil
}

func (r *StokRepo) GetHistoryAll(ctx context.Context, page, limit int) ([]models.HistoryStok, int, error) {
    if page < 1 { page = 1 }
    if limit < 1 { limit = 10 }
    offset := (page - 1) * limit

    const countQ = `SELECT COUNT(*) FROM history_stok`
    var total int
    if err := r.DB.QueryRowContext(ctx, countQ).Scan(&total); err != nil { return nil, 0, err }

    const q = `SELECT h.id, h.barang_id, h.user_id, h.jenis_transaksi, h.jumlah, h.stok_sebelum, h.stok_sesudah, h.created_at,
        b.id, b.kode_barang, b.nama_barang, b.deskripsi, b.satuan, b.harga_beli, b.harga_jual,
        u.id, u.username, u.password, u.email, u.full_name, u.role
        FROM history_stok h
        JOIN master_barang b ON b.id = h.barang_id
        JOIN users u ON u.id = h.user_id
        ORDER BY h.created_at DESC
        LIMIT $1 OFFSET $2`

    rows, err := r.DB.QueryContext(ctx, q, limit, offset)
    if err != nil { return nil, 0, err }
    defer rows.Close()

    list := []models.HistoryStok{}
    for rows.Next() {
        var hs models.HistoryStok
        var b models.Barang
        var u models.User
        var desc sql.NullString
        if err := rows.Scan(&hs.ID, &hs.BarangID, &hs.UserID, &hs.JenisTransaksi, &hs.Jumlah, &hs.StokSebelum, &hs.StokSesudah, &hs.CreatedAt,
            &b.ID, &b.KodeBarang, &b.NamaBarang, &desc, &b.Satuan, &b.HargaBeli, &b.HargaJual,
            &u.ID, &u.Username, &u.Password, &u.Email, &u.FullName, &u.Role); err != nil {
            return nil, 0, err
        }
        if desc.Valid { v := desc.String; b.Deskripsi = &v }
        hs.BarangDetail = &b
        hs.UserDetail = &u
        list = append(list, hs)
    }
    if err := rows.Err(); err != nil { return nil, 0, err }
    return list, total, nil
}

func (r *StokRepo) GetStokByBarangID(ctx context.Context, barangID int64) (*models.Mstok, error) {
    const q = `SELECT s.id, s.barang_id, s.stok_akhir,
        b.id, b.kode_barang, b.nama_barang, b.deskripsi, b.satuan, b.harga_beli, b.harga_jual
        FROM mstok s
        JOIN master_barang b ON b.id = s.barang_id
        WHERE s.barang_id = $1`
    var m models.Mstok
    var b models.Barang
    var desc sql.NullString
    err := r.DB.QueryRowContext(ctx, q, barangID).Scan(&m.ID, &m.BarangID, &m.StokAkhir,
        &b.ID, &b.KodeBarang, &b.NamaBarang, &desc, &b.Satuan, &b.HargaBeli, &b.HargaJual)
    if err != nil {
        if err == sql.ErrNoRows { return nil, nil }
        return nil, err
    }
    if desc.Valid { v := desc.String; b.Deskripsi = &v }
    m.Barang = &b
    return &m, nil
}

func (r *StokRepo) GetHistoryByBarangID(ctx context.Context, barangID int64, page, limit int) ([]models.HistoryStok, int, error) {
    if page < 1 { page = 1 }
    if limit < 1 { limit = 10 }
    offset := (page - 1) * limit

    const countQ = `SELECT COUNT(*) FROM history_stok WHERE barang_id = $1`
    var total int
    if err := r.DB.QueryRowContext(ctx, countQ, barangID).Scan(&total); err != nil { return nil, 0, err }

    const q = `SELECT h.id, h.barang_id, h.user_id, h.jenis_transaksi, h.jumlah, h.stok_sebelum, h.stok_sesudah, h.created_at,
        b.id, b.kode_barang, b.nama_barang, b.deskripsi, b.satuan, b.harga_beli, b.harga_jual,
        u.id, u.username, u.password, u.email, u.full_name, u.role
        FROM history_stok h
        JOIN master_barang b ON b.id = h.barang_id
        JOIN users u ON u.id = h.user_id
        WHERE h.barang_id = $1
        ORDER BY h.created_at DESC
        LIMIT $2 OFFSET $3`
    rows, err := r.DB.QueryContext(ctx, q, barangID, limit, offset)
    if err != nil { return nil, 0, err }
    defer rows.Close()
    list := []models.HistoryStok{}
    for rows.Next() {
        var hs models.HistoryStok
        var b models.Barang
        var u models.User
        var desc sql.NullString
        if err := rows.Scan(&hs.ID, &hs.BarangID, &hs.UserID, &hs.JenisTransaksi, &hs.Jumlah, &hs.StokSebelum, &hs.StokSesudah, &hs.CreatedAt,
            &b.ID, &b.KodeBarang, &b.NamaBarang, &desc, &b.Satuan, &b.HargaBeli, &b.HargaJual,
            &u.ID, &u.Username, &u.Password, &u.Email, &u.FullName, &u.Role); err != nil {
            return nil, 0, err
        }
        if desc.Valid { v := desc.String; b.Deskripsi = &v }
        hs.BarangDetail = &b
        hs.UserDetail = &u
        list = append(list, hs)
    }
    if err := rows.Err(); err != nil { return nil, 0, err }
    return list, total, nil
}
