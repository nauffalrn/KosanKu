-- +goose Up
CREATE TABLE tenants (
    id                      UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name               VARCHAR(100) NOT NULL,
    phone                   VARCHAR(20)  NOT NULL UNIQUE,
    -- Nomor KTP opsional. Data pribadi sensitif: jangan tampilkan di daftar atau log.
    id_card_number          VARCHAR(32),
    emergency_contact_name  VARCHAR(100),
    emergency_contact_phone VARCHAR(20),
    notes                   TEXT,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE tenants;
