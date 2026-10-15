CREATE TABLE IF NOT EXISTS account_types (
    id BIGSERIAL PRIMARY KEY,
    type_code varchar(30) NOT NULL UNIQUE,
    type_name varchar(100) NOT NULL,
    description text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

-- INSERT INTO account_types (type_code, type_name) VALUES
--   ('main', 'Saldo Utama'),
--   ('savings', 'Tabungan'),
--   ('investment', 'Investasi'),
--   ('deposit', 'Deposito')
-- ON CONFLICT (type_code) DO NOTHING;