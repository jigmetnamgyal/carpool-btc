CREATE TABLE IF NOT EXISTS carpools(
    id bigserial PRIMARY KEY,
    departure_point text NOT NULL,
    destination text NOT NULL,
    departure_time timestamptz NOT NULL,
    available_seats INT NOT NULL,
    price_per_seat numeric NOT NULL,
    payment_method TEXT NOT NULL DEFAULT 'online'
)