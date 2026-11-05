-- +goose Up
-- invoices = tagihan untuk satu periode dari satu kontrak sewa.
CREATE TABLE invoices (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    lease_id     UUID        NOT NULL REFERENCES leases (id),
    -- Periode memakai rentang setengah terbuka [period_start, period_end):
    -- period_end adalah tanggal awal periode berikutnya. Ini menghindari
    -- masalah "tanggal terakhir bulan" dan membuat periode tidak saling tumpang tindih.
    period_start DATE        NOT NULL,
    period_end   DATE        NOT NULL,
    due_date     DATE        NOT NULL,
    status       VARCHAR(16) NOT NULL DEFAULT 'unpaid'
                 CHECK (status IN ('unpaid', 'partially_paid', 'paid', 'void')),
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT invoices_valid_period CHECK (period_end > period_start),
    -- Mencegah tagihan ganda untuk periode yang sama (idempotent),
    -- penting saat nanti tagihan dibuat otomatis oleh scheduled job.
    CONSTRAINT invoices_one_per_period UNIQUE (lease_id, period_start)
);

-- Untuk query "tagihan belum lunas yang jatuh tempo" (dipakai dashboard dan MCP).
CREATE INDEX invoices_status_due_date_idx ON invoices (status, due_date);

-- invoice_items = rincian tagihan. Total tagihan = jumlah semua item.
-- Inilah "dynamic billing": isi tagihan bisa berbeda tiap periode.
CREATE TABLE invoice_items (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id  UUID         NOT NULL REFERENCES invoices (id) ON DELETE CASCADE,
    kind        VARCHAR(20)  NOT NULL
                CHECK (kind IN ('rent', 'deposit', 'electricity', 'water', 'penalty', 'other')),
    description VARCHAR(200) NOT NULL,
    amount      BIGINT       NOT NULL CHECK (amount > 0),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX invoice_items_invoice_id_idx ON invoice_items (invoice_id);

-- payments = pembayaran. Satu tagihan bisa dibayar beberapa kali (cicilan).
-- Pembayaran bersifat immutable (tidak di-update): koreksi dilakukan
-- dengan pencatatan baru, supaya jejak audit keuangan tetap utuh.
CREATE TABLE payments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID        NOT NULL REFERENCES invoices (id),
    amount     BIGINT      NOT NULL CHECK (amount > 0),
    method     VARCHAR(20) NOT NULL CHECK (method IN ('cash', 'bank_transfer', 'e_wallet', 'other')),
    paid_at    TIMESTAMPTZ NOT NULL,
    notes      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX payments_invoice_id_idx ON payments (invoice_id);
-- Untuk laporan arus kas per rentang tanggal.
CREATE INDEX payments_paid_at_idx ON payments (paid_at);

-- +goose Down
DROP TABLE payments;
DROP TABLE invoice_items;
DROP TABLE invoices;
