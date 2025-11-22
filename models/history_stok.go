package models

import "time"

// HistoryStok represents a row in the history_stok table, capturing each stock movement.
// BarangDetail and UserDetail are optional relation holders populated after JOIN queries.
type HistoryStok struct {
	ID             int64     `json:"id" db:"id"`
	BarangID       int64     `json:"barang_id" db:"barang_id"`
	UserID         int64     `json:"user_id" db:"user_id"`
	JenisTransaksi string    `json:"jenis_transaksi" db:"jenis_transaksi"`
	Jumlah         int64     `json:"jumlah" db:"jumlah"`
	StokSebelum    int64     `json:"stok_sebelum" db:"stok_sebelum"`
	StokSesudah    int64     `json:"stok_sesudah" db:"stok_sesudah"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	BarangDetail   *Barang   `json:"barang_detail,omitempty" db:"-"`
	UserDetail     *User     `json:"user_detail,omitempty" db:"-"`
}

