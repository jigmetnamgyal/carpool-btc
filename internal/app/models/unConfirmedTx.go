package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type UnConfirmedTx struct {
	ID                  int64           `json:"id"`
	UserId              int64           `json:"user_id"`
	UnConfirmedBalances decimal.Decimal `json:"un_confirmed_balances"`
	CreatedAt           *time.Time      `json:"created_at"`
	UpdatedAt           *time.Time      `json:"updated_at"`
	TxID                string          `json:"tx_id"`
	Confirmed           bool            `json:"confirmed"`
	Type                string          `json:"type"`
}
