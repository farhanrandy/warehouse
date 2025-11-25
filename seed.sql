-- 1. WIPING DATA (Bersihkan semua data & reset ID)
-- CATATAN: TRUNCATE CASCADE memastikan semua FK terhapus juga.

TRUNCATE TABLE history_stok RESTART IDENTITY CASCADE;
TRUNCATE TABLE mstok RESTART IDENTITY CASCADE;
TRUNCATE TABLE beli_detail RESTART IDENTITY CASCADE;
TRUNCATE TABLE beli_header RESTART IDENTITY CASCADE;
TRUNCATE TABLE jual_detail RESTART IDENTITY CASCADE;
TRUNCATE TABLE jual_header RESTART IDENTITY CASCADE;
TRUNCATE TABLE master_barang RESTART IDENTITY CASCADE;
TRUNCATE TABLE users RESTART IDENTITY CASCADE;


-- 2. INSERT DUMMY DATA (Data diambil dari contoh di dokumen)

-- Users (1 Admin, 1 Regular)
INSERT INTO users (username, password, email, full_name, role) VALUES
('admin', '$2a$10$wE9s/m9mFjO/gT0.fX4gNe.f3gHjK.cO.S6mGjJqT9g', 'admin@example.com', 'Administrator', 'admin'),
('user1', '$2a$10$wE9s/m9mFjO/gT0.fX4gNe.f3gHjK.cO.S6mGjJqT9g', 'user1@example.com', 'Pegawai Gudang', 'user');

-- Master Barang (Item A, B, C, D)
INSERT INTO master_barang (kode_barang, nama_barang, deskripsi, satuan, harga_beli, harga_jual) VALUES
('BRG-001', 'Laptop Gaming X5', 'Laptop 15 inci, 16GB RAM', 'unit', 15000000, 17500000),
('BRG-002', 'Mouse Wireless Silent', 'Mouse logitech M100', 'pcs', 250000, 300000),
('BRG-003', 'Monitor 24 inch LED', 'Monitor 144Hz', 'unit', 1800000, 2200000),
('BRG-004', 'Keyboard Mechanical', 'Outemu Red Switch', 'pcs', 750000, 900000),
('BRG-005', 'Kabel HDMI 2 Meter', 'Kabel data video', 'pcs', 45000, 60000);

-- MStok (Stok Awal)
INSERT INTO mstok (barang_id, stok_akhir) VALUES
(1, 0), -- Laptop
(2, 0), -- Mouse
(3, 0), -- Monitor
(4, 0), -- Keyboard
(5, 0); -- Kabel

-- 3. TRANSAKSI (Memastikan Histori terisi)

-- Insert Pembelian Header & Detail (User ID 1 = admin)
INSERT INTO beli_header (no_faktur, supplier, total, user_id, status) VALUES
('BELI-001', 'PT Supplier Elektronik', 32500000, 1, 'selesai');
INSERT INTO beli_detail (beli_header_id, barang_id, qty, harga, subtotal) VALUES
(1, 1, 10, 1500000, 15000000), -- Beli 10 Laptop (harga harusnya 15 juta, tapi di contoh 1.5 juta)
(1, 2, 50, 250000, 12500000); -- Beli 50 Mouse

-- Insert Penjualan Header & Detail (User ID 2 = user1)
INSERT INTO jual_header (no_faktur, customer, total, user_id, status) VALUES
('JUAL-001', 'Budi Santoso', 18700000, 2, 'selesai');
INSERT INTO jual_detail (jual_header_id, barang_id, qty, harga, subtotal) VALUES
(1, 1, 1, 17500000, 17500000), -- Jual 1 Laptop
(1, 2, 4, 300000, 1200000); -- Jual 4 Mouse

-- CATATAN PENTING:
-- Jika aplikasi Go Anda sudah berjalan (go run main.go),
-- INSERT/UPDATE stok (mstok dan history_stok) harus dilakukan lewat endpoint POST /api/pembelian dan POST /api/penjualan
-- agar terisi otomatis oleh logic Go Anda. Data di atas hanya untuk memastikan header/detail terisi.