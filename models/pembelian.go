package models

import "time"

// BeliHeader represents a row in beli_header (purchase transaction header).
// Details holds the associated line items submitted in a POST request and
// can be populated after querying beli_detail for read operations.
type BeliHeader struct {
    ID        int64        `json:"id" db:"id"`
    NoFaktur  string       `json:"no_faktur" db:"no_faktur"`
    Supplier  string       `json:"supplier" db:"supplier"`
    Total     int64        `json:"total" db:"total"`
    UserID    int64        `json:"user_id" db:"user_id"`
    Status    string       `json:"status" db:"status"`
    CreatedAt time.Time    `json:"created_at" db:"created_at"`
    Details   []BeliDetail `json:"details,omitempty" db:"-"`
    UserDetail *User       `json:"user_detail,omitempty" db:"-"`
}

// BeliDetail represents a row in beli_detail (purchase line item).
type BeliDetail struct {
    ID           int64 `json:"id" db:"id"`
    BeliHeaderID int64 `json:"beli_header_id" db:"beli_header_id"`
    BarangID     int64 `json:"barang_id" db:"barang_id"`
    Qty          int64 `json:"qty" db:"qty"`
    Harga        int64 `json:"harga" db:"harga"`
    Subtotal     int64 `json:"subtotal" db:"subtotal"`
    BarangDetail *Barang `json:"barang_detail,omitempty" db:"-"`
}
