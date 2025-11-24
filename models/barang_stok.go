package models

type BarangWithStok struct {
	ID         int64   `json:"id"`
	KodeBarang string  `json:"kode_barang"`
	NamaBarang string  `json:"nama_barang"`
	Deskripsi  *string `json:"deskripsi,omitempty"`
	Satuan     string  `json:"satuan"`
	HargaBeli  int64   `json:"harga_beli"`
	HargaJual  int64   `json:"harga_jual"`
	StokAkhir  int64   `json:"stok_akhir"`
}
