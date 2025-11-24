package models

// Barang represents the master_barang table.
type Barang struct {
    ID         int64   `json:"id" db:"id"`
    KodeBarang string  `json:"kode_barang" db:"kode_barang"`
    NamaBarang string  `json:"nama_barang" db:"nama_barang"`
    Deskripsi  *string `json:"deskripsi,omitempty" db:"deskripsi"`
    Satuan     string  `json:"satuan" db:"satuan"`
    HargaBeli  int64   `json:"harga_beli" db:"harga_beli"`
    HargaJual  int64   `json:"harga_jual" db:"harga_jual"`
}
