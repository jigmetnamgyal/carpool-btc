CREATE TABLE IF NOT EXISTS un_confirmed_balances(
    id bigserial PRIMARY KEY,
    user_id bigint REFERENCES users(id) NOT NULL,
    un_confirmed_balance numeric NOT NULL DEFAULT 0.0,
    created_at timestamptz NOT NULL DEFAULT (now()) NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT (now()) NOT NULL,
    tx_id text NOT NULL,
    type varchar NOT NULL DEFAULT 'deposit'
)