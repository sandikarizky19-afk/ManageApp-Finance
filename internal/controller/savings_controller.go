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

type SavingsController struct {
	DB   *sql.DB
	JWT  *utils.JWTManager
	Repo *repository.SavingsRepository
}

func NewSavingsController(db *sql.DB, jwt *utils.JWTManager) *SavingsController {
	return &SavingsController{
		DB:   db,
		JWT:  jwt,
		Repo: repository.NewSavingsRepository(db),
	}
}

func (c *SavingsController) CreateSavings(ctx context.Context, userId int, req model.CreateSavingsRequest) (*model.SavingsResponse, error) {
	req.Title = strings.ToLower(strings.TrimSpace(req.Title))
	if req.Title == "" || len(req.Title) > 100 {
		return nil, utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "nama tabungan wajib diisi (maks 100 karakter)",
		}
	}

	if req.TargetAmount <= 0 {
		return nil, utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Target tabungan harus lebih besar dari 0",
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	createdSavings, err := c.Repo.CreateSavings(ctx, userId, req)
	if err != nil {
		return nil, err
	}
	return createdSavings, nil
}

func (c *SavingsController) GetAllSavings(ctx context.Context, userId int64) ([]model.SavingsResponse, error) {
	savingsEntities, err := c.Repo.GetAllSavings(ctx, int(userId))
	if err != nil {
		return nil, err
	}

	responses := make([]model.SavingsResponse, 0)
	for _, s := range savingsEntities {
		responses = append(responses, model.SavingsResponse{
			ID:            s.Id,
			Title:         s.Title,
			Icon:          s.Icon,
			TargetAmount:  s.TargetAmount,
			CurrentAmount: s.CurrentAmount,
			IsCompleted:   s.IsCompleted,
		})
	}

	return responses, nil
}
