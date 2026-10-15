package repository

import (
	"context"
	"database/sql"
	"errors"
	"finance-service/internal/model"
	"finance-service/internal/utils"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type AccountTypesRepository struct {
	DB *sql.DB
}

func NewAccountTypesRepository(db *sql.DB) *AccountTypesRepository {
	return &AccountTypesRepository{DB: db}
}

func (r *AccountTypesRepository) GetAccountTypes(ctx context.Context) ([]model.AccountTypes, error) {
	q := `select
			type_code,
			type_name,
			description
		from
			public.account_types
		where
			deleted_at is null
		order by id`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal mengambil tipe saldo. " + err.Error(),
		}
	}
	defer rows.Close()

	result := make([]model.AccountTypes, 0)
	for rows.Next() {
		var m model.AccountTypes
		if err := rows.Scan(
			&m.TypeCode,
			&m.TypeName,
			&m.Description,
		); err != nil {
			return nil, utils.AppError{
				Code:    http.StatusInternalServerError,
				Message: "Gagal membaca data tipe saldo. " + err.Error(),
			}
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal mengambil tipe saldo",
		}
	}

	return result, nil
}

func (r *AccountTypesRepository) CreateAccountTypes(ctx context.Context, m model.AccountTypesModel) (*model.AccountTypesModel, error) {
	q := `INSERT INTO 
			account_types
		(type_code, type_name, description) 
			VALUES 
		($1, $2, $3)
		RETURNING id, created_at, updated_at`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, q,
		m.TypeCode,
		m.TypeName,
		m.Description,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError // github.com/jackc/pgx/v5/pgconn
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, utils.AppError{
				Code:    http.StatusConflict,
				Message: "Kode tipe saldo sudah digunakan",
			}
		}
		slog.ErrorContext(ctx, "create account type failed", "err", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal membuat tipe saldo",
		}
	}

	return &m, nil
}

func (r *AccountTypesRepository) UpdateAccountTypes(ctx context.Context, id int64, m model.UpdateAccountTypeRequest) (*model.AccountTypeResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q :=
		`UPDATE account_types 
		 SET 
			type_code = $2,
			type_name = $3,
			description = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			id = $1
		RETURNING
			id, type_code, type_name, description, created_at, updated_at
		`
	var response model.AccountTypeResponse

	err := r.DB.QueryRowContext(
		ctx,
		q,
		id,
		m.TypeCode,
		m.TypeName,
		m.Description,
	).Scan(
		&response.ID,
		&response.TypeCode,
		&response.TypeName,
		&response.Description,
		&response.CreatedAt,
		&response.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.AppError{
				Code:    http.StatusNotFound,
				Message: "Account type tidak ditemukan",
			}
		}

		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal memperbarui account type",
		}
	}

	return &response, nil

}

func (r *AccountTypesRepository) DeleteAccountTypes(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		UPDATE acoount_types
		SET
			deleted_at = CURRENT_TIMESTAMPT,
			updated_at = CURRENT_TIMESTAMPT
		WHERE
			id=$1
		AND deleted_at is null
	`
	result, err := r.DB.ExecContext(ctx, q, id)
	if err != nil {
		return utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal menonaktifkan acoount type",
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal memeriksa hasil soft delete",
		}
	}

	if rowsAffected == 0 {
		return utils.AppError{
			Code:    http.StatusNotFound,
			Message: "Account type tidak ditemukan",
		}
	}

	return nil
}
