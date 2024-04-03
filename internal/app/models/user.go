package models

type User struct {
	ID            int64  `json:"id"`
	EmailAddress  string `json:"email_address"`
	PhoneNumber   string `json:"phone_number"`
	UserName      string `json:"user_name"`
	WalletAddress string `json:"wallet_address"`
}
