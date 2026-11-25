package controller

import (
	"context"
	"database/sql"
	"finance-service/internal/model"
	"finance-service/internal/repository"
	"finance-service/internal/utils"
	"net/http"
	"strings"
	"time"
)

type AccountBalanceController struct {
	DB   *sql.DB
	JWT  *utils.JWTManager
	Repo *repository.AccountBalanceRepository
}

func NewAccountBalanceController(db *sql.DB, jwt *utils.JWTManager) *AccountBalanceController {
	return &AccountBalanceController{
		DB:   db,
		JWT:  jwt,
		Repo: repository.NewAccountBalanceRepository(db),
	}
}

func (c *AccountBalanceController) CreateAccountBalance(ctx context.Context, req model.CreateAccountBalanceRequest, userId int64) (*model.AccountBalanceModel, error) {
	name := strings.TrimSpace(req.AccountName)
	if name == "" {
		return nil, utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "account_name tidak boleh kosong",
		}
	}

	// UserId selalu berasal dari token, bukan dari body request.
	m := model.AccountBalanceModel{
		UserId:      userId,
		AccountName: name,
		AccountType: req.AccountType,
		Balance:     req.Balance,
	}

	return c.Repo.CreateAccountBalance(ctx, m)
}

func (c *AccountBalanceController) AddBalance(ctx context.Context, userId int, req model.AddBalanceRequest) (*model.AccountBalanceModel, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.Repo.AddBalance(ctx, userId, req)
}

func (c *AccountBalanceController) GetBalance(ctx context.Context, userId int, req model.GetBalanceRequest) (*model.GetBalanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.Repo.GetBalance(ctx, userId, req)
}

func (c *AccountBalanceController) TransferBalance(ctx context.Context, userId int, req model.TransferBalanceRequest) (*model.TransferBalanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.Repo.TransferBalance(ctx, userId, req)
}
