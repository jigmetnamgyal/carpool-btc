CREATE TABLE IF NOT EXISTS users(
    id bigserial primary key,
    email_address varchar(255) NOT NULL,
    phone_number varchar(255) NOT NULL,
    user_name varchar(255) NOT NULL,
    wallet_address text NOT NULL,
    amount numeric,
    txID text
)