package models

import "github.com/shopspring/decimal"

type Carpool struct {
	ID             int64           `json:"id"`
	DeparturePoint string          `json:"departure_point"`
	Destination    string          `json:"destination"`
	AvailableSeats int64           `json:"available_seats"`
	PricePerSeat   decimal.Decimal `json:"price_per_seat"`
	PaymentMethod  string          `json:"payment_method"`
}
