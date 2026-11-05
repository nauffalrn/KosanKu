-- +goose Up
-- leases = kontrak sewa: menghubungkan satu kamar dengan satu penyewa.
CREATE TABLE leases (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id                 UUID        NOT NULL REFERENCES rooms (id),
    tenant_id               UUID        NOT NULL REFERENCES tenants (id),

    start_date              DATE        NOT NULL,
    -- NULL berarti kontrak berjalan terus sampai diakhiri.
    end_date                DATE,

    -- Siklus tagihan. Contoh:
    --   mingguan            : unit = week,  interval = 1
    --   bulanan             : unit = month, interval = 1
    --   tiap 3 bulan        : unit = month, interval = 3
    --   tahunan bayar dimuka: unit = year,  interval = 1
    billing_unit            VARCHAR(10) NOT NULL CHECK (billing_unit IN ('week', 'month', 'year')),
    billing_interval        SMALLINT    NOT NULL DEFAULT 1 CHECK (billing_interval > 0),

    -- Snapshot harga per siklus saat kontrak dibuat. Jika harga kamar berubah,
    -- kontrak yang sudah berjalan tidak ikut berubah.
    price_per_period        BIGINT      NOT NULL CHECK (price_per_period > 0),

    deposit_amount          BIGINT      NOT NULL DEFAULT 0 CHECK (deposit_amount >= 0),
    -- NULL berarti deposit belum dikembalikan/diproses.
    deposit_refunded_amount BIGINT,

    -- Snapshot dari settings saat kontrak dibuat; bisa diubah per kontrak.
    electricity_included    BOOLEAN     NOT NULL,
    water_included          BOOLEAN     NOT NULL,

    status                  VARCHAR(10) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'ended')),
    ended_at                TIMESTAMPTZ,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT leases_end_after_start CHECK (end_date IS NULL OR end_date > start_date),
    CONSTRAINT leases_ended_has_timestamp CHECK ((status = 'ended') = (ended_at IS NOT NULL)),
    CONSTRAINT leases_refund_within_deposit CHECK (
        deposit_refunded_amount IS NULL
        OR deposit_refunded_amount BETWEEN 0 AND deposit_amount
    )
);

-- Satu kamar hanya boleh punya SATU kontrak aktif. Aturan ini dijaga database,
-- jadi tetap aman walaupun ada dua request bersamaan (race condition).
CREATE UNIQUE INDEX leases_one_active_per_room ON leases (room_id) WHERE status = 'active';

-- PostgreSQL tidak otomatis membuat index untuk foreign key.
CREATE INDEX leases_room_id_idx ON leases (room_id);
CREATE INDEX leases_tenant_id_idx ON leases (tenant_id);

-- +goose Down
DROP TABLE leases;
