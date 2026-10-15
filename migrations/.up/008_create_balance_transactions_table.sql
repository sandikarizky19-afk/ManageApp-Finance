CREATE TABLE IF NOT EXISTS balance_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    transaction_type VARCHAR(20) NOT NULL
        CHECK (transaction_type IN ('income','expense','transfer')),
    amount NUMERIC(20,2) NOT NULL CHECK (amount > 0),
    from_account_id BIGINT REFERENCES account_balance(id),
    to_account_id BIGINT REFERENCES account_balance(id),
    category_id BIGINT REFERENCES transaction_categories(id),
    reference_id VARCHAR(100),
    description TEXT,
    from_balance_after NUMERIC(20,2),
    to_balance_after NUMERIC(20,2),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        from_account_id IS NULL
        OR to_account_id IS NULL
        OR from_account_id <> to_account_id
    ),
    CHECK (
        (transaction_type = 'income'
            AND from_account_id IS NULL AND to_account_id IS NOT NULL)
        OR
        (transaction_type = 'expense'
            AND from_account_id IS NOT NULL AND to_account_id IS NULL)
        OR
        (transaction_type = 'transfer'
            AND from_account_id IS NOT NULL AND to_account_id IS NOT NULL)
    ),
    UNIQUE (user_id, reference_id)
);

CREATE INDEX idx_bt_user_date
    ON balance_transactions(user_id, occurred_at DESC);
CREATE INDEX idx_bt_from_account_date
    ON balance_transactions(from_account_id, occurred_at DESC);
CREATE INDEX idx_bt_to_account_date
    ON balance_transactions(to_account_id, occurred_at DESC);