package models

// Mstok represents a row in the mstok table (current stock per barang).
// The Barang field is included for JOIN operations when you want to return
// stock together with its related master barang details. It is omitted
// from database scans (db:"-") because it is populated manually after a JOIN query.
type Mstok struct {
	ID        int64   `json:"id" db:"id"`
	BarangID  int64   `json:"barang_id" db:"barang_id"`
	StokAkhir int64   `json:"stok_akhir" db:"stok_akhir"`
	Barang    *Barang `json:"barang,omitempty" db:"-"`
}

