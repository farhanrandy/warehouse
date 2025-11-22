-- Warehouse Inventory Schema
-- Create tables according to REQUIREMENTS.md

-- 0) Optional: set sane defaults for this session
SET client_min_messages = WARNING;

-- 1) users
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(50)  NOT NULL UNIQUE,
    password    TEXT         NOT NULL,
    email       VARCHAR(120) NOT NULL UNIQUE,
    full_name   VARCHAR(120) NOT NULL,
    role        VARCHAR(20)  NOT NULL
);

-- 2) master_barang
CREATE TABLE IF NOT EXISTS master_barang (
    id           BIGSERIAL PRIMARY KEY,
    kode_barang  VARCHAR(50)  NOT NULL UNIQUE,
    nama_barang  VARCHAR(120) NOT NULL,
    deskripsi    TEXT,
    satuan       VARCHAR(30)  NOT NULL,
    harga_beli   INTEGER      NOT NULL CHECK (harga_beli >= 0),
    harga_jual   INTEGER      NOT NULL CHECK (harga_jual >= 0)
);

-- Helpful index for name searches
CREATE INDEX IF NOT EXISTS idx_master_barang_nama ON master_barang (nama_barang);

-- 3) mstok (current stock per barang)
CREATE TABLE IF NOT EXISTS mstok (
    id          BIGSERIAL PRIMARY KEY,
    barang_id   BIGINT      NOT NULL UNIQUE REFERENCES master_barang(id) ON DELETE CASCADE,
    stok_akhir  INTEGER     NOT NULL DEFAULT 0 CHECK (stok_akhir >= 0)
);

-- 4) beli_header (purchase header)
CREATE TABLE IF NOT EXISTS beli_header (
    id         BIGSERIAL PRIMARY KEY,
    no_faktur  VARCHAR(50)  NOT NULL UNIQUE,
    supplier   VARCHAR(120) NOT NULL,
    total      INTEGER      NOT NULL CHECK (total >= 0),
    user_id    BIGINT       NOT NULL REFERENCES users(id),
    status     VARCHAR(20)  NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_beli_header_user ON beli_header (user_id);

-- 5) beli_detail (purchase detail)
CREATE TABLE IF NOT EXISTS beli_detail (
    id              BIGSERIAL PRIMARY KEY,
    beli_header_id  BIGINT  NOT NULL REFERENCES beli_header(id) ON DELETE CASCADE,
    barang_id       BIGINT  NOT NULL REFERENCES master_barang(id),
    qty             INTEGER NOT NULL CHECK (qty > 0),
    harga           INTEGER NOT NULL CHECK (harga >= 0),
    subtotal        INTEGER NOT NULL CHECK (subtotal >= 0)
);
CREATE INDEX IF NOT EXISTS idx_beli_detail_header ON beli_detail (beli_header_id);
CREATE INDEX IF NOT EXISTS idx_beli_detail_barang ON beli_detail (barang_id);

-- 6) jual_header (sales header)
CREATE TABLE IF NOT EXISTS jual_header (
    id         BIGSERIAL PRIMARY KEY,
    no_faktur  VARCHAR(50)  NOT NULL UNIQUE,
    customer   VARCHAR(120) NOT NULL,
    total      INTEGER      NOT NULL CHECK (total >= 0),
    user_id    BIGINT       NOT NULL REFERENCES users(id),
    status     VARCHAR(20)  NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_jual_header_user ON jual_header (user_id);

-- 7) jual_detail (sales detail)
CREATE TABLE IF NOT EXISTS jual_detail (
    id              BIGSERIAL PRIMARY KEY,
    jual_header_id  BIGINT  NOT NULL REFERENCES jual_header(id) ON DELETE CASCADE,
    barang_id       BIGINT  NOT NULL REFERENCES master_barang(id),
    qty             INTEGER NOT NULL CHECK (qty > 0),
    harga           INTEGER NOT NULL CHECK (harga >= 0),
    subtotal        INTEGER NOT NULL CHECK (subtotal >= 0)
);
CREATE INDEX IF NOT EXISTS idx_jual_detail_header ON jual_detail (jual_header_id);
CREATE INDEX IF NOT EXISTS idx_jual_detail_barang ON jual_detail (barang_id);

-- 8) history_stok (stock movement history)
CREATE TABLE IF NOT EXISTS history_stok (
    id               BIGSERIAL PRIMARY KEY,
    barang_id        BIGINT       NOT NULL REFERENCES master_barang(id),
    user_id          BIGINT       NOT NULL REFERENCES users(id),
    jenis_transaksi  VARCHAR(30)  NOT NULL, -- e.g., pembelian, penjualan, penyesuaian
    jumlah           INTEGER      NOT NULL CHECK (jumlah >= 0),
    stok_sebelum     INTEGER      NOT NULL CHECK (stok_sebelum >= 0),
    stok_sesudah     INTEGER      NOT NULL CHECK (stok_sesudah >= 0),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_history_stok_barang ON history_stok (barang_id);
CREATE INDEX IF NOT EXISTS idx_history_stok_user   ON history_stok (user_id);

-- End of schema
