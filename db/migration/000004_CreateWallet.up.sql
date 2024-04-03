CREATE TABLE IF NOT EXISTS wallets(
    id bigserial primary key,
    user_id bigint NOT NULL,
    balance numeric NOT NULL DEFAULT 0.0,
    created_at timestamptz NOT NULL DEFAULT (now()) NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT (now()) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
)