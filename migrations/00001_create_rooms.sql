-- +goose Up
CREATE TABLE rooms (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    number        VARCHAR(20) NOT NULL UNIQUE,
    -- Harga dalam rupiah, disimpan sebagai bilangan bulat (bukan float).
    monthly_price BIGINT      NOT NULL CHECK (monthly_price > 0),
    status        VARCHAR(20) NOT NULL DEFAULT 'available'
                  CHECK (status IN ('available', 'occupied', 'maintenance')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE rooms;
