package model

import "time"

type SavingsModel struct {
	Id            string     `json:"id"`
	UserId        int64      `json:"user_id"`
	Title         string     `json:"title"`
	Icon          string     `json:"icon"`
	TargetAmount  float64    `json:"target_amount"`
	CurrentAmount float64    `json:"current_amount"`
	IsCompleted   bool       `json:"is_completed"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}

type CreateSavingsRequest struct {
	Title        string  `json:"title" validate:"required,max=100"`
	Icon         string  `json:"icon"`
	TargetAmount float64 `json:"target_amount" validate:"required,gt=0"`
}

type SavingsResponse struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Icon          string  `json:"icon"`
	TargetAmount  float64 `json:"target_amount"`
	CurrentAmount float64 `json:"current_amount"`
	IsCompleted   bool    `json:"is_completed"`
	CreatedAt     string  `json:"created_at"`
}

type TopUpSavingsRequest struct {
	Id     string  `json:"id" validate:"required"`
	Amount float64 `json:"current_amount" validate:"required,gt=0"`
}
