package models

import "time"

// JualHeader represents a row in jual_header (sales transaction header).
// Details holds associated line items submitted in a POST request and can be
// populated after querying jual_detail for read operations.
type JualHeader struct {
    ID        int64        `json:"id" db:"id"`
    NoFaktur  string       `json:"no_faktur" db:"no_faktur"`
    Customer  string       `json:"customer" db:"customer"`
    Total     int64        `json:"total" db:"total"`
    UserID    int64        `json:"user_id" db:"user_id"`
    Status    string       `json:"status" db:"status"`
    CreatedAt time.Time    `json:"created_at" db:"created_at"`
    Details   []JualDetail `json:"details,omitempty" db:"-"`
    UserDetail *User       `json:"user_detail,omitempty" db:"-"`
}

// JualDetail represents a row in jual_detail (sales line item).
type JualDetail struct {
    ID            int64 `json:"id" db:"id"`
    JualHeaderID  int64 `json:"jual_header_id" db:"jual_header_id"`
    BarangID      int64 `json:"barang_id" db:"barang_id"`
    Qty           int64 `json:"qty" db:"qty"`
    Harga         int64 `json:"harga" db:"harga"`
    Subtotal      int64 `json:"subtotal" db:"subtotal"`
    BarangDetail  *Barang `json:"barang_detail,omitempty" db:"-"`
}
