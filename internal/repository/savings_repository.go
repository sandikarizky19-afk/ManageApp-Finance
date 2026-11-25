package repository

import (
	"context"
	"database/sql"
	"errors"
	"finance-service/internal/model"
	"finance-service/internal/utils"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type SavingsRepository struct {
	DB *sql.DB
}

func NewSavingsRepository(db *sql.DB) *SavingsRepository {
	return &SavingsRepository{DB: db}
}

func (r *SavingsRepository) CreateSavings(ctx context.Context, userId int, req model.CreateSavingsRequest) (*model.SavingsResponse, error) {
	var (
		id            string
		createdAt     time.Time
		currentAmount float64
		isCompleted   bool
	)

	qSavings := `
        INSERT INTO savings (user_id, title, icon, target_amount, current_amount)
        VALUES ($1, $2, $3, $4, 0)
        RETURNING id, created_at, current_amount, is_completed`
	err := r.DB.QueryRowContext(ctx, qSavings, userId, req.Title, req.Icon, req.TargetAmount).Scan(&id, &createdAt, &currentAmount, &isCompleted)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, utils.AppError{
				Code:    http.StatusConflict,
				Message: "Data tabungan sudah pernah dibuat",
			}
		}
		slog.ErrorContext(ctx, "create savings failed", "err", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal membuat data tabungan",
		}
	}

	res := &model.SavingsResponse{
		ID:            id,
		Title:         req.Title,
		Icon:          req.Icon,
		TargetAmount:  req.TargetAmount,
		CurrentAmount: currentAmount,
		IsCompleted:   isCompleted,
		CreatedAt:     createdAt.Format(time.RFC3339), // Otomatis format ke ISO-8601 (YYYY-MM-DDTHH:mm:ssZ)
	}

	return res, nil

}

func (r *SavingsRepository) GetAllSavings(ctx context.Context, userId int) ([]model.SavingsModel, error) {
	q := `SELECT
		  id, user_id, title, icon, target_amount, current_amount, is_completed, created_at
		  FROM savings
		  WHERE user_id = $1 and deleted_at IS NULL`
	rows, err := r.DB.QueryContext(ctx, q, userId)
	if err != nil {
		slog.ErrorContext(ctx, "gagal mengambil data tabungan", "err", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal mengambil data tabungan",
		}
	}
	defer rows.Close()

	result := make([]model.SavingsModel, 0)
	for rows.Next() {
		var m model.SavingsModel
		if err := rows.Scan(
			&m.Id,
			&m.UserId,
			&m.Title,
			&m.Icon,
			&m.TargetAmount,
			&m.CurrentAmount,
			&m.IsCompleted,
			&m.CreatedAt,
		); err != nil {
			slog.ErrorContext(ctx, "gagal memproses data tabungan", "err", err)
			return nil, utils.AppError{
				Code:    http.StatusInternalServerError,
				Message: "Gagal memproses data tabungan",
			}
		}
		result = append(result, m)
	}

	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "error saat iterasi data tabungan", "err", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal memproses seluruh data tabungan",
		}
	}

	return result, nil
}

func (r *SavingsRepository) GetSaving(ctx context.Context, userId int, savingId string, m *model.SavingsModel) (*model.SavingsModel, error) {
	q := `SELECT
		  id, user_id, title, icon, target_amount, current_amount, is_completed, created_at
		  FROM savings
		  WHERE user_id = $1 and id = $2 and deleted_at IS NULL`
	err := r.DB.QueryRowContext(ctx, q, userId, savingId).Scan(
		&m.Id,
		&m.UserId,
		&m.Title,
		&m.Icon,
		&m.TargetAmount,
		&m.CurrentAmount,
		&m.IsCompleted,
		&m.CreatedAt,
	)
	var response model.SavingsModel
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.AppError{
				Code:    http.StatusNotFound,
				Message: "Data tabungan tidak ditemukan",
			}
		}
		log.Printf("[ERROR ] Gagal mengambil data tabungan (UserID: %d, SavingID: %s): %v", userId, savingId, err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal mengambil data tabungan",
		}
	}

	return &response, nil

}

// func (r *SavingsRepository) TopUpSavings(ctx context.Context, userId int, req *model.TopUpSavingsRequest) (*model.SavingsResponse, error) {
// 	q := ``
// }
