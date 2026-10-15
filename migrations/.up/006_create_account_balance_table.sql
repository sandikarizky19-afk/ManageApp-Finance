CREATE TABLE IF NOT EXISTS account_balance (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    account_name varchar (100),
    account_type varchar(30) NOT NULL
        REFERENCES account_types(type_code),
    balance NUMERIC(20, 2) NOT NULL DEFAULT 0.00
        CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_account_balance_user
    ON account_balance(user_id);

CREATE UNIQUE INDEX uq_user_main_account
    ON account_balance(user_id)
    WHERE account_type = 'main' AND deleted_at IS NULL;