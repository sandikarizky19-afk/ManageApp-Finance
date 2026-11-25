package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// AccountBalanceModel merepresentasikan satu baris di tabel account_balance.
type AccountBalanceModel struct {
	ID          int64           `json:"id"`
	UserId      int64           `json:"user_id"`
	AccountName string          `json:"account_name"`
	AccountType string          `json:"account_type"`
	Balance     decimal.Decimal `json:"balance"` // satuan terkecil (rupiah), kolom DB: BIGINT
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   *time.Time      `json:"-"` // soft delete, tidak perlu diekspos ke client
}

type CreateAccountBalanceRequest struct {
	AccountName string          `json:"account_name" validate:"required,min=1,max=100"`
	AccountType string          `json:"account_type"`
	Balance     decimal.Decimal `json:"balance" validate:"gte=0"`
}

type AddBalanceRequest struct {
	Balance     decimal.Decimal `json:"balance" validate:"required,gt=0"`
	AccountType string          `json:"account_type"`
}

type AccountBalanceResponse struct {
	ID          int64           `json:"id"`
	AccountName string          `json:"account_name"`
	AccountType string          `json:"account_type"`
	Balance     decimal.Decimal `json:"balance"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type GetBalanceRequest struct {
	ID          int64  `json:"id"`
	AccountType string `json:"account_type"`
}

type GetBalanceResponse struct {
	ID          int64           `json:"id"`
	UserId      int64           `json:"user_id"`
	AccountType string          `json:"account_type"`
	AccountName string          `json:"account_name"`
	Balance     decimal.Decimal `json:"balance"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type TransferBalanceRequest struct {
	FromAccountType string          `json:"from_account_type" validate:"required"`
	ToAccountType   string          `json:"to_account_type" validate:"required"`
	Amount          decimal.Decimal `json:"amount" validate:"required,gt=0"`
}

type TransferBalanceResponse struct {
	FromAccountType string          `json:"from_account_type"`
	ToAccountType   string          `json:"to_account_type"`
	Amount          decimal.Decimal `json:"amount"`
	// UpdatedAt       time.Time       `json:"updated_at"`
}
