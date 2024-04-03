package models

import "github.com/shopspring/decimal"

type User struct {
	ID            int64            `json:"id"`
	EmailAddress  string           `json:"email_address"`
	PhoneNumber   string           `json:"phone_number"`
	UserName      string           `json:"user_name"`
	WalletAddress string           `json:"wallet_address"`
	RoleID        int64            `json:"role_id"`
	Balance       *decimal.Decimal `json:"balance"`
	Role          *Role            `json:"role"`
}
