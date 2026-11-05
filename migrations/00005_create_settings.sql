-- +goose Up
-- settings = pengaturan kos yang bisa diubah pemilik.
-- Tabel singleton: kolom id hanya boleh bernilai true, jadi maksimal satu baris.
CREATE TABLE settings (
    id                           BOOLEAN     PRIMARY KEY DEFAULT true CHECK (id),
    allow_partial_payment        BOOLEAN     NOT NULL DEFAULT true,
    -- Nilai default saat membuat kontrak baru; tiap kontrak tetap bisa diubah.
    default_electricity_included BOOLEAN     NOT NULL DEFAULT false,
    default_water_included       BOOLEAN     NOT NULL DEFAULT true,
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO settings DEFAULT VALUES;

-- +goose Down
DROP TABLE settings;
